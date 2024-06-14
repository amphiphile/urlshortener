package main

import (
	"flag"
	"github.com/amphiphile/urlshrtnr/internal/app"
	"github.com/amphiphile/urlshrtnr/internal/config"
	"github.com/gin-gonic/gin"
	"strings"
)

func main() {

	cfg := new(config.Config)
	cfg.DBConfig.DBFileName = "db.json"

	flag.Var(&cfg.ServerURLConfig, "a", "HTTP server startup address")
	flag.Var(&cfg.AppConfig, "b", "Base address of the shortened URL")

	flag.Parse()

	if serverAddress := cfg.ServerURLConfig.String(); serverAddress == "" {
		_ = cfg.ServerURLConfig.Set(config.GetFromEnv("SERVER_ADDRESS", "localhost:8080"))

	}
	if baseURL := cfg.AppConfig.String(); baseURL == "" {
		_ = cfg.AppConfig.Set(config.GetFromEnv("BASE_URL", "http://localhost:8080/"))
		if !strings.HasSuffix(cfg.AppConfig.String(), "/") {
			_ = cfg.AppConfig.Set(cfg.AppConfig.String() + "/")
		}
	}

	urlHandler := &app.URLHandler{
		BaseURL: cfg.AppConfig.BaseURL,
		Storage: &app.URLStorage{
			DBFileName: cfg.DBConfig.DBFileName,
		},
	}

	router := gin.Default()

	router.POST("/", urlHandler.ShrinkURLTextHandler)
	router.POST("/api/shorten", urlHandler.ShrinkURLJsonHandler)
	router.GET("/:id", urlHandler.UnwrapURLHandler)

	err := router.Run(cfg.ServerURLConfig.String())
	if err != nil {
		panic(err)
	}

}
