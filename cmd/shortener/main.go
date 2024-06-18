package main

import (
	"flag"
	"github.com/amphiphile/urlshortener/internal/app"
	"github.com/amphiphile/urlshortener/internal/config"
	"github.com/gin-gonic/gin"
	"log"
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
	}

	urlHandler := &app.URLHandler{
		BaseURL: cfg.AppConfig.BaseURL,
		Storage: &app.URLStorage{
			DBFileName: cfg.DBConfig.DBFileName,
			BaseURL:    cfg.AppConfig.BaseURL,
		},
	}

	router := gin.Default()

	router.POST("/", urlHandler.HandleShrinkURLText)
	router.POST("/api/shorten", urlHandler.HandleShrinkURLJSON)
	router.GET("/:id", urlHandler.HandleUnwrapURL)

	err := router.Run(cfg.ServerURLConfig.String())
	if err != nil {
		log.Fatalf(err.Error())
	}

}
