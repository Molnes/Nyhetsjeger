package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/labstack/echo/v4"
)

type publicApiHandler struct {
	sharedData *config.SharedData
}

// Creates a new PublicApiHandler
func NewPublicApiHandler(sharedData *config.SharedData) *publicApiHandler {
	return &publicApiHandler{sharedData}
}

// Registers the public api handlers to the given echo group
func (h *publicApiHandler) RegisterPublicApiHandlers(g *echo.Group) {
	g.POST("answer", h.postAnswer)
}

func (h *publicApiHandler) postAnswer(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
