package db_integration_test_suite

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

// code in this file is inspired by and partly borrowed from https://dev.to/kliukovkin/integration-tests-with-go-and-testcontainers-6o5

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

func NewPostgreSQLContainer(ctx context.Context) (*PostgreSQLContainer, error) {

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

type DbIntegrationTestBaseSuite struct {
	suite.Suite
	psqlContainer *PostgreSQLContainer
	DB            *sql.DB
}

func (s *DbIntegrationTestBaseSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	psqlContainer, err := NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	db, err := database.NewDatabaseConnection(s.psqlContainer.getDBUrl())
	s.Require().NoError(err)
	s.DB = db
}

func (s *DbIntegrationTestBaseSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.DB.Close())
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
}

//go:embed migrations/*.sql
var fs embed.FS

func getMigrator(dbUrl string) (*migrate.Migrate, error) {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, dbUrl)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// run before each test
func (s *DbIntegrationTestBaseSuite) SetupTest() {
	migrator, err := getMigrator(s.psqlContainer.getDBUrl())
	s.Require().NoError(err)
	s.Require().NoError(migrator.Up())

	db_populator.PopulateDbWithTestData(s.DB)
}

// run after each test
func (s *DbIntegrationTestBaseSuite) TearDownTest() {
	migrator, err := getMigrator(s.psqlContainer.getDBUrl())
	s.Require().NoError(err)
	s.Require().NoError(migrator.Down())
}
