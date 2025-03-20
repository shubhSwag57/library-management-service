package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"library-management-service/internal/service"
	pb "library-management-service/proto/library/v1"
)

// Mock Repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, name, email, password string) (*pb.User, error) {
	args := m.Called(ctx, name, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.User), args.Error(1)
}

func (m *MockUserRepository) VerifyCredentials(ctx context.Context, email, password string) (*pb.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*pb.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.User), args.Error(1)
}

type MockBookRepository struct {
	mock.Mock
}

func (m *MockBookRepository) Create(ctx context.Context, book *pb.Book) (*pb.Book, error) {
	args := m.Called(ctx, book)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Book), args.Error(1)
}

func (m *MockBookRepository) GetByID(ctx context.Context, id string) (*pb.Book, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Book), args.Error(1)
}

func (m *MockBookRepository) List(ctx context.Context, limit, offset int32) ([]*pb.Book, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pb.Book), args.Error(1)
}

func (m *MockBookRepository) BorrowBook(ctx context.Context, userID, bookID string, dueDate time.Time) (string, error) {
	args := m.Called(ctx, userID, bookID, dueDate)
	return args.String(0), args.Error(1)
}

func (m *MockBookRepository) ReturnBook(ctx context.Context, borrowID string) error {
	args := m.Called(ctx, borrowID)
	return args.Error(0)
}

// Test RegisterUser
func TestLibraryService_RegisterUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockBookRepo := new(MockBookRepository)
		service := service.NewLibraryService(mockUserRepo, mockBookRepo)
		ctx := context.Background()
		req := &pb.RegisterUserRequest{Name: "John Doe", Email: "john@example.com", Password: "password123"}
		expectedUser := &pb.User{Id: "user-id-1", Name: req.Name, Email: req.Email}
		mockUserRepo.On("Create", ctx, req.Name, req.Email, req.Password).Return(expectedUser, nil)
		response, err := service.RegisterUser(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedUser, response.User)
		mockUserRepo.AssertExpectations(t)
	})
}

// Test LoginUser
func TestLibraryService_LoginUser(t *testing.T) {
	t.Run("Invalid Credentials", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockBookRepo := new(MockBookRepository)
		service := service.NewLibraryService(mockUserRepo, mockBookRepo)
		ctx := context.Background()
		req := &pb.LoginUserRequest{Email: "john@example.com", Password: "wrongpassword"}
		mockUserRepo.On("VerifyCredentials", ctx, req.Email, req.Password).Return(nil, errors.New("invalid credentials"))
		response, err := service.LoginUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, response)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})
}

// Test BorrowBook
func TestLibraryService_BorrowBook(t *testing.T) {
	t.Run("Book Borrowed Successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockBookRepo := new(MockBookRepository)
		service := service.NewLibraryService(mockUserRepo, mockBookRepo)
		ctx := context.Background()

		// Simulating the due date calculation
		dueDate := time.Now().AddDate(0, 0, 14)
		expectedDueDate := dueDate.Format(time.RFC3339)

		mockBookRepo.On("BorrowBook", ctx, "user-id-1", "book-id-1", mock.AnythingOfType("time.Time")).Return("borrow-id-123", nil)
		req := &pb.BorrowBookRequest{UserId: "user-id-1", BookId: "book-id-1"}
		response, err := service.BorrowBook(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "borrow-id-123", response.BorrowId)
		assert.Equal(t, expectedDueDate, response.DueDate)

		mockBookRepo.AssertExpectations(t)
	})
}

// Test ReturnBook
func TestLibraryService_ReturnBook(t *testing.T) {
	t.Run("Return Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockBookRepo := new(MockBookRepository)
		service := service.NewLibraryService(mockUserRepo, mockBookRepo)
		ctx := context.Background()
		mockBookRepo.On("ReturnBook", ctx, "borrow-id-123").Return(nil)
		req := &pb.ReturnBookRequest{BorrowId: "borrow-id-123"}
		response, err := service.ReturnBook(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Success)
		mockBookRepo.AssertExpectations(t)
	})
}
