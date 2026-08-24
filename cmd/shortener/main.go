package main

import (
	"net/http"
	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/infra/httpserver"
	"ya_url_shortener/internal/service"
)

func main() {
	c := config.Load()
	controller := service.NewResourceController()
	h := handler.NewResourceHandler(controller)
	router := handler.NewRouter(h)
	server := httpserver.NewServer(c, router)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

}
