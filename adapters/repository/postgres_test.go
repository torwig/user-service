package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/torwig/user-service/adapters/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torwig/user-service/entities"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pkg/errors"
)

const (
	postgresContainerExpirationSeconds = 180
	postgresUsername                   = "users"
	postgresPassword                   = "users"
	postgresDatabaseName               = "users"
	migrationsLocation                 = "../../migrations"
)

var repo *repository.PostgresRepository

func TestMain(m *testing.M) {
	pool, resource := createPostgresContainer()

	hostAndPort := resource.GetHostPort("5432/tcp")
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		postgresUsername, postgresPassword, hostAndPort, postgresDatabaseName)

	err := prepareDatabaseConnection(pool, dsn)
	if err != nil {
		purgeResource(pool, resource)

		log.Fatalf("Failed to connect to Postgres: %s", err)
	}

	code := m.Run()

	purgeResource(pool, resource)

	os.Exit(code)
}

func createPostgresContainer() (*dockertest.Pool, *dockertest.Resource) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Failed to construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		Env: []string{
			"POSTGRES_PASSWORD=" + postgresPassword,
			"POSTGRES_USER=" + postgresUsername,
			"POSTGRES_DB=" + postgresDatabaseName,
			"POSTGRES_HOST_AUTH_METHOD=trust",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Failed to start container with Postgres: %s", err)
	}

	_ = resource.Expire(postgresContainerExpirationSeconds)

	pool.MaxWait = postgresContainerExpirationSeconds * time.Second

	return pool, resource
}

func prepareDatabaseConnection(pool *dockertest.Pool, dsn string) error {
	log.Printf("Connecting to Postgres on: %s", dsn)

	if err := pool.Retry(func() error {
		conn, err := sql.Open("pgx", dsn)
		if err != nil {
			return errors.Wrap(err, "failed to create a database connection")
		}

		err = applyMigrations(conn, migrationsLocation)
		if err != nil {
			return errors.Wrap(err, "failed to apply migrations")
		}

		_ = conn.Close()

		r, err := repository.NewPostgresRepository(repository.Config{DSN: dsn})
		if err != nil {
			return errors.Wrap(err, "failed to create user repository")
		}

		repo = r

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to connect to the database inside the container")
	}

	return nil
}

func applyMigrations(sqlDB *sql.DB, location string) error {
	driver, err := migratePostgres.WithInstance(sqlDB, &migratePostgres.Config{})
	if err != nil {
		return errors.Wrap(err, "failed to create database driver for the migration process")
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+location,
		postgresDatabaseName, driver)
	if err != nil {
		return errors.Wrap(err, "failed to init migrations")
	}

	err = m.Up()
	if err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}

	return nil
}

func purgeResource(pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		log.Printf("Failed to purge resource: %s", err)
	}
}

func TestPostgresRepository_Create(t *testing.T) {
	userParams := entities.CreateUserParams{
		FirstName:   "John",
		LastName:    "Wick",
		PhoneNumber: "+1234567890",
		Address:     "New York, 123 Lincoln Square",
	}

	createdUser, err := repo.Create(context.Background(), userParams)
	require.NoError(t, err)
	assert.Greater(t, createdUser.ID, int64(0))

	assert.False(t, createdUser.IsDeleted())
	assert.Equal(t, userParams.FirstName, createdUser.FirstName)
	assert.Equal(t, userParams.LastName, createdUser.LastName)
	assert.Equal(t, userParams.PhoneNumber, createdUser.PhoneNumber)
	assert.Equal(t, userParams.Address, createdUser.Address)
}

