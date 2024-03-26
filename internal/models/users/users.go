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
	Username           string
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

// Returns a user from the database with the uuid provided
func GetUserByID(db *sql.DB, id uuid.UUID) (*User, error) {
	row := db.QueryRow(
		`SELECT id, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token,
		CONCAT(username_adjective, ' ', username_noun) AS username
		FROM users
		WHERE id = $1`,
		id)
	return scanUserFromFullRow(row)
}

// Returns a user from the database with the SSO ID provided
func GetUserBySsoID(db *sql.DB, sso_id string) (*User, error) {
	row := db.QueryRow(
		`SELECT id, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token,
		CONCAT(username_adjective, ' ', username_noun) AS username
		FROM users
		WHERE sso_user_id = $1`,
		sso_id)
	return scanUserFromFullRow(row)
}

// Creates a new user in the database
func CreateUser(db *sql.DB, user *User, ctx context.Context) (*User, error) {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var adjective string
	var noun string
	if err = tx.QueryRowContext(ctx,
		`SELECT *
			FROM available_usernames 
			OFFSET floor(random() * (
				SELECT COUNT(*) FROM available_usernames)
		) 
		LIMIT 1;`).Scan(&adjective, &noun); err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx,
		`INSERT INTO users (id, sso_user_id, username_adjective, username_noun,email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
		--RETURNING id, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token
		`,
		user.ID, user.SsoID, adjective, noun, user.Email, user.Phone, user.OptInRanking, user.Role.String(),
		user.AccessTokenCypher, user.Token_expire, user.RefreshtokenCypher)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return GetUserByID(db, user.ID)
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
		&user.Username,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	user.Role = user_roles.RoleFromString(roleString)
	return &user, err
}

// Assigns a random username to a user in the database
func AssignUsernameToUser(db *sql.DB, userID uuid.UUID, ctx context.Context) (string, error) {
	var username string
	err := db.QueryRow(`
			UPDATE users
			SET
				username_adjective = random_username.adjective,
				username_noun = random_username.noun
			FROM (
				SELECT adjective, noun
            	FROM available_usernames 
            	OFFSET floor(random() * (
                	SELECT COUNT(*) FROM available_usernames)
        		) 
        	LIMIT 1) AS random_username
			WHERE users.id = $1
			RETURNING CONCAT(random_username.adjective, ' ', random_username.noun);`,
		userID,
	).Scan(&username)

	return username, err
}

// Assigns the phone number to a user in the database
func AssignPhonenumberToUser(db *sql.DB, userID uuid.UUID, phonenumber string) error {
	_, err := db.Exec(
		`UPDATE users
			SET phone = $2
			WHERE id = $1`,
		userID, phonenumber,
	)
	return err
}

// Assigns the opt-in ranking value to a user in the database
func AssignOptInRankingToUser(db *sql.DB, userID uuid.UUID, optInRanking bool) error {
	_, err := db.Exec(
		`UPDATE users
			SET opt_in_ranking = $2
			WHERE id = $1`,
		userID, optInRanking,
	)
	return err
}

func DeleteUserByID(db *sql.DB, userID uuid.UUID) error {
	_, err := db.Exec(
		`DELETE FROM users
			WHERE id = $1`,
		userID,
	)
	return err
}
