package handlers

import (
	"fmt"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/auth"
	"github.com/Molnes/Nyhetsjeger/internal/data/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/data/users"
	"github.com/Molnes/Nyhetsjeger/internal/data/users/user_roles"
	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Registers the auth related handlers to the given echo group
func RegisterAuthHandlers(e *echo.Group) {
	e.GET("/google/login", oauthGoogleLogin)
	e.GET("/google/callback", oauthGoogleCallback)
	e.POST("/logout", logout)
}

// Redirects the user to the Google OAuth2 login page
func oauthGoogleLogin(c echo.Context) error {
	oauthState := auth.GenerateAndSetStateOauthCookie(c)
	url := auth.GoogleOauthConfig.AuthCodeURL(oauthState)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// Handles the callback from the Google OAuth2 login
// If the user is not in the user store, a new user is created
// If the user is in the user store, the user is updated with the new access token and refresh token
// The user is then logged in and a session is created
func oauthGoogleCallback(c echo.Context) error {
	oauthState, err := c.Cookie(auth.OauthStateCookieName)
	if err != nil {
		return fmt.Errorf("failed to get oauth state cookie: %s", err.Error())
	}
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

	user, err := users.GetUserBySsoID(database.DB, googleUser.ID)
	if err != nil {
		return fmt.Errorf("failed to get user from user store: %s", err.Error())
	}

	if user == nil {
		accessTokenCypher, err := utils.Encrypt([]byte(token.AccessToken))
		if err != nil {
			return fmt.Errorf("failed to encrypt access token: %s", err.Error())
		}
		refreshTokenCypher, err := utils.Encrypt([]byte(token.RefreshToken))
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
		createdUser, err := users.CreateUser(database.DB, &newUser)
		if err != nil {
			return fmt.Errorf("failed to create user: %s", err.Error())
		}
		user = createdUser
	} else {
		accessTokenCypher, err := utils.Encrypt([]byte(token.AccessToken))
		if err != nil {
			return fmt.Errorf("failed to encrypt access token: %s", err.Error())
		}
		refreshTokenCypher, err := utils.Encrypt([]byte(token.RefreshToken))
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %s", err.Error())
		}

		user.AccessTokenCypher = accessTokenCypher
		user.Token_expire = token.Expiry
		user.RefreshtokenCypher = refreshTokenCypher
		err = users.UpdateUser(database.DB, user)
		if err != nil {
			return fmt.Errorf("failed to update user: %s", err.Error())
		}
	}

	session, err := sessions.Store.New(c.Request(), sessions.SessionName)
	if err != nil {
		return fmt.Errorf("failed to create session: %s", err.Error())
	}
	session.Values["user"] = user
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("failed to save session: %s", err.Error())
	}

	cookieRedirectTo, err := c.Cookie("redirect-after-login")
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/quiz")
	}
	redirectTo := cookieRedirectTo.Value

	return c.Redirect(http.StatusTemporaryRedirect, redirectTo)
}

// Logs the user out by deleting the session
func logout(c echo.Context) error {
	session, err := sessions.Store.Get(c.Request(), sessions.SessionName)
	if err != nil {
		return fmt.Errorf("failed to get session: %s", err.Error())
	}
	session.Options.MaxAge = -1
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("failed to save session: %s", err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
