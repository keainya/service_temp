package router

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/keainya/service_temp/config"
	"github.com/keainya/service_temp/service"
)

func InitRouter(webFS embed.FS, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// ---- CORS 中间件 ----
	r.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ---- Session 中间件 ----
	store := cookie.NewStore([]byte("change-me-to-a-secure-random-key"))
	r.Use(sessions.Sessions("service_session", store))

	// ---- OAuth 处理器 ----
	oauth := service.NewOAuthHandler(cfg)

	// ---- 路由 ----
	r.GET("/login", oauth.Login)
	r.GET("/callback", oauth.Callback)
	r.GET("/logout", oauth.Logout)
	r.GET("/status", oauth.Status)

	// ---- 嵌入式前端静态文件 ----
	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(staticFS))

	// NoRoute：优先尝试返回静态文件，找不到再返回 JSON 404
	r.NoRoute(func(c *gin.Context) {
		// API 路径未匹配到，返回 JSON 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"code": -1, "msg": "not found"})
			return
		}
		// 尝试提供静态文件
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	return r
}
