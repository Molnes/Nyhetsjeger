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
		userSessionDatas[i] = user.IntoSessionData()
		_, err := users.CreateUser(db, &user, context.Background())
		if err != nil {
			log.Errorf("Test users: Error inserting user: ", err)
		}
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

func createUserList() []users.User {
	uuid1, _ := uuid.Parse("7511ce90-87f1-4cef-8980-6b9aef71529c")
	uuid2, _ := uuid.Parse("6d0d8550-703e-41c8-aed9-959d80752a4b")
	uuid3, _ := uuid.Parse("c0f84533-7613-4fe2-90a5-c452a1e57121")

	testUsers := []users.User{
		{
			ID:                 uuid1,
			SsoID:              "test_user_sso_id",
			Username:           "test user",
			Email:              "test1@email.com",
			Phone:              "00000001",
			OptInRanking:       true,
			Role:               user_roles.User,
			AccessTokenCypher:  []byte("token1"),
			Token_expire:       time.Now().Add(time.Hour * 24),
			RefreshtokenCypher: []byte("refresh1"),
		},
		{
			ID:                 uuid2,
			SsoID:              "test_admin_sso_id",
			Username:           "test admin",
			Email:              "test2@email.com",
			Phone:              "00000002",
			OptInRanking:       false,
			Role:               user_roles.QuizAdmin,
			AccessTokenCypher:  []byte("token2"),
			Token_expire:       time.Now().Add(time.Hour * 24),
			RefreshtokenCypher: []byte("refresh2"),
		},
		{
			ID:                 uuid3,
			SsoID:              "test_organization_admin_sso_id",
			Username:           "test organization admin",
			Email:              "test3@email.com",
			Phone:              "00000003",
			OptInRanking:       false,
			Role:               user_roles.OrganizationAdmin,
			AccessTokenCypher:  []byte("token3"),
			Token_expire:       time.Now().Add(time.Hour * 24),
			RefreshtokenCypher: []byte("refresh3"),
		},
	}

	return testUsers
}
