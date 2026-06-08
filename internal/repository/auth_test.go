package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azzt17/finance-tracker/internal/model"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

func TestUserRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test Create
	user, err := repo.Create(ctx, "testuser", "hashedpassword")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	if user.ID == 0 {
		t.Errorf("expected non-zero ID")
	}
	if user.Username != "testuser" {
		t.Errorf("expected username testuser, got %v", user.Username)
	}
	if user.Role != model.RoleUser {
		t.Errorf("expected default role user, got %v", user.Role)
	}

	// Test GetByUsername
	foundUser, err := repo.GetByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("failed to get user by username: %v", err)
	}
	if foundUser.ID != user.ID {
		t.Errorf("expected ID %v, got %v", user.ID, foundUser.ID)
	}

	// Test GetByUsername NotFound
	_, err = repo.GetByUsername(ctx, "nonexistent")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Test GetByID
	foundUserByID, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to get user by ID: %v", err)
	}
	if foundUserByID.Username != "testuser" {
		t.Errorf("expected username testuser, got %v", foundUserByID.Username)
	}
}

func TestSessionRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "sessionuser", "hash")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	session := &model.Session{
		ID:        "session-123",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Test Create
	err = sessionRepo.Create(ctx, session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Test GetByID
	foundSession, err := sessionRepo.GetByID(ctx, "session-123")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if foundSession.UserID != user.ID {
		t.Errorf("expected userID %v, got %v", user.ID, foundSession.UserID)
	}

	// Test Delete
	err = sessionRepo.Delete(ctx, "session-123")
	if err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}

	_, err = sessionRepo.GetByID(ctx, "session-123")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}

	// Test DeleteByUserID
	session2 := &model.Session{
		ID:        "session-456",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	_ = sessionRepo.Create(ctx, session2)
	err = sessionRepo.DeleteByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to delete by user ID: %v", err)
	}
	_, err = sessionRepo.GetByID(ctx, "session-456")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound after DeleteByUserID, got %v", err)
	}
}
