package repository_test

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"library-management-service/internal/database"
	"library-management-service/internal/repository"
)

// MockRow implements a mock for database row
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

// MockPgxPool implements the PgxPool interface for testing
type MockPgxPool struct {
	mock.Mock
}

func (m *MockPgxPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	args := m.Called(ctx)
	return args.Get(0).(*pgxpool.Conn), args.Error(1)
}

func (m *MockPgxPool) Close() {
	m.Called()
}

func (m *MockPgxPool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	callArgs := m.Called(ctx, sql, args)
	return callArgs.Get(0).(pgconn.CommandTag), callArgs.Error(1)
}

func (m *MockPgxPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	return callArgs.Get(0).(pgx.Rows), callArgs.Error(1)
}

func (m *MockPgxPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	callArgs := m.Called(ctx, sql, args)
	return callArgs.Get(0).(pgx.Row)
}

func TestUserRepository_Create(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	name := "Test User"
	email := "test@example.com"
	password := "password123"

	// Expectations - correctly handle variadic arguments as a slice
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		// Simulate filling the user ID, name, and email
		dests := args.Get(0).([]interface{})
		*(dests[0].(*string)) = "user-id-123"
		*(dests[1].(*string)) = name
		*(dests[2].(*string)) = email
	}).Return(nil)

	// Execute
	user, err := repo.Create(ctx, name, email, password)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-id-123", user.Id)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, email, user.Email)

	// Verify that password was hashed - need to access it from the args slice
	calls := mockPool.Calls[0]
	argsSlice := calls.Arguments[2].([]interface{})
	assert.Equal(t, name, argsSlice[0])
	assert.Equal(t, email, argsSlice[1])
	hashedPassword := argsSlice[2].(string)
	assert.NotEqual(t, password, hashedPassword, "Password should be hashed")

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestUserRepository_Create_DatabaseError(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	name := "Test User"
	email := "test@example.com"
	password := "password123"

	// Expectations
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(errors.New("database error"))

	// Execute
	user, err := repo.Create(ctx, name, email, password)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to create user")

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestUserRepository_Create_DatabaseError1(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	name := "Test User"
	email := "test@example.com"
	password := "password123"

	// Expectations - correctly handle variadic arguments as a slice
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(errors.New("database error"))

	// Execute
	user, err := repo.Create(ctx, name, email, password)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to create user")

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestUserRepository_GetByID1(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	userID := "user-id-123"
	name := "Test User"
	email := "test@example.com"

	// Expectations - use mock.Anything for the variadic argument slice
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		// Simulate filling the user data
		dests := args.Get(0).([]interface{})
		*(dests[0].(*string)) = userID
		*(dests[1].(*string)) = name
		*(dests[2].(*string)) = email
	}).Return(nil)

	// Execute
	user, err := repo.GetByID(ctx, userID)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.Id)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, email, user.Email)

	// Verify correct ID was passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, userID, argsSlice[0])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestUserRepository_GetByID_NotFound1(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	userID := "nonexistent-id"

	// Expectations - use mock.Anything for the variadic argument slice
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

	// Execute
	user, err := repo.GetByID(ctx, userID)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")

	// Verify correct ID was passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, userID, argsSlice[0])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestUserRepository_GetByID_DatabaseError(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	userID := "user-id-123"

	// Expectations - use mock.Anything for the variadic argument slice
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(errors.New("database error"))

	// Execute
	user, err := repo.GetByID(ctx, userID)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "database error")

	// Verify correct ID was passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, userID, argsSlice[0])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestUserRepository_VerifyCredentials_Success(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	userID := "user-id-123"
	name := "Test User"
	email := "test@example.com"
	password := "password123"

	// Hash the password as it would be in the database
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Expectations - use mock.Anything for the variadic argument slice
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		// Simulate filling the user data including password hash
		dests := args.Get(0).([]interface{})
		*(dests[0].(*string)) = userID
		*(dests[1].(*string)) = name
		*(dests[2].(*string)) = email
		*(dests[3].(*string)) = string(hashedPassword)
	}).Return(nil)

	// Execute
	user, err := repo.VerifyCredentials(ctx, email, password)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.Id)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, email, user.Email)

	// Verify correct email was passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, email, argsSlice[0])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestUserRepository_VerifyCredentials_UserNotFound(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	email := "nonexistent@example.com"
	password := "password123"

	// Expectations - use mock.Anything for the variadic argument slice
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

	// Execute
	user, err := repo.VerifyCredentials(ctx, email, password)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")

	// Verify correct email was passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, email, argsSlice[0])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestUserRepository_VerifyCredentials_WrongPassword(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	userID := "user-id-123"
	name := "Test User"
	email := "test@example.com"
	correctPassword := "password123"
	wrongPassword := "wrongpassword"

	// Hash the correct password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	// Expectations - use mock.Anything for the variadic argument slice
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		dests := args.Get(0).([]interface{})
		*(dests[0].(*string)) = userID
		*(dests[1].(*string)) = name
		*(dests[2].(*string)) = email
		*(dests[3].(*string)) = string(hashedPassword)
	}).Return(nil)

	// Execute with wrong password
	user, err := repo.VerifyCredentials(ctx, email, wrongPassword)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")

	// Verify correct email was passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, email, argsSlice[0])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestUserRepository_VerifyCredentials_DatabaseError(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test data
	email := "test@example.com"
	password := "password123"

	// Expectations - use mock.Anything for the variadic argument slice
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(errors.New("database error"))

	// Execute
	user, err := repo.VerifyCredentials(ctx, email, password)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "database error")

	// Verify correct email was passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, email, argsSlice[0])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}
