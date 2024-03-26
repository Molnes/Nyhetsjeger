package access_control

import (
	"database/sql"
	"errors"

	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/lib/pq"
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

var ErrEmailAlreadyAdmin = errors.New("email is already an admin")

// Adds a QuizAdmin role to a user with the given email.
//
// If the user does not exist, the role is preassigned and will be assigned upon registration.
//
// Returns ErrEmailAlreadyAdmin if the user already has the role.
func AddAdmin(db *sql.DB, email string) (*UserAdmin, error) {
	var isActive bool
	err := assignRoleToUserByEmail(db, email, user_roles.QuizAdmin)
	if err == nil {
		isActive = true
	} else if err == errUserAlreadyHasRole {
		return nil, ErrEmailAlreadyAdmin
	} else {
		if err == errNoUserFound {
			err = preAssignRoleToUserByEmail(db, email, user_roles.QuizAdmin)
			if err != nil {
				return nil, err
			}
			isActive = false
		} else {
			return nil, err
		}
	}

	return &UserAdmin{
		Email:    email,
		IsActive: isActive,
		Role:     user_roles.QuizAdmin,
	}, nil
}

var errNoUserFound = errors.New("no user with given email found")
var errUserAlreadyHasRole = errors.New("user already has the given role")

// Assigns a role to a user with the given email.
func assignRoleToUserByEmail(db *sql.DB, email string, role user_roles.Role) error {
	user, err := users.GetUserByEmail(db, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return errNoUserFound
		} else {
			return err
		}
	}

	if user.Role == role {
		return errUserAlreadyHasRole
	}

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

	return err

}

// Preassigns a role to a user with the given email.
func preAssignRoleToUserByEmail(db *sql.DB, email string, role user_roles.Role) error {
	_, err := db.Exec(`
	INSERT INTO preassigned_roles(email, role)
	VALUES ($1, $2);`, email, role.String())
	if err != nil {
		// if unique constraint violated, the role is already preassigned
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrEmailAlreadyAdmin
		}
	}
	return err
}

var errNoPreassignedRoleForEmail = errors.New("no preassigned role for the given email")

// Removes a preassigned QuizAdmin role from the given email.
func removePreassignedAdmin(db *sql.DB, email string) error {
	result, err := db.Exec(`
	DELETE FROM preassigned_roles
	WHERE email=$1;`, email)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected < 1 {
		return errNoPreassignedRoleForEmail
	}
	return nil

}

var ErrNoAdminWithGivenEmail = errors.New("no admin with given email found")

// Removes the QuizAdmin from the given email. If the user exists the role is revoked. If the user does not exist, the preassigned role is removed.
//
// Returns ErrNoAdminWithGivenEmail if there is no user or preassigned role for the given email.
func RemoveAdminByEmail(db *sql.DB, email string) error {
	err := assignRoleToUserByEmail(db, email, user_roles.User)
	if err == nil {
		return nil
	} else if err == errUserAlreadyHasRole {
		return ErrNoAdminWithGivenEmail
	} else if err == errNoUserFound {
		err = removePreassignedAdmin(db, email)
		if err != nil {
			if err == errNoPreassignedRoleForEmail {
				return ErrNoAdminWithGivenEmail
			}
		}
	}
	return err
}
