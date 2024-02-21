package users

import (
	"database/sql"
	"encoding/gob"
	"time"

	"database/sql"
	"encoding/gob"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/data/users/user_roles"

	"github.com/Molnes/Nyhetsjeger/internal/data/users/user_roles"
	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID
	SsoID              string
	Username           string
	Email              string
	Phone              string
	OptInRanking       bool
	Role               user_roles.Role
	AccessTokenCypher  []byte
	Token_expire       time.Time
	RefreshtokenCypher []byte
}

func init() {
	// Register the User struct for gob encoding
	gob.Register(User{})
}

func GetUserByID(db *sql.DB, id uuid.UUID) (*User, error) {
	row := db.QueryRow(
		`SELECT id, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token
		FROM users
		WHERE id = $1`,
		id)
	return scanUserFromFullRow(row)
}
func GetUserBySsoID(db *sql.DB, sso_id string) (*User, error) {
	row := db.QueryRow(
		`SELECT id, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token
		FROM users
		WHERE sso_user_id = $1`,
		sso_id)
	return scanUserFromFullRow(row)
}

func CreateUser(db *sql.DB, user *User) (*User, error) {
	row := db.QueryRow(
		`INSERT INTO users (id, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token`,
		user.ID, user.SsoID, user.Email, user.Phone, user.OptInRanking, user.Role.String(), user.AccessTokenCypher, user.Token_expire, user.RefreshtokenCypher)
	return scanUserFromFullRow(row)
}

// Updates the user with the ID of the user provided
func UpdateUser(db *sql.DB, user *User) error {

	_, err := db.Exec(
		`UPDATE users
		SET sso_user_id = $2, email = $3, phone = $4, opt_in_ranking = $5, role = $6, access_token = $7, token_expires_at = $8, refresh_token = $9
		WHERE id = $1`,
		user.ID, user.SsoID, user.Email, user.Phone, user.OptInRanking, user.Role.String(), user.AccessTokenCypher, user.Token_expire, user.RefreshtokenCypher)
	return err
}

func scanUserFromFullRow(row *sql.Row) (*User, error) {
	user := User{}
	roleString := ""
	err := row.Scan(
		&user.ID,
		&user.SsoID,
		&user.Email,
		&user.Phone,
		&user.OptInRanking,
		&roleString,
		&user.AccessTokenCypher,
		&user.Token_expire,
		&user.RefreshtokenCypher,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	user.Role = user_roles.RoleFromString(roleString)
	return &user, err
}
