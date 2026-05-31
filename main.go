package main

import (
	"fmt"
	"os"

	"github.com/keainya/service_temp/config"
	"github.com/keainya/service_temp/object"
	"github.com/keainya/service_temp/router"
)

func main() {
	if object.Database == nil {
		fmt.Println("database error")
		os.Exit(1)
	}

	cfg, err := config.Load("config.toml")
	if err != nil {
		fmt.Printf("load config error: %v\n", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}
	if port == "" {
		port = "8081"
	}

	router.InitRouter(webFS, cfg).Run(":" + port)
}
