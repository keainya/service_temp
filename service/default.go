package service

import "github.com/gin-gonic/gin"

func Status(c *gin.Context) {
	c.JSON(200, DefaultResponse())
}
