package repository

import (
	"context"
	"errors"
	"testing"

	"go1/pkg/apperrors"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestPostgresUserRepository_DeleteByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewPostgresUserRepositoryWithInterface(mock)

	userID := int64(1)

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteByID(context.Background(), userID)

	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresUserRepository_DeleteByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewPostgresUserRepositoryWithInterface(mock)

	userID := int64(999)

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err = repo.DeleteByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Equal(t, apperrors.ErrUserNotFound, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresUserRepository_DeleteByID_DatabaseError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewPostgresUserRepositoryWithInterface(mock)

	userID := int64(1)
	dbError := errors.New("database connection error")

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnError(dbError)

	err = repo.DeleteByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "postgresUserRepository.DeleteByID")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
