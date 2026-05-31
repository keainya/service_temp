package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/keainya/service_temp/config"
)

// OAuthHandler 封装 OAuth 相关处理器
type OAuthHandler struct {
	Cfg *config.Config
}

// NewOAuthHandler 创建 OAuth 处理器
func NewOAuthHandler(cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{Cfg: cfg}
}

// Login 发起 OAuth 授权请求
func (h *OAuthHandler) Login(c *gin.Context) {
	session := sessions.Default(c)
	state := generateState()
	session.Set("oauth_state", state)
	_ = session.Save()

	authURL := fmt.Sprintf("%s/oauth/authorize?%s",
		h.Cfg.OAuth.AccountURL,
		url.Values{
			"response_type": {"code"},
			"client_id":     {h.Cfg.OAuth.ClientID},
			"redirect_uri":  {h.Cfg.OAuth.RedirectURI},
			"state":         {state},
		}.Encode(),
	)
	c.Redirect(http.StatusFound, authURL)
}

// Callback 处理 OAuth 回调
func (h *OAuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errParam := c.Query("error")

	if errParam != "" {
		c.JSON(http.StatusOK, Response{Code: -1, Msg: "授权失败: " + errParam})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Msg: "缺少 authorization code"})
		return
	}

	// 校验 state
	session := sessions.Default(c)
	savedState := session.Get("oauth_state")
	if savedState == nil || savedState.(string) != state {
		c.JSON(http.StatusForbidden, Response{Code: -1, Msg: "state 不匹配，可能存在 CSRF 攻击"})
		return
	}
	session.Delete("oauth_state")
	_ = session.Save()

	// 用 code 换取 access_token
	tokenResp, err := h.exchangeToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: -1, Msg: "换取 token 失败: " + err.Error()})
		return
	}

	// 获取用户信息
	userInfo, err := h.getUserInfo(tokenResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: -1, Msg: "获取用户信息失败: " + err.Error()})
		return
	}

	// 存入 session
	session.Set("user_id", userInfo.Sub)
	session.Set("username", userInfo.Username)
	session.Set("role", userInfo.Role)
	session.Set("access_token", tokenResp.AccessToken)
	_ = session.Save()

	c.Redirect(http.StatusFound, "/")
}

// 状态信息
func (h *OAuthHandler) Status(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	username := session.Get("username")
	role := session.Get("role")

	if userID == nil {
		c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: gin.H{
			"logged_in": false,
		}})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: gin.H{
		"logged_in": true,
		"user_id":   userID,
		"username":  username,
		"role":      role,
	}})
}

// 退出登录
func (h *OAuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	_ = session.Save()
	c.Redirect(http.StatusFound, "/")
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

type userInfoResponse struct {
	Sub      string `json:"sub"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func (h *OAuthHandler) exchangeToken(code string) (*tokenResponse, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {h.Cfg.OAuth.RedirectURI},
		"client_id":     {h.Cfg.OAuth.ClientID},
		"client_secret": {h.Cfg.OAuth.ClientSecret},
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(h.Cfg.OAuth.AccountURL+"/oauth/token", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Code int            `json:"code"`
		Msg  string         `json:"msg"`
		Data *tokenResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("%s", apiResp.Msg)
	}
	return apiResp.Data, nil
}

func (h *OAuthHandler) getUserInfo(accessToken string) (*userInfoResponse, error) {
	req, err := http.NewRequest("GET", h.Cfg.OAuth.AccountURL+"/oauth/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo userInfoResponse
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
