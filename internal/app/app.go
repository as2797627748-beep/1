package app

import (
	"net/http"

	"autocode-platform/internal/httpapi"
	"autocode-platform/internal/service"
)

type App struct {
	platform *service.Platform
	server   *httpapi.Server
}

func New() *App {
	platform := service.NewPlatform()
	server := httpapi.New(platform)
	return &App{platform: platform, server: server}
}

func (a *App) Handler() http.Handler {
	return a.server.Handler()
}