func TestPostgresRepository_Get(t *testing.T) {
	t.Run("Get non-existing user", func(t *testing.T) {
		_, err := repo.Get(context.Background(), 999_999_999)
		require.ErrorIs(t, err, entities.ErrUserNotFound)
	})

	t.Run("Get existing user", func(t *testing.T) {
		userParams := entities.CreateUserParams{
			FirstName:   "Mike",
			LastName:    "Brown",
			PhoneNumber: "+0987654321",
			Address:     "Los Angeles, 17 Beach Road",
		}

		createdUser, err := repo.Create(context.Background(), userParams)
		require.NoError(t, err)

		_, err = repo.Get(context.Background(), createdUser.ID)
		require.NoError(t, err)
		assert.False(t, createdUser.IsDeleted())
		assert.Equal(t, userParams.FirstName, createdUser.FirstName)
		assert.Equal(t, userParams.LastName, createdUser.LastName)
		assert.Equal(t, userParams.PhoneNumber, createdUser.PhoneNumber)
		assert.Equal(t, userParams.Address, createdUser.Address)
	})
}

func TestPostgresRepository_Update(t *testing.T) {
	t.Run("Update non-existing user", func(t *testing.T) {
		updateParams := entities.UpdateUserParams{
			Address: stringPtr("Some address"),
		}

		_, err := repo.Update(context.Background(), 999_999_999, updateParams)
		require.ErrorIs(t, err, entities.ErrUserNotFound)
	})

	t.Run("Update with no fields set", func(t *testing.T) {
		userParams := entities.CreateUserParams{
			FirstName:   "Amy",
			LastName:    "Pink",
			PhoneNumber: "+4785692130",
			Address:     "Seattle, 555 Park Square",
		}

		createdUser, err := repo.Create(context.Background(), userParams)
		require.NoError(t, err)

		updatedUser, err := repo.Update(context.Background(), createdUser.ID, entities.UpdateUserParams{})
		require.NoError(t, err)
		assert.Equal(t, userParams.FirstName, updatedUser.FirstName)
		assert.Equal(t, userParams.LastName, updatedUser.LastName)
		assert.Equal(t, userParams.PhoneNumber, updatedUser.PhoneNumber)
		assert.Equal(t, userParams.Address, updatedUser.Address)
	})

	t.Run("Update existing user", func(t *testing.T) {
		userParams := entities.CreateUserParams{
			FirstName:   "Jackie",
			LastName:    "Black",
			PhoneNumber: "+123987456",
			Address:     "Washington, 321 Central Street",
		}

		createdUser, err := repo.Create(context.Background(), userParams)
		require.NoError(t, err)

		updateParams := entities.UpdateUserParams{
			PhoneNumber: stringPtr("Different phone"),
			Address:     stringPtr("Different address"),
		}

		updatedUser, err := repo.Update(context.Background(), createdUser.ID, updateParams)
		require.NoError(t, err)
		assert.Equal(t, userParams.FirstName, updatedUser.FirstName)
		assert.Equal(t, userParams.LastName, updatedUser.LastName)
		assert.Equal(t, *updateParams.PhoneNumber, updatedUser.PhoneNumber)
		assert.Equal(t, *updateParams.Address, updatedUser.Address)
	})
}

func TestPostgresRepository_Delete(t *testing.T) {
	t.Run("Delete non-existing user", func(t *testing.T) {
		err := repo.Delete(context.Background(), 999_999_999)
		require.NoError(t, err)
	})

	t.Run("Delete existing user", func(t *testing.T) {
		userParams := entities.CreateUserParams{
			FirstName:   "Robert",
			LastName:    "Speed",
			PhoneNumber: "+3621478950",
			Address:     "Miami, 777 Star Avenue",
		}

		createdUser, err := repo.Create(context.Background(), userParams)
		require.NoError(t, err)

		err = repo.Delete(context.Background(), createdUser.ID)
		require.NoError(t, err)

		deletedUser, err := repo.Get(context.Background(), createdUser.ID)
		require.NoError(t, err)
		assert.True(t, deletedUser.IsDeleted())

		// deleting already deleted user returns no error
		err = repo.Delete(context.Background(), createdUser.ID)
		require.NoError(t, err)
	})
}

func stringPtr(s string) *string {
	return &s
}
