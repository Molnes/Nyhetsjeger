package middlewares

import (
	"fmt"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/data/users"
	"github.com/labstack/echo/v4"
)

type AuthenticationMiddleware struct {
	sharedData      *config.SharedData
	redirectToLogin bool
}

const (
	REDIRECT_COOKIE_NAME = "redirect-after-login"
	USER_ID_KEY          = "userID"
)

func NewAuthenticationMiddleware(data *config.SharedData, redirectToLogin bool) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{data, redirectToLogin}
}

// Checks if the user is authenticated
// If not, returns a 401 Unauthorized response
func (am *AuthenticationMiddleware) EncofreAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := am.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
		if err != nil {
			return err
		}
		userData := session.Values[sessions.USER_DATA_VALUE]
		if userData != nil {
			sessiondata, ok := userData.(users.UserSessionData)
			if !ok {
				return fmt.Errorf("AuthenticationMiddleware: failed to cast user data to UserSessionData")
			}
			userID := sessiondata.ID
			c.Set(USER_ID_KEY, userID)
			return next(c)
		} else {
			if am.redirectToLogin {
				userPath := c.Request().URL.Path
				cookie := http.Cookie{
					Name:   REDIRECT_COOKIE_NAME,
					Path:   "/",
					Value:  userPath,
					MaxAge: 3600,
				}
				c.SetCookie(&cookie)
				return c.Redirect(http.StatusFound, "/login")
			} else {
				return c.JSON(http.StatusUnauthorized, "Unauthorized")
			}
		}
	}
}
