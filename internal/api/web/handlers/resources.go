package handlers

import "github.com/labstack/echo/v4"



func RegisterResourcesHandlers(e *echo.Echo) {
	e.Static("/static", "assets")
}