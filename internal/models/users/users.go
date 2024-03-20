package users

import (
	"context"
	"database/sql"
	"encoding/gob"
	"log"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID
	SsoID              string
	usernameAdjective  string
	usernameNoun       string
	Email              string
	Phone              string
	OptInRanking       bool
	Role               user_roles.Role
	AccessTokenCypher  []byte
	Token_expire       time.Time
	RefreshtokenCypher []byte
}

type UserSessionData struct {
	ID    uuid.UUID
	SsoID string
	Email string
}

func (u *User) IntoSessionData() UserSessionData {
	return UserSessionData{
		ID:    u.ID,
		SsoID: u.SsoID,
		Email: u.Email,
	}
}

func init() {
	// Register the UserSessionData struct for gob encoding, needed for session storage
	gob.Register(UserSessionData{})
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
		`INSERT INTO users (id, username, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token`,
		user.ID, user.Username, user.SsoID, user.Email, user.Phone, user.OptInRanking, user.Role.String(), user.AccessTokenCypher, user.Token_expire, user.RefreshtokenCypher)
	return scanUserFromFullRow(row)
}

// Returns the role of the user with the ID provided
func GetUserRole(db *sql.DB, id uuid.UUID) (user_roles.Role, error) {
	var role string
	err := db.QueryRow(
		`SELECT role
		FROM users
		WHERE id = $1`,
		id).Scan(&role)
	return user_roles.RoleFromString(role), err
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

// Returns a random available username from the database
func getRandomAvailableUsername(db *sql.DB) (string, error) {
	var username string
	err := db.QueryRow(
		`SELECT *
			FROM available_usernames 
			OFFSET floor(random() * (
				SELECT COUNT(*) FROM available_usernames)
		) 
		LIMIT 1;`).Scan(&username)
	return username, err
}
