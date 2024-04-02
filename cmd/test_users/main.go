package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/Molnes/Nyhetsjeger/internal/models/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/antonlindstrom/pgstore"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// Simple server used in API tests to create sessions for different user roles.
//
// This code is used only in testing and is NEVER deployed to production.
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Test users: Error loading .env")
	}

	dburl, ok := os.LookupEnv("POSTGRESQL_URL_DEV")
	if !ok {
		log.Fatal("Test users: No database url provided. Expected POSTGRESQL_URL_DEV")
	}

	db, err := database.NewDatabaseConnection(dburl)
	if err != nil {
		log.Fatal("Test users: Error connecting to database: ", err)
	}
	defer db.Close()

	ss, err := sessions.NewSessionStore(db, []byte(os.Getenv("SESSION_SECRET")))
	if err != nil {
		log.Fatal("Test users: Error creating session store: ", err)
	}

	userSessionDatas := make([]users.UserSessionData, 3)
	usersList := createUserList()

	for i, user := range usersList {
		createdUser, err := users.CreateUser(db, context.Background(), &user)
		if err != nil {
			createdUser, err = users.GetUserByEmail(db, user.Email)
			if err != nil {
				log.Fatal("Test users: Error getting user: ", err)
			}
		}
		err = setRole(db, createdUser.ID, user_roles.Role(i))
		if err != nil {
			log.Errorf("Test users: Error setting role: ", err)
		}
		userSessionDatas[i] = createdUser.IntoSessionData()
	}

	e := echo.New()
	// e.Logger.SetLevel(log.DEBUG)
	// e.Use(middleware.Logger())

	h1 := &handler{db, ss, userSessionDatas[0]}
	h2 := &handler{db, ss, userSessionDatas[1]}
	h3 := &handler{db, ss, userSessionDatas[2]}
	e.POST("/user", h1.createSession)
	e.POST("/admin", h2.createSession)
	e.POST("/organization-admin", h3.createSession)
	e.POST("/shutdown", func(c echo.Context) error {
		log.Info("Shutting down test server...")
		// shutdown in 1 second
		time.AfterFunc(time.Second, func() {
			os.Exit(0)
		})
		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(":8089"))
}

type handler struct {
	db          *sql.DB
	ss          *pgstore.PGStore
	userSession users.UserSessionData
}

func (h *handler) createSession(c echo.Context) error {
	session, err := h.ss.New(c.Request(), sessions.SESSION_NAME)
	if err != nil {
		return fmt.Errorf("failed to create session: %s", err.Error())
	}
	thisUser := h.userSession

	session.Values[sessions.USER_DATA_VALUE] = thisUser
	session.Options.Domain = "localhost"
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("failed to save session: %s", err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func setRole(db *sql.DB, id uuid.UUID, role user_roles.Role) error {
	_, err := db.Exec(`
	UPDATE users
	SET role = $1
	WHERE id = $2
	`, role.String(), id)
	return err
}

func createUserList() []users.PartialUser {
	return []users.PartialUser{
		{
			SsoID:        "test_user_sso_id",
			Email:        "test1@email.com",
			AccessToken:  "token1",
			TokenExpire:  time.Now().Add(time.Hour * 24),
			RefreshToken: "refresh1",
		},
		{
			SsoID:        "test_admin_sso_id",
			Email:        "test2@email.com",
			AccessToken:  "token2",
			TokenExpire:  time.Now().Add(time.Hour * 24),
			RefreshToken: "refresh2",
		},
		{
			SsoID:        "test_organization_admin_sso_id",
			Email:        "test3@email.com",
			AccessToken:  "token3",
			TokenExpire:  time.Now().Add(time.Hour * 24),
			RefreshToken: "refresh3",
		},
	}
}
