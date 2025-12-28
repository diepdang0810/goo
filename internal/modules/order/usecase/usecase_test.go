package usecase

import (
	"context"
	"testing"
	"time"

	"go1/internal/modules/order/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	if args.Error(0) == nil {
		// Simulate DB setting ID
		order.ID = 123
		order.CreatedAt = time.Now()
		order.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func TestOrderUsecase_Create(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewOrderUsecase(mockRepo, nil) // nil temporal client for now

	t.Run("Success Food Order", func(t *testing.T) {
		input := CreateOrderInput{
			UserID:       1,
			Amount:       100,
			Type:         domain.TypeFood,
			RestaurantID: "rest-1",
			Items:        []string{"item-1"},
		}

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *domain.Order) bool {
			return o.UserID == 1 && o.ServiceType == domain.TypeFood && o.Status == domain.StatusCreated
		})).Return(nil)

		output, err := uc.Create(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(123), output.ID)
		assert.Equal(t, domain.StatusCreated, output.Status)
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("ValidationError", func(t *testing.T) {
		// Missing RestaurantID for Food
		input := CreateOrderInput{
			UserID: 1,
			Amount: 100,
			Type:   domain.TypeFood,
		}

		output, err := uc.Create(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "restaurant_id is required")
	})
}
