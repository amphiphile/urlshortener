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

	flag.Var(&cfg.ServerUrlConfig, "a", "HTTP server startup address")
	flag.Var(&cfg.AppConfig, "b", "Base address of the shortened URL")

	flag.Parse()

	if serverAddress := cfg.ServerUrlConfig.String(); serverAddress == "" {
		_ = cfg.ServerUrlConfig.Set(config.GetFromEnv("SERVER_ADDRESS", "localhost:8080"))

	}
	if baseUrl := cfg.AppConfig.String(); baseUrl == "" {
		_ = cfg.AppConfig.Set(config.GetFromEnv("BASE_URL", "http://localhost:8080/"))
		if !strings.HasSuffix(cfg.AppConfig.String(), "/") {
			_ = cfg.AppConfig.Set(cfg.AppConfig.String() + "/")
		}
	}

	urlHandler := &app.UrlHandler{
		BaseUrl: cfg.AppConfig.BaseUrl,
		Storage: &app.UrlStorage{
			DBFileName: cfg.DBConfig.DBFileName,
		},
	}

	router := gin.Default()

	router.POST("/", urlHandler.ShrinkUrlTextHandler)
	router.POST("/api/shorten", urlHandler.ShrinkUrlJsonHandler)
	router.GET("/:id", urlHandler.UnwrapUrlHandler)

	err := router.Run(cfg.ServerUrlConfig.String())
	if err != nil {
		panic(err)
	}

}
