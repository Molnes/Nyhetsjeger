package api

import (
	"fmt"
	"log"
	"os"

	"github.com/Molnes/Nyhetsjeger/internal/bucket"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/Molnes/Nyhetsjeger/internal/models/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/router"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Sets up the web server and starts it.
//
// Tries reading the PORT environment variable, and falls back to 8080 if it's not found.
func Api() {
	e := echo.New()

	var dbUrl string
	if os.Getenv("COMPOSE_PROFILES") == "dev" {
		dbUrl = os.Getenv("POSTGRESQL_URL_DEV")
	} else {
		dbUrl = os.Getenv("POSTGRESQL_URL_PROD")
	}
	if dbUrl == "" {
		log.Fatal("No database url provided. Expected POSTGRESQL_URL_DEV or POSTGRESQL_URL_PROD")
	}

	databaseConn, err := database.NewDatabaseConnection(dbUrl)
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	defer databaseConn.Close()

	sessionKey, ok := os.LookupEnv("SESSION_SECRET")
	if !ok {
		log.Fatal("No session secret provided. Expected SESSION_SECRET")
	}

	sessionStore, err := sessions.NewSessionStore(databaseConn, []byte(sessionKey))
	if err != nil {
		log.Fatal("Error creating session store: ", err)
	}

	key_str, ok := os.LookupEnv("AES_KEY")
	if !ok {
		log.Fatal("No AES key provided. Expected AES_KEY")
	}
	cryptoKey := []byte(key_str)

	googleOauthConfig, err := getGoogleOauthConfig()
	if err != nil {
		log.Fatal("Error getting google oauth config: ", err)
	}

	endpoint, ok := os.LookupEnv("BUCKET_URL")
	accessKeyID, ok := os.LookupEnv("BUCKET_ACCESS_KEY")
	secretAccessKey, ok := os.LookupEnv("BUCKET_SECRET_KEY")
	useSSL, ok := os.LookupEnv("BUCKET_USE_SSL")

	minioClient, err := bucket.NewBucketConnection(endpoint, accessKeyID, secretAccessKey, useSSL == "true")
	if err != nil {
		log.Fatal("Error connecting to bucket: ", err)
	}

	openAIKey, ok := os.LookupEnv("OPENAI_KEY")
	if !ok {
		log.Fatal("No OpenAI key provided. Expected OPENAI_KEY")
	}

	sharedData := &config.SharedData{
		DB:           databaseConn,
		SessionStore: sessionStore,
		CryptoKey:    cryptoKey,
		Bucket:       minioClient,
		OpenAIKey:    openAIKey,
	}

	router.SetupRouter(e, sharedData, googleOauthConfig)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Println("No PORT found, using 8080")
		port = "8080"
	}
	address := fmt.Sprint(":", port)

	e.Logger.Fatal(e.Start(address))
}

func getGoogleOauthConfig() (*oauth2.Config, error) {

	redirectUrl, ok := os.LookupEnv("GOOGLE_REDIRECT_URL")
	if !ok {
		return nil, fmt.Errorf("No redirect url provided. Expected GOOGLE_REDIRECT_URL")
	}

	clientId, ok := os.LookupEnv("GOOGLE_CLIENT_ID")
	if !ok {
		return nil, fmt.Errorf("No client id provided. Expected GOOGLE_CLIENT_ID")
	}

	clientSecret, ok := os.LookupEnv("GOOGLE_CLIENT_SECRET")
	if !ok {
		return nil, fmt.Errorf("No client secret provided. Expected GOOGLE_CLIENT_SECRET")
	}

	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	}, nil
}
