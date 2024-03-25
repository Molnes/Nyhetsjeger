package middlewares

import (
	"net/http"
	"slices"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/Molnes/Nyhetsjeger/internal/utils"

	"github.com/labstack/echo/v4"
)

// Middleware for enforcing user roles
// Use AuthenticationMiddleware BEFORE this middleware, as it expects the user to be authenticated
type AuthorizationMiddleware struct {
	sharedData   *config.SharedData
	allowedRoles []user_roles.Role
}

// Creates a new AuthorizationMiddleware.
//
// AllowedRoles: A list of roles that are allowed to pass through the middleware.
//
// Roles not in the list will receive a 403 Forbidden response
func NewAuthorizationMiddleware(data *config.SharedData, allowedRoles []user_roles.Role) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{data, allowedRoles}
}

// Checks if the role is allowed to pass through the middleware
func (am *AuthorizationMiddleware) isRoleAllowed(role user_roles.Role) bool {
	return slices.Contains(am.allowedRoles, role)
}

// Checks if the user has a role that is allowed to pass through the middleware.
// If not, returns a 403 Forbidden response
//
// If the user is allowed, the role is added to the context under the key user_roles.ROLE_CONTEXT_KEY
func (am *AuthorizationMiddleware) EnforceRole(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := am.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
		if err != nil {
			return err
		}
		user := session.Values["user"].(users.UserSessionData)
		role, err := users.GetUserRole(am.sharedData.DB, user.ID)
		if err != nil {
			return err
		}
		if !am.isRoleAllowed(role) {
			return echo.NewHTTPError(http.StatusForbidden, "You are not allowed to access this resource")
		}
		utils.AddToContext(c, user_roles.ROLE_CONTEXT_KEY, role)
		return next(c)
	}
}
