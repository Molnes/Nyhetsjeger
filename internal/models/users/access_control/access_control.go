package access_control

import (
	"database/sql"

	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
)

type UserAdmin struct {
	Email    string
	IsActive bool // Whether the role is in use or it waits for a user to register (preassigned role)
	Role     user_roles.Role
}

// Gets all users with role QuizAdmin either assigned or preassigned.
//
// Preassigned means user has not yet registered, but the role is reserved for them and will be assigned upon registration.
func GetAllAdmins(db *sql.DB) (*[]UserAdmin, error) {
	rows, err := db.Query(`
	SELECT u.email, true AS is_active, u.role
	FROM users u
	WHERE u.role=$1
	UNION
	SELECT pr.email, false AS is_active, pr.role
	FROM preassigned_roles pr
	WHERE pr.role=$1
	ORDER BY email;`, user_roles.QuizAdmin.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []UserAdmin
	for rows.Next() {
		var user UserAdmin
		roleString := ""
		if err := rows.Scan(&user.Email, &user.IsActive, &roleString); err != nil {
			return nil, err
		}
		user.Role = user_roles.RoleFromString(roleString)
		admins = append(admins, user)
	}

	return &admins, nil
}
