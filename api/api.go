package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/api/web/views"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// This custom Render replaces Echo's echo.Context.Render() with templ's templ.Component.Render().
func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}
func Api() {
         e := echo.New()
         e.GET("/", func(c echo.Context) error {
                 return Render(c, http.StatusOK, views.Layout(views.Index()))
         })

         // Return static files from the "static" folder.
         e.Static("/static", "assets")

         e.Logger.Fatal(e.Start(":8080"))
}
         

