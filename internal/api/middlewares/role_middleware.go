package middlewares

import (
	"database/sql"
	"net/http"
	"slices"

	"github.com/Molnes/Nyhetsjeger/internal/data/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/data/users"
	"github.com/Molnes/Nyhetsjeger/internal/data/users/user_roles"
	"github.com/labstack/echo/v4"
)

// Middleware for enforcing user roles
// Use AuthenticationMiddleware BEFORE this middleware, as it expects the user to be authenticated
type AuthorizationMiddleware struct {
	db                             *sql.DB
	allowedRoles                   []user_roles.Role
	redirectToPermissionDeniedPage bool
}

// Creates a new AuthorizationMiddleware.
//
// AllowedRoles: A list of roles that are allowed to pass through the middleware.
//
// Roles not in the list will receive a 403 Forbidden response
func NewAuthorizationMiddleware(db *sql.DB, allowedRoles []user_roles.Role, redirect_to_page bool) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{db, allowedRoles, redirect_to_page}
}

func (am *AuthorizationMiddleware) isRoleAllowed(role user_roles.Role) bool {
	return slices.Contains(am.allowedRoles, role)
}

// Checks if the user is an admin
// If not, returns a 403 Forbidden response
func (am *AuthorizationMiddleware) EnforceRole(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessions.Store.Get(c.Request(), sessions.SESSION_NAME)
		if err != nil {
			return err
		}
		user := session.Values["user"].(users.UserSessionData)
		role, err := users.GetUserRole(am.db, user.ID)
		if err != nil {
			return err
		}
		if !am.isRoleAllowed(role) {
			if am.redirectToPermissionDeniedPage {
				return c.Redirect(http.StatusFound, "/forbidden")
			} else {
				return c.JSON(http.StatusForbidden, "Forbidden")
			}
		}
		return next(c)
	}
}
