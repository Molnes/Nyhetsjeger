package middlewares

import (
	"testing"

	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
)

const expectedRoleError = "Expected role to be allowed"

// TestNewAuthorizationMiddleware tests the NewAuthorizationMiddleware function
func TestIsRoleAllowed(t *testing.T) {
	allowedRoles := []user_roles.Role{user_roles.QuizAdmin}
	mw := NewAuthorizationMiddleware(nil, allowedRoles)
	if !mw.isRoleAllowed(user_roles.QuizAdmin) {
		t.Error(expectedRoleError)
	}
}

// TestIsRoleAllowedTwo tests the NewAuthorizationMiddleware function
func TestIsRoleAllowedTwo(t *testing.T) {
	allowedRoles := []user_roles.Role{user_roles.QuizAdmin, user_roles.OrganizationAdmin}
	mw := NewAuthorizationMiddleware(nil, allowedRoles)
	if !mw.isRoleAllowed(user_roles.QuizAdmin) {
		t.Error(expectedRoleError)
	}
	if !mw.isRoleAllowed(user_roles.OrganizationAdmin) {
		t.Error(expectedRoleError)
	}
}

// TestIsRoleNotAllowed tests the NewAuthorizationMiddleware function
func TestIsRoleNotAllowed(t *testing.T) {
	allowedRoles := []user_roles.Role{user_roles.QuizAdmin}
	mw := NewAuthorizationMiddleware(nil, allowedRoles)
	if mw.isRoleAllowed(user_roles.User) {
		t.Error(expectedRoleError)
	}
}
