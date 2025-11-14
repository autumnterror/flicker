package net

import (
	"net/http"

	"github.com/autumnterror/breezynotes/pkg/log"
	"github.com/autumnterror/breezynotes/views"
	"github.com/labstack/echo/v4"
)

// Healthz godoc
// @Summary check health of gateway
// @Description
// @Tags healthz
// @Produce json
// @Success 200 {object} views.SWGMessage
// @Router /api/health [get]
func (e *Echo) Healthz(c echo.Context) error {
	const op = "net.Healthz"
	log.Info(op, "")

	return c.JSON(http.StatusOK, views.SWGMessage{Message: "HEALTHZ"})
}
