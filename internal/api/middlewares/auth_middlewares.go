package middlewares

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/data/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/data/users"
	"github.com/Molnes/Nyhetsjeger/internal/data/users/user_roles"
	"github.com/labstack/echo/v4"
)

// Checks if the user is authenticated
// If not, returns a 401 Unauthorized response
func IsAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessions.Store.Get(c.Request(), sessions.SessionName)
		if err != nil {
			return err
		}
		if session.Values["user"] == nil {
			return c.JSON(http.StatusUnauthorized, "Unauthorized")
		}
		return next(c)
	}
}

// Checks if the user is an admin
// If not, returns a 403 Forbidden response
func IsAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessions.Store.Get(c.Request(), sessions.SessionName)
		if err != nil {
			return err
		}
		user := session.Values["user"].(users.User)
		if user.Role != user_roles.QuizAdmin {
			return c.JSON(http.StatusForbidden, "Forbidden")
		}
		return next(c)
	}
}

// Checks if the user is an organization admin
// If not, returns a 403 Forbidden response
func IsOrganizationAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessions.Store.Get(c.Request(), sessions.SessionName)
		if err != nil {
			return err
		}
		user := session.Values["user"].(users.User)
		if user.Role != user_roles.OrganizationAdmin {
			return c.JSON(http.StatusForbidden, "Forbidden")
		}
		return next(c)
	}
}
