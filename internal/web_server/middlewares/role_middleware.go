package middlewares

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"

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
		role, ok := c.Get(user_roles.ROLE_CONTEXT_KEY).(user_roles.Role)
		if !ok {
			return fmt.Errorf("EnforceRole: Role not in context. Authentication middleware must be used before EnforceRole")
		}
		if !am.isRoleAllowed(role) {
			return echo.NewHTTPError(http.StatusForbidden, "You are not allowed to access this resource")
		}
		return next(c)
	}
}
