package user

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go1/internal/modules/user/usecase"
	"go1/pkg/apperrors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserUsecase is a mock for usecase.UserUsecaseInterface
type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) Create(ctx context.Context, input usecase.CreateUserInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *MockUserUsecase) GetByID(ctx context.Context, id int64) (*usecase.UserOutput, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.UserOutput), args.Error(1)
}

func (m *MockUserUsecase) Fetch(ctx context.Context) ([]usecase.UserOutput, error) {
	args := m.Called(ctx)
	return args.Get(0).([]usecase.UserOutput), args.Error(1)
}

func (m *MockUserUsecase) DeleteByID(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

func TestUserHandler_DeleteByID_Success(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	handler := NewUserHandler(mockUsecase)

	router := setupTestRouter()
	router.DELETE("/users/:id", handler.DeleteByID)

	mockUsecase.On("DeleteByID", mock.Anything, int64(1)).Return(nil)

	req, _ := http.NewRequest("DELETE", "/users/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User deleted successfully")
	mockUsecase.AssertExpectations(t)
}

func TestUserHandler_DeleteByID_InvalidID(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	handler := NewUserHandler(mockUsecase)

	router := setupTestRouter()
	router.DELETE("/users/:id", handler.DeleteByID)

	req, _ := http.NewRequest("DELETE", "/users/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid ID")
}

func TestUserHandler_DeleteByID_UserNotFound(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	handler := NewUserHandler(mockUsecase)

	router := setupTestRouter()
	router.DELETE("/users/:id", handler.DeleteByID)

	mockUsecase.On("DeleteByID", mock.Anything, int64(999)).Return(apperrors.ErrUserNotFound)

	req, _ := http.NewRequest("DELETE", "/users/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUsecase.AssertExpectations(t)
}

func TestUserHandler_DeleteByID_InternalError(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	handler := NewUserHandler(mockUsecase)

	router := setupTestRouter()
	router.DELETE("/users/:id", handler.DeleteByID)

	mockUsecase.On("DeleteByID", mock.Anything, int64(1)).Return(errors.New("internal error"))

	req, _ := http.NewRequest("DELETE", "/users/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUsecase.AssertExpectations(t)
}
