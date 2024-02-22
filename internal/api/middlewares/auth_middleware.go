package middlewares

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/data/sessions"
	"github.com/labstack/echo/v4"
)

type AuthenticationMiddleware struct {
	redirectToLogin bool
}

func NewAuthenticationMiddleware(redirectToLogin bool) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{redirectToLogin}
}

// Checks if the user is authenticated
// If not, returns a 401 Unauthorized response
func (am *AuthenticationMiddleware) EncofreAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessions.Store.Get(c.Request(), sessions.SessionName)
		if err != nil {
			return err
		}
		if session.Values["user"] == nil {
			if am.redirectToLogin {
				userPath := c.Request().URL.Path
				cookie := http.Cookie{
					Name:   "redirect-after-login",
					Path:   "/",
					Value:  userPath,
					MaxAge: 3600,
				}
				c.SetCookie(&cookie)
				if err != nil {
					return err
				}
				return c.Redirect(http.StatusFound, "/login")
			} else {

				return c.JSON(http.StatusUnauthorized, "Unauthorized")
			}

		}
		return next(c)
	}
}