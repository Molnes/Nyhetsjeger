package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/public_pages"
	"github.com/Molnes/Nyhetsjeger/internal/data/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

func RegisterPublicPages(e *echo.Echo) {
	e.GET("", homePage)
	e.GET("/login", loginPage)
}

func homePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, public_pages.HomePage())
}

func loginPage(c echo.Context) error {
	session, err := sessions.Store.Get(c.Request(), sessions.SessionName)
	if err != nil {
		return err
	}
	if session.Values["user"] != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/quiz")
	}

	return utils.Render(c, http.StatusOK, public_pages.LoginPage())
}