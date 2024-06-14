package main

import (
	"flag"
	"github.com/amphiphile/urlshrtnr/internal/app"
	"github.com/amphiphile/urlshrtnr/internal/config"
	"github.com/gin-gonic/gin"
)

var cfg = config.Config{
	ServerUrlConfig: config.ServerUrlConfig{
		ServerHost: "localhost",
		ServerPort: 8080,
	},
	AppConfig: config.AppConfig{
		BaseUrl: "http://localhost:8080/",
	},
	DBFileName: "db.json",
}

func main() {

	flag.Var(cfg.ServerUrlConfig, "a", "HTTP server startup address")
	flag.Var(cfg.AppConfig, "b", "Base address of the shortened URL")

	flag.Parse()

	urlHandler := &app.UrlHandler{
		BaseUrl: cfg.AppConfig.BaseUrl,
		Storage: &app.UrlStorage{
			DBFileName: cfg.DBFileName,
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
