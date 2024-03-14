package handlers

import (
	"fmt"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/auth"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/middlewares"
	"github.com/google/uuid"
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
	if err != nil {
		return fmt.Errorf("failed to get user from user store: %s", err.Error())
	}

	if user == nil {
		accessTokenCypher, err := utils.Encrypt([]byte(token.AccessToken), ah.sharedData.CryptoKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt access token: %s", err.Error())
		}
		refreshTokenCypher, err := utils.Encrypt([]byte(token.RefreshToken), ah.sharedData.CryptoKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %s", err.Error())
		}

		newUser := users.User{
			ID:                 uuid.New(),
			SsoID:              googleUser.ID,
			Username:           "", // TODO set random username and let user re-roll it on registration
			Email:              googleUser.Email,
			Phone:              "",   // TODO get phone number from user at registration
			OptInRanking:       true, // TODO get opt in from user at registration
			Role:               user_roles.User,
			AccessTokenCypher:  accessTokenCypher,
			Token_expire:       token.Expiry,
			RefreshtokenCypher: refreshTokenCypher,
		}
		createdUser, err := users.CreateUser(ah.sharedData.DB, &newUser)
		if err != nil {
			return fmt.Errorf("failed to create user: %s", err.Error())
		}
		user = createdUser
	} else {
		accessTokenCypher, err := utils.Encrypt([]byte(token.AccessToken), ah.sharedData.CryptoKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt access token: %s", err.Error())
		}
		refreshTokenCypher, err := utils.Encrypt([]byte(token.RefreshToken), ah.sharedData.CryptoKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %s", err.Error())
		}

		user.AccessTokenCypher = accessTokenCypher
		user.Token_expire = token.Expiry
		user.RefreshtokenCypher = refreshTokenCypher
		err = users.UpdateUser(ah.sharedData.DB, user)
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
	redirectTo := "/quiz"
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
	return c.Redirect(http.StatusOK, "/")
}
