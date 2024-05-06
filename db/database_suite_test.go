package db_test

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"testing"
	"time"

	"github.com/Molnes/Nyhetsjeger/db/db_populator"
	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
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

type DbTestSuite struct {
	suite.Suite
	psqlContainer *PostgreSQLContainer
	db            *sql.DB
}

func TestDBSuite(t *testing.T) {
	suite.Run(t, new(DbTestSuite))
}

func (s *DbTestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	psqlContainer, err := NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	db, err := database.NewDatabaseConnection(s.psqlContainer.getDBUrl())
	s.Require().NoError(err)
	s.db = db
}

func (s *DbTestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.db.Close())
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
func (s *DbTestSuite) SetupTest() {
	migrator, err := getMigrator(s.psqlContainer.getDBUrl())
	s.Require().NoError(err)
	s.Require().NoError(migrator.Up())

	db_populator.PopulateDbWithTestData(s.db)
}

// run after each test
func (s *DbTestSuite) TearDownTest() {
	migrator, err := getMigrator(s.psqlContainer.getDBUrl())
	s.Require().NoError(err)
	s.Require().NoError(migrator.Down())
}

func (s *DbTestSuite) TestDatabaseConnection() {
	var id uuid.UUID
	err := s.db.QueryRow(`
	SELECT id
	FROM quizzes LIMIT 1;`).Scan(&id)
	s.Require().NoError(err)

	_, err = s.db.Exec(`DELETE FROM quizzes WHERE id=$1`, id)
	s.Require().NoError(err)

	var number int
	err = s.db.QueryRow(`
	SELECT count(*)
	FROM quizzes;`).Scan(&number)
	s.Require().NoError(err)

	s.Require().Equal(1, number)
}

func (s *DbTestSuite) TestDatabaseConnection2() {
	var number int
	err := s.db.QueryRow(`
	SELECT count(*)
	FROM quizzes;`).Scan(&number)
	s.Require().NoError(err)

	s.Require().Equal(2, number)
}
