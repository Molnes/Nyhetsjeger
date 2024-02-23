package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/public_pages"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

type PublicPagesHandler struct {
	sharedData *config.SharedData
}

// Creates a new PublicPagesHandler
func NewPublicPagesHandler(sharedData *config.SharedData) *PublicPagesHandler {
	return &PublicPagesHandler{sharedData}
}

// Registers handlers for public pages
func (pph *PublicPagesHandler) RegisterPublicPages(e *echo.Echo) {
	e.GET("", pph.homePage)
	e.GET("/login", pph.loginPage)
	e.GET("/forbidden", pph.forbiddenPage)
}



func (pph *PublicPagesHandler) homePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, public_pages.HomePage())
}

func (pph *PublicPagesHandler) loginPage(c echo.Context) error {
	session, err := pph.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
	if err != nil {
		return err
	}
	if session.Values["user"] != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/quiz")
	}

	return utils.Render(c, http.StatusOK, public_pages.LoginPage())
}

func (pph *PublicPagesHandler) forbiddenPage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, public_pages.ForbiddenPage())
}
