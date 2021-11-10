package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/encoding/protojson"
)

type HTTPTransport struct {
	aqService *AquaService
}

func (h *HTTPTransport) Province(c echo.Context) error {
	ctx := c.Request().Context()
	provinces, err := h.aqService.GetProvince(ctx)
	if err != nil {
		hs := httpStatusPbFromRPC(gRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(http.StatusOK, b)
	}
	return c.JSON(http.StatusOK, echo.Map{"provinces": provinces})
}

func (h *HTTPTransport) install(e *echo.Echo) {
	v1 := e.Group("/v1/tapwater")
	v1.POST("/province", h.Province)
}