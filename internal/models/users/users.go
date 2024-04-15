package users

import (
	"context"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/google/uuid"
)

const (
	USER_ID_CONTEXT_KEY = "userID" // The key used to store the user ID in the context
)

type User struct {
	ID                 uuid.UUID
	SsoID              string
	Username           string
	Email              string
	Phone              string
	OptInRanking       bool
	AcceptedTerms      bool
	Role               user_roles.Role
	AccessTokenCypher  []byte
	TokenExpire        time.Time
	RefreshTokenCypher []byte
}

// Parital User struct contains only fields needed for creating a new user
type PartialUser struct {
	SsoID        string
	Email        string
	AccessToken  string
	TokenExpire  time.Time
	RefreshToken string
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
		`SELECT id, sso_user_id, email, phone, opt_in_ranking, accepted_terms, role, access_token, token_expires_at, refresh_token,
		CONCAT(username_adjective, ' ', username_noun) AS username
		FROM users
		WHERE id = $1`,
		id)
	return scanUserFromFullRow(row)
}

// Returns a user from the database with the email provided
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow(
		`SELECT id, sso_user_id, email, phone, opt_in_ranking, accepted_terms, role, access_token, token_expires_at, refresh_token,
		CONCAT(username_adjective, ' ', username_noun) AS username
		FROM users
		WHERE email = $1`,
		email)
	return scanUserFromFullRow(row)
}

// Returns a user from the database with the SSO ID provided
func GetUserBySsoID(db *sql.DB, ssoID string) (*User, error) {
	row := db.QueryRow(
		`SELECT id, sso_user_id, email, phone, opt_in_ranking, accepted_terms, role, access_token, token_expires_at, refresh_token,
		CONCAT(username_adjective, ' ', username_noun) AS username
		FROM users
		WHERE sso_user_id = $1`,
		ssoID)
	return scanUserFromFullRow(row)
}

// Creates a new user in the database
func CreateUser(db *sql.DB, ctx context.Context, partialUser *PartialUser) (*User, error) {
	accessTokenCypher := []byte("TODO")
	refreshtokenCypher := []byte("TODO")
	user := User{
		ID:                 uuid.New(),
		SsoID:              partialUser.SsoID,
		Email:              partialUser.Email,
		Phone:              "No phone number provided",
		OptInRanking:       true,
		Role:               user_roles.User,
		AccessTokenCypher:  accessTokenCypher,
		TokenExpire:        partialUser.TokenExpire,
		RefreshTokenCypher: refreshtokenCypher,
	}

	row := db.QueryRow(
		`INSERT INTO users
		(id, sso_user_id, email, phone, opt_in_ranking, role, access_token, token_expires_at, refresh_token, username_adjective, username_noun)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, random_username.adjective, random_username.noun
		FROM (
			SELECT adjective, noun
			FROM available_usernames 
			OFFSET floor(random() * (SELECT COUNT(*) FROM available_usernames)) 
		LIMIT 1) AS random_username
		RETURNING
		id, sso_user_id, email, phone, opt_in_ranking, accepted_terms, role, access_token,
		token_expires_at, refresh_token,CONCAT(username_adjective, ' ', username_noun) AS username;`,
		user.ID, user.SsoID, user.Email, user.Phone, user.OptInRanking, user.Role.String(),
		user.AccessTokenCypher, user.TokenExpire, user.RefreshTokenCypher)

	insertedUser, err := scanUserFromFullRow(row)
	if err != nil {
		return nil, err
	}
	newRole, err := applyPreassignedRole(db, ctx, user.Email)
	if err != nil {
		if err == errNoPreassignedRole {
			newRole = user_roles.User
		} else {
			return nil, err
		}
	}
	insertedUser.Role = newRole

	return insertedUser, nil
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

// Updates the token of the given user in the database.
func UpdateUserToken(db *sql.DB, userId uuid.UUID, newAccessToken string, newExpiry time.Time, newRefreshToken string) error {
	accessTokenCypher := []byte("TODO")
	refreshTokenCypher := []byte("TODO")
	_, err := db.Exec(
		`UPDATE users
		SET access_token = $2, token_expires_at = $3, refresh_token = $4
		WHERE id = $1`,
		userId, accessTokenCypher, newExpiry, refreshTokenCypher,
	)
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
		&user.AcceptedTerms,
		&roleString,
		&user.AccessTokenCypher,
		&user.TokenExpire,
		&user.RefreshTokenCypher,
		&user.Username,
	)
	if err != nil {
		return nil, err
	}
	user.Role = user_roles.RoleFromString(roleString)
	return &user, nil
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

// If there is no preassigned role for the given email.
var errNoPreassignedRole = errors.New("no preassigned role found")

// Assigns the preassigned role to the user with the given email. The preassigned role is then removed.
//
// Returns the role that was assigned.
//
// If there is no preassigned role for the given email, ErrNoPreassignedRole is returned.
func applyPreassignedRole(db *sql.DB, ctx context.Context, email string) (user_roles.Role, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
	UPDATE users u
	SET role=pr.role
	FROM preassigned_roles pr
	WHERE u.email=$1 AND pr.email=$1
	RETURNING u.role;`, email)
	var roleString string
	err = row.Scan(&roleString)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errNoPreassignedRole
		}
		return 0, err
	}

	_, err = tx.ExecContext(ctx, `
	DELETE FROM preassigned_roles
	WHERE email=$1;`, email)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return user_roles.RoleFromString(roleString), nil
}

// Sets the accepted_terms value in database for the user with given id.
func UpdateAcceptedTermsByUserID(db *sql.DB, userid uuid.UUID, isAccepted bool) error {
	result, err := db.Exec(`
	UPDATE users
	SET accepted_terms = $2
	WHERE id = $1;`, userid, isAccepted)
	if err != nil {
		return err
	}
	rowsAffecter, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffecter != 1 {
		return fmt.Errorf("users.UpdateAcceptedTermsByUserID: no rows affected, no user found")
	}

	return nil
}
