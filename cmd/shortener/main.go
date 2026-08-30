package main

import (
	"net/http"
	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/infra/httpserver"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"
)

func main() {
	config := config.Load()
	repository := repository.NewStore()
	controller := service.NewResourceController(repository)
	h := handler.NewResourceHandler(config.BaseURL, controller)
	router := handler.NewRouter(h)
	server := httpserver.NewServer(config.HTTPServer.URL, router)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
