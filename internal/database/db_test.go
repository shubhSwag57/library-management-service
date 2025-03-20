package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDB mocks the database interactions
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetExchangeRate(ctx context.Context, baseCurrency, targetCurrency string) (float64, error) {
	args := m.Called(ctx, baseCurrency, targetCurrency)
	return args.Get(0).(float64), args.Error(1)
}

func TestGetExchangeRate_Success(t *testing.T) {
	mockDB := new(MockDB)
	ctx := context.Background()

	// Mock the expected response
	mockDB.On("GetExchangeRate", ctx, "USD", "EUR").Return(0.85, nil)

	rate, err := mockDB.GetExchangeRate(ctx, "USD", "EUR")

	assert.NoError(t, err)
	assert.Equal(t, 0.85, rate)

	mockDB.AssertExpectations(t)
}

func TestGetExchangeRate_Error(t *testing.T) {
	mockDB := new(MockDB)
	ctx := context.Background()

	// Mock the expected error response
	mockDB.On("GetExchangeRate", ctx, "USD", "XYZ").Return(0.0, errors.New("currency not found"))

	rate, err := mockDB.GetExchangeRate(ctx, "USD", "XYZ")

	assert.Error(t, err)
	assert.Equal(t, 0.0, rate)
	assert.Equal(t, "currency not found", err.Error())

	mockDB.AssertExpectations(t)
}
