package middlewares

import (
	"testing"

	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
)

func TestIsRoleAllowed(t *testing.T) {
	allowedRoles := []user_roles.Role{user_roles.QuizAdmin}
	mw := NewAuthorizationMiddleware(nil, allowedRoles)
	if !mw.isRoleAllowed(user_roles.QuizAdmin) {
		t.Error("Expected role to be allowed")
	}
}

func TestIsRoleAllowedTwo(t *testing.T) {
	allowedRoles := []user_roles.Role{user_roles.QuizAdmin, user_roles.OrganizationAdmin}
	mw := NewAuthorizationMiddleware(nil, allowedRoles)
	if !mw.isRoleAllowed(user_roles.QuizAdmin) {
		t.Error("Expected role to be allowed")
	}
	if !mw.isRoleAllowed(user_roles.OrganizationAdmin) {
		t.Error("Expected role to be allowed")
	}
}

func TestIsRoleNotAllowed(t *testing.T) {
	allowedRoles := []user_roles.Role{user_roles.QuizAdmin}
	mw := NewAuthorizationMiddleware(nil, allowedRoles)
	if mw.isRoleAllowed(user_roles.User) {
		t.Error("Expected role to not be allowed")
	}
}
