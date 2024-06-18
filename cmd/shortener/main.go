package main

import (
	"github.com/amphiphile/urlshortener/internal/app"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {

	cfg := new(Config)

	err := parseConfig(cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}

	urlHandler := &app.URLHandler{
		Storage: &app.URLStorage{
			BaseURL:    cfg.AppConfig.BaseURL,
			DBFileName: cfg.DBConfig.DBFileName,
		},
	}

	router := setupRouter(*urlHandler)

	err = router.Run(cfg.ServerURLConfig.String())
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func setupRouter(urlHandler app.URLHandler) *gin.Engine {
	router := gin.Default()
	router.POST("/", urlHandler.HandleShrinkURLText)
	router.POST("/api/shorten", urlHandler.HandleShrinkURLJSON)
	router.GET("/:id", urlHandler.HandleUnwrapURL)
	return router
}
