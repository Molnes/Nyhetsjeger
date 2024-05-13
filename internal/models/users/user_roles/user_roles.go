package user_roles

type Role int

const (
	User Role = iota
	QuizAdmin
	OrganizationAdmin
)

const (
	ROLE_CONTEXT_KEY        = "user-role" // The key used to store the user role in the context
	userString              = "user"
	quizAdminString         = "quiz_admin"
	organizationAdminString = "organization_admin"
)

// String returns the string representation of the role
func (r Role) String() string {
	var roleString string
	switch r {
	case User:
		roleString = userString
	case QuizAdmin:
		roleString = quizAdminString
	case OrganizationAdmin:
		roleString = organizationAdminString
	}
	return roleString
}

// RoleFromString returns the role from the string representation
func RoleFromString(role string) Role {
	var r Role
	switch role {
	case userString:
		r = User
	case quizAdminString:
		r = QuizAdmin
	case organizationAdminString:
		r = OrganizationAdmin
	}
	return r
}

// IsAdministrator returns true if the role is an administrator role
func (r Role) IsAdministrator() bool {
	return r == QuizAdmin || r == OrganizationAdmin
}
