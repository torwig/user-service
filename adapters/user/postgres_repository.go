package user

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/torwig/user-service/entities"
)

const (
	maxConnLifetime       = time.Hour
	maxConnLifetimeJitter = time.Minute
	maxConnIdleTime       = 30 * time.Minute
	maxPoolConns          = 10
	minPoolConns          = 2
	connHealthCheckPeriod = time.Minute
	databaseConnTimeout   = 5 * time.Second
	userTableName         = "users"
)

type Config struct {
	DSN string
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(repoCfg Config) (*PostgresRepository, error) {
	cfg, err := pgxpool.ParseConfig(repoCfg.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse connection string")
	}

	cfg.MaxConnLifetime = maxConnLifetime
	cfg.MaxConnLifetimeJitter = maxConnLifetimeJitter
	cfg.MaxConnIdleTime = maxConnIdleTime
	cfg.MaxConns = maxPoolConns
	cfg.MinConns = minPoolConns
	cfg.HealthCheckPeriod = connHealthCheckPeriod

	ctx, cancel := context.WithTimeout(context.Background(), databaseConnTimeout)
	defer cancel()

	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create connection pool")
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping the database")
	}

	return &PostgresRepository{db: conn}, nil
}

func (r *PostgresRepository) Close() {
	r.db.Close()
}

func (r *PostgresRepository) Create(ctx context.Context, params entities.CreateUserParams) (entities.User, error) {
	var user entities.User

	stmt := sq.
		Insert(userTableName).
		Columns("first_name", "last_name", "phone_number", "address").
		Values(params.FirstName, params.LastName, params.PhoneNumber, params.Address).
		Suffix("RETURNING id, first_name, last_name, phone_number, address, deleted, created_at, deleted_at").
		PlaceholderFormat(sq.Dollar)

	sql, args, err := stmt.ToSql()
	if err != nil {
		return user, errors.Wrap(err, "failed to build a query")
	}

	err = pgxscan.Get(ctx, r.db, &user, sql, args...)
	if err != nil {
		return user, errors.Wrap(err, "failed to execute a query")
	}

	return user, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id int64) (entities.User, error) {
	var user entities.User

	stmt := sq.
		Select("id", "first_name", "last_name", "phone_number", "address", "deleted", "created_at", "deleted_at").
		From(userTableName).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := stmt.ToSql()
	if err != nil {
		return user, errors.Wrap(err, "failed to build a query")
	}

	err = pgxscan.Get(ctx, r.db, &user, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, entities.ErrUserNotFound
		}

		return user, errors.Wrap(err, "failed to execute a query")
	}

	return user, nil
}

func (r *PostgresRepository) Update(
	ctx context.Context,
	id int64,
	params entities.UpdateUserParams,
) (entities.User, error) {
	nothingToUpdate := params == entities.UpdateUserParams{}
	if nothingToUpdate {
		return r.Get(ctx, id)
	}

	var user entities.User

	stmt := sq.
		Update(userTableName)
	if params.FirstName != nil {
		stmt = stmt.Set("first_name", *params.FirstName)
	}
	if params.LastName != nil {
		stmt = stmt.Set("last_name", *params.LastName)
	}
	if params.PhoneNumber != nil {
		stmt = stmt.Set("phone_number", *params.PhoneNumber)
	}
	if params.Address != nil {
		stmt = stmt.Set("address", *params.Address)
	}

	stmt = stmt.
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, first_name, last_name, phone_number, address, deleted, created_at, deleted_at").
		PlaceholderFormat(sq.Dollar)

	sql, args, err := stmt.ToSql()
	if err != nil {
		return user, errors.Wrap(err, "failed to build a query")
	}

	err = pgxscan.Get(ctx, r.db, &user, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, entities.ErrUserNotFound
		}

		return user, errors.Wrap(err, "failed to execute a query")
	}

	return user, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
	stmt := sq.
		Update(userTableName).
		Set("deleted", true).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := stmt.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build a query")
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return errors.Wrap(err, "failed to execute a query")
	}

	return nil
}
