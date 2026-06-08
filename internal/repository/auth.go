package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, username, passwordHash string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByID(ctx context.Context, id string) (*model.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID int64) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, username, passwordHash string) (*model.User, error) {
	query := `
		INSERT INTO users (username, password_hash)
		VALUES (?, ?)
		RETURNING id, username, password_hash, role, created_at
	`
	var user model.User
	var createdAtStr string
	err := r.db.QueryRowContext(ctx, query, username, passwordHash).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&createdAtStr,
	)
	if err != nil {
		return nil, err
	}
	user.CreatedAt, _ = parseDBTime(createdAtStr)
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`
	var user model.User
	var createdAtStr string
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&createdAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	user.CreatedAt, _ = parseDBTime(createdAtStr)
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE id = ?`
	var user model.User
	var createdAtStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&createdAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	user.CreatedAt, _ = parseDBTime(createdAtStr)
	return &user, nil
}

type sessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *model.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?, ?, ?)
		RETURNING created_at
	`
	var createdAtStr string
	err := r.db.QueryRowContext(ctx, query, session.ID, session.UserID, dbTime(session.ExpiresAt)).Scan(&createdAtStr)
	if err != nil {
		return err
	}
	session.CreatedAt, _ = parseDBTime(createdAtStr)
	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id string) (*model.Session, error) {
	query := `SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = ?`
	var session model.Session
	var expiresAtStr, createdAtStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&expiresAtStr,
		&createdAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	session.ExpiresAt, _ = parseDBTime(expiresAtStr)
	session.CreatedAt, _ = parseDBTime(createdAtStr)
	return &session, nil
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}
