package main

import (
	"github.com/amphiphile/urlshrtnr/internal/app"
	"net/http"
)

func main() {
	urlHandler := &app.UrlHandler{
		Storage: &app.UrlStorage{
			DBFileName: "db.json",
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, urlHandler.RequestHandler)
	err := http.ListenAndServe(`:8080`, mux)

	if err != nil {
		panic(err)
	}

}
