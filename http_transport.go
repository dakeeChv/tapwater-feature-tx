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
	provinces, err := h.aqService.Province(ctx)
	if err != nil {
		hs := httpStatusPbFromRPC(gRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(http.StatusOK, b)
	}
	return c.JSON(http.StatusOK, echo.Map{"provinces": provinces})
}

func (h *HTTPTransport) Info(c echo.Context) error {
	var in InfoQuery
	if err := c.Bind(&in); err != nil {
		hs := httpStatusPbFromRPC(StatusBindingFailure)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(http.StatusOK, b)
	}
	
	ctx := c.Request().Context()
	info, err := h.aqService.Info(ctx, in)
	if err != nil {
		hs := httpStatusPbFromRPC(gRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(http.StatusOK, b)
	}
	return c.JSON(http.StatusOK, echo.Map{"info": info})
}

func (h *HTTPTransport) install(e *echo.Echo) {
	v1 := e.Group("/v1/billing/tapwater")
	v1.POST("/provinces", h.Province)
	v1.POST("/info", h.Info)
}
