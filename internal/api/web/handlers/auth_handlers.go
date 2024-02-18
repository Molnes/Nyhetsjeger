package handlers

import (
	"fmt"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/auth"
	"github.com/labstack/echo/v4"
)

func RegisterAuthHandlers(e *echo.Group) {
	e.GET("/google/login", oauthGoogleLogin)
	e.GET("/google/callback", oauthGoogleCallback)
}

func oauthGoogleLogin(c echo.Context) error {
	oauthState := auth.GenerateStateOauthCookie(c)
	url := auth.GoogleOauthConfig.AuthCodeURL(oauthState)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func oauthGoogleCallback(c echo.Context) error {
	oauthState, _ := c.Cookie(auth.OauthStateCookieName)
	if c.FormValue("state") != oauthState.Value {
		return c.JSON(http.StatusUnauthorized, "invalid oauth state")
	}

	token, err := auth.GoogleOauthConfig.Exchange(c.Request().Context(), c.FormValue("code"))
	if err != nil {
		return fmt.Errorf("code exchange failed: %s", err.Error())
	}

	googleUser, err := auth.GetGoogleUserData(token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get user info: %s", err.Error())
	}

	// TODO
	// store token and user data in db
	// create a session

	return c.JSON(http.StatusOK, googleUser)
}
