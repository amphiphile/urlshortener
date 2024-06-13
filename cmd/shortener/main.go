package main

import (
	"github.com/amphiphile/urlshrtnr/internal/app"
	"github.com/gin-gonic/gin"
)

func main() {
	urlHandler := &app.UrlHandler{
		Storage: &app.UrlStorage{
			DBFileName: "db.json",
		},
	}

	router := gin.Default()
	router.POST("/", urlHandler.ShrinkUrlHandler)
	router.GET("/:id", urlHandler.UnwrapUrlHandler)
	err := router.Run()
	if err != nil {
		panic(err)
	}

	//mux := http.NewServeMux()
	//mux.HandleFunc(`/`, urlHandler.RequestHandler)
	//err := http.ListenAndServe(`:8080`, mux)
	//
	//if err != nil {
	//	panic(err)
	//}

}
