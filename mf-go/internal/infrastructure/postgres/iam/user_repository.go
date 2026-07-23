package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/model"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/database"
)

// UserRepo implements repository.UserRepository with PostgreSQL.
type UserRepo struct {
	db         *pgxpool.Pool
	usersTable string
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *pgxpool.Pool, schema string) *UserRepo {
	return &UserRepo{
		db:         db,
		usersTable: database.QualifyTable(schema, "users"),
	}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		fmt.Sprintf(`INSERT INTO %s (id, email, password_hash, first_name, last_name, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, r.usersTable),
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Status, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainErr.New(domainErr.ErrAlreadyExists, "user with this email already exists", err)
		}
		return domainErr.New(domainErr.ErrInternal, "failed to create user", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		fmt.Sprintf(`SELECT id, email, password_hash, first_name, last_name, status, created_at, updated_at
		 FROM %s WHERE id = $1`, r.usersTable), id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "user not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get user", err)
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		fmt.Sprintf(`SELECT id, email, password_hash, first_name, last_name, status, created_at, updated_at
		 FROM %s WHERE email = $1`, r.usersTable), email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "user not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get user by email", err)
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		fmt.Sprintf(`UPDATE %s SET email=$1, first_name=$2, last_name=$3, status=$4, updated_at=$5 WHERE id=$6`, r.usersTable),
		user.Email, user.FirstName, user.LastName, user.Status, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to update user", err)
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE id=$1`, r.usersTable), id)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to delete user", err)
	}
	return nil
}

func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int, error) {
	var total int
	err := r.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s`, r.usersTable)).Scan(&total)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to count users", err)
	}

	rows, err := r.db.Query(ctx,
		fmt.Sprintf(`SELECT id, email, password_hash, first_name, last_name, status, created_at, updated_at
		 FROM %s ORDER BY created_at DESC LIMIT $1 OFFSET $2`, r.usersTable), limit, offset,
	)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to list users", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to scan user", err)
		}
		users = append(users, &u)
	}
	return users, total, nil
}
