package db_integration_test_suite

// This file constains base test suite that can be used by other modules to perform integration testing.
// Make sure to tag files containing integration tests with `//go:build integration` at the top.

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/Molnes/Nyhetsjeger/db/db_populator"
	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgreSQLContainer struct {
	testcontainers.Container
	MappedPort string
	Host       string
}

const (
	psqlImage  = "postgres"
	psqlPort   = "5432"
	imageTag   = "16.2"
	dbUser     = "user"
	dbPassword = "password"
	dbName     = "db_test"
)

func (c *PostgreSQLContainer) getDBUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, c.Host, c.MappedPort, dbName)
}

// Creates new postgresql testcontainer
// Container setup code in this file is inspired by and partly borrowed from https://dev.to/kliukovkin/integration-tests-with-go-and-testcontainers-6o5
func newPostgreSQLContainer(ctx context.Context) (*PostgreSQLContainer, error) {

	containerPort := psqlPort + "/tcp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"POSTGRES_USER":     dbUser,
				"POSTGRES_PASSWORD": dbPassword,
				"POSTGRES_DB":       dbName,
			},
			ExposedPorts: []string{
				containerPort,
			},
			Image:      fmt.Sprintf("%s:%s", psqlImage, imageTag),
			WaitingFor: wait.ForListeningPort(nat.Port(containerPort)),
			Cmd:        []string{"-c", "log_statement=all"},
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("getting request provider: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting host for: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		return nil, fmt.Errorf("getting mapped port for (%s): %w", containerPort, err)
	}

	return &PostgreSQLContainer{
		Container:  container,
		MappedPort: mappedPort.Port(),
		Host:       host,
	}, nil
}

// Base test suite for integration tests requiring sql database connection
type DbIntegrationTestBaseSuite struct {
	suite.Suite
	psqlContainer *PostgreSQLContainer
	DB            *sql.DB
}

// Runs once at test suite setup.
// Creates test psql container, creates a databse connection and sets a pointer to it as the struct field.
func (s *DbIntegrationTestBaseSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer ctxCancel()

	psqlContainer, err := newPostgreSQLContainer(ctx)
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	db, err := database.NewDatabaseConnection(s.psqlContainer.getDBUrl())
	s.Require().NoError(err)
	s.DB = db
}

// Ran when the suite is done. Closes the DB conenction and terminates the container.
func (s *DbIntegrationTestBaseSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.DB.Close())
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
}

// This is required for the migration files to be available for the migrator utility.
//
//go:embed migrations/*.sql
var fs embed.FS

var migrator *migrate.Migrate

// Creates the migration utility.
func getMigrator(dbUrl string) (*migrate.Migrate, error) {
	if migrator == nil {
		d, err := iofs.New(fs, "migrations")
		if err != nil {
			return nil, err
		}
		migrator, err = migrate.NewWithSourceInstance("iofs", d, dbUrl)
		if err != nil {
			return nil, err
		}
	}
	return migrator, nil
}

// Runs before each test.
// Migrates the database up, runs population (seeding) function.
func (s *DbIntegrationTestBaseSuite) SetupTest() {
	migrator, err := getMigrator(s.psqlContainer.getDBUrl())
	s.Require().NoError(err)
	s.Require().NoError(migrator.Up())

	db_populator.PopulateDbWithTestData(s.DB)
}

// Runs after each test.
// Migrates the DB all the way down, removing all data.
func (s *DbIntegrationTestBaseSuite) TearDownTest() {
	migrator, err := getMigrator(s.psqlContainer.getDBUrl())
	s.Require().NoError(err)
	s.Require().NoError(migrator.Down())
}
