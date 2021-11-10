package main

import (
	"github.com/labstack/echo/v4"
)

type HTTPTransport struct {
	aqService *AquaService
}

func (h *HTTPTransport) install(e *echo.Echo) {
	v1 := e.Group("/v1/tapwater")
	_ = v1
}
