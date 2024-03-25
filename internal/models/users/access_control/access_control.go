package access_control

import (
	"database/sql"
	"errors"

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

// Adds a QuizAdmin role to a user with the given email.
//
// If the user does not exist, the role is preassigned and will be assigned upon registration.
func AddAdmin(db *sql.DB, email string) (*UserAdmin, error) {
	var isActive bool
	err := assignRoleToUseByEmail(db, email, user_roles.QuizAdmin)
	if err == nil {
		isActive = true
	} else {
		if err != errNoUserFound {
			return nil, err
		} else {
			err = preAssignRoleToUserByEmail(db, email, user_roles.QuizAdmin)
			if err != nil {
				return nil, err
			}
			isActive = false
		}
	}

	return &UserAdmin{
		Email:    email,
		IsActive: isActive,
		Role:     user_roles.QuizAdmin,
	}, nil
}

var errNoUserFound = errors.New("no user with given email found")

// Assigns a role to a user with the given email.
func assignRoleToUseByEmail(db *sql.DB, email string, role user_roles.Role) error {
	result, err := db.Exec(`
	UPDATE users
	SET role=$1
	WHERE email=$2;`, role.String(), email)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected < 1 {
		return errNoUserFound
	}
	return nil
}

// Preassigns a role to a user with the given email.
func preAssignRoleToUserByEmail(db *sql.DB, email string, role user_roles.Role) error {
	_, err := db.Exec(`
	INSERT INTO preassigned_roles(email, role)
	VALUES ($1, $2);`, email, role.String())
	if err != nil {
		return err
	}
	return nil
}

// Revokes the QuizAdmin role from a user with the given email.
func RevokeAdmin(db *sql.DB, email string) error {
	return assignRoleToUseByEmail(db, email, user_roles.User)
}

// Removes a preassigned QuizAdmin role from the given email.
func RemovePreassignedAdmin(db *sql.DB, email string) error {
	_, err := db.Exec(`
	DELETE FROM preassigned_roles
	WHERE email=$1;`, email)
	return err
}
