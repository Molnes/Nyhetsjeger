package user_roles

import (
	"testing"
)

func TestIsRoleAdmin(t *testing.T) {
	quizAdmin := QuizAdmin
	if !quizAdmin.IsAdministrator() {
		t.Error("Expected quizAdmin to be an administrator role")
	}
}
func TestIsRoleAdminOrganizationAdmin(t *testing.T) {
	organizationAdmin := OrganizationAdmin
	if !organizationAdmin.IsAdministrator() {
		t.Error("Expected organizationAdmin to be an administrator role")
	}
}

func TestIsRoleAdminUser(t *testing.T) {
	userRole := User
	if userRole.IsAdministrator() {
		t.Error("Expected userRole to not be an administrator role")
	}
}
