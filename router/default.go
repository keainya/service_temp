package router

import (
	"github.com/gin-gonic/gin"
	"github.com/keainya/service_temp/service"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/status", service.Status)

	return r
}
