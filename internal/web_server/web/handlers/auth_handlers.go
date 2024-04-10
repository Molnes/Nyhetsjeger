package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/auth"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/middlewares"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	sharedData        *config.SharedData
	googleOauthConfig *oauth2.Config
}

// Creates a new handler for auth related requests
func NewAuthHandler(sharedData *config.SharedData, googleOauthConfig *oauth2.Config) *AuthHandler {
	return &AuthHandler{sharedData, googleOauthConfig}
}

// Registers the auth related handlers to the given echo group
func (ah *AuthHandler) RegisterAuthHandlers(g *echo.Group) {
	g.GET("/google/login", ah.oauthGoogleLogin)
	g.GET("/google/callback", ah.oauthGoogleCallback)
	g.POST("/logout", ah.logout)
}

// Redirects the user to the Google OAuth2 login page
func (ah *AuthHandler) oauthGoogleLogin(c echo.Context) error {
	oauthState := auth.GenerateAndSetStateOauthCookie(c)
	url := ah.googleOauthConfig.AuthCodeURL(oauthState)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// Handles the callback from the Google OAuth2 login
// If the user is not in the user store, a new user is created
// If the user is in the user store, the user is updated with the new access token and refresh token
// The user is then logged in and a session is created
func (ah *AuthHandler) oauthGoogleCallback(c echo.Context) error {
	oauthState, err := c.Cookie(auth.OAUTH_STATE_COOKIE)
	if err != nil {
		return fmt.Errorf("failed to get oauth state cookie: %s", err.Error())
	}
	if c.FormValue("state") != oauthState.Value {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid oauth state")
	}

	token, err := ah.googleOauthConfig.Exchange(c.Request().Context(), c.FormValue("code"))
	if err != nil {
		return fmt.Errorf("code exchange failed: %s", err.Error())
	}

	googleUser, err := auth.GetGoogleUserData(token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get user info: %s", err.Error())
	}

	user, err := users.GetUserBySsoID(ah.sharedData.DB, googleUser.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get user from user store: %s", err.Error())
	}

	if err == sql.ErrNoRows {
		newUser := users.PartialUser{
			SsoID:        googleUser.ID,
			Email:        googleUser.Email,
			AccessToken:  token.AccessToken,
			TokenExpire:  token.Expiry,
			RefreshToken: token.RefreshToken,
		}
		createdUser, err := users.CreateUser(ah.sharedData.DB, c.Request().Context(), &newUser)
		if err != nil {
			return fmt.Errorf("failed to create user: %s", err.Error())
		}
		user = createdUser
	} else {
		err = users.UpdateUserToken(ah.sharedData.DB, user.ID, token.AccessToken, token.Expiry, token.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to update user: %s", err.Error())
		}
	}

	session, err := ah.sharedData.SessionStore.New(c.Request(), sessions.SESSION_NAME)
	if err != nil {
		return fmt.Errorf("failed to create session: %s", err.Error())
	}

	userSessionData := user.IntoSessionData()
	session.Values[sessions.USER_DATA_VALUE] = userSessionData
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("failed to save session: %s", err.Error())
	}

	cookieRedirectTo, err := c.Cookie(middlewares.REDIRECT_COOKIE_NAME)
	redirectTo := "/"
	if err == nil { // if redirect cookie is set, use it
		redirectTo = cookieRedirectTo.Value

		replacementCookie := http.Cookie{
			Name:   middlewares.REDIRECT_COOKIE_NAME,
			Path:   "/",
			Value:  "",
			MaxAge: -1,
		}
		c.SetCookie(&replacementCookie)
	}

	return c.Redirect(http.StatusTemporaryRedirect, redirectTo)
}

// Logs the user out by deleting the session
func (ah *AuthHandler) logout(c echo.Context) error {
	session, err := ah.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
	if err != nil {
		return fmt.Errorf("failed to get session: %s", err.Error())
	}
	session.Options.MaxAge = -1
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("failed to save session: %s", err.Error())
	}
	c.Response().Header().Set("HX-Redirect", "/")
	return c.NoContent(http.StatusNoContent)
}
