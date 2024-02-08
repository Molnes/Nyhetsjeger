package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

func RegisterDashboardHandlers(e *echo.Echo) {
	g := e.Group("/dashboard")
	g.GET("", dashboardHomePage)
}

func dashboardHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, pages.DashboardPage())
}
