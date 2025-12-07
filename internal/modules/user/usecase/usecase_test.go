package usecase

import (
	"context"
	"errors"
	"testing"

	"go1/internal/modules/user/domain"
	"go1/pkg/apperrors"
	"go1/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	// Initialize logger for tests
	logger.SetLogger(logger.NewZapLogger("development"))
}

// MockUserRepository is a mock for domain.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Fetch(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) DeleteByID(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockUserCache is a mock for domain.UserCache
type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) Get(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserCache) Set(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserCache) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockUserEvent is a mock for domain.UserEvent
type MockUserEvent struct {
	mock.Mock
}

func (m *MockUserEvent) PublishUserCreated(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestUserUsecase_DeleteByID_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockUserCache)
	mockEvent := new(MockUserEvent)

	uc := NewUserUsecase(mockRepo, mockCache, mockEvent)

	userID := int64(1)

	mockRepo.On("DeleteByID", mock.Anything, userID).Return(nil)
	mockCache.On("Delete", mock.Anything, userID).Return(nil)

	err := uc.DeleteByID(context.Background(), userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestUserUsecase_DeleteByID_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockUserCache)
	mockEvent := new(MockUserEvent)

	uc := NewUserUsecase(mockRepo, mockCache, mockEvent)

	userID := int64(999)

	mockRepo.On("DeleteByID", mock.Anything, userID).Return(apperrors.ErrUserNotFound)

	err := uc.DeleteByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Equal(t, apperrors.ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_DeleteByID_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockUserCache)
	mockEvent := new(MockUserEvent)

	uc := NewUserUsecase(mockRepo, mockCache, mockEvent)

	userID := int64(1)
	dbError := errors.New("database error")

	mockRepo.On("DeleteByID", mock.Anything, userID).Return(dbError)

	err := uc.DeleteByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_DeleteByID_CacheDeleteError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockUserCache)
	mockEvent := new(MockUserEvent)

	uc := NewUserUsecase(mockRepo, mockCache, mockEvent)

	userID := int64(1)
	cacheError := errors.New("cache error")

	mockRepo.On("DeleteByID", mock.Anything, userID).Return(nil)
	mockCache.On("Delete", mock.Anything, userID).Return(cacheError)

	err := uc.DeleteByID(context.Background(), userID)

	// Should not fail even if cache delete fails (logged only)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
