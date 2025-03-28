package service_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"

	"library-management-service/internal/mocks"
	"library-management-service/internal/service"
	pb "library-management-service/proto/library/v1"
)

// TestLibraryService_CheckBookAvailability tests the CheckBookAvailability method

// Test CheckBookAvailability method with mocks
func TestLibraryService_CheckBookAvailability(t *testing.T) {
	t.Run("Available Book", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service with mock repositories
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		bookID := "book-id-123"
		req := &pb.CheckBookAvailabilityRequest{
			BookId: bookID,
		}

		// Set up mock expectation
		expectedBook := &pb.Book{
			Id:        bookID,
			Title:     "Test Book",
			Author:    "Test Author",
			Isbn:      "1234567890",
			Available: true,
		}
		mockBookRepo.On("GetByID", ctx, bookID).Return(expectedBook, nil)

		// Execute
		response, err := svc.CheckBookAvailability(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Available)
		assert.Equal(t, "Available", response.Status)

		// Verify mock was called as expected
		mockBookRepo.AssertExpectations(t)
	})

	t.Run("Borrowed Book", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service with mock repositories
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		bookID := "book-id-456"
		req := &pb.CheckBookAvailabilityRequest{
			BookId: bookID,
		}

		// Set up mock expectation
		borrowedBook := &pb.Book{
			Id:        bookID,
			Title:     "Borrowed Book",
			Author:    "Some Author",
			Isbn:      "0987654321",
			Available: false,
		}
		mockBookRepo.On("GetByID", ctx, bookID).Return(borrowedBook, nil)

		// Execute
		response, err := svc.CheckBookAvailability(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.False(t, response.Available)
		assert.Equal(t, "Borrowed", response.Status)

		// Verify mock was called as expected
		mockBookRepo.AssertExpectations(t)
	})

	t.Run("Book Not Found", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service with mock repositories
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		bookID := "nonexistent-id"
		req := &pb.CheckBookAvailabilityRequest{
			BookId: bookID,
		}

		// Set up mock expectation
		mockBookRepo.On("GetByID", ctx, bookID).Return(nil, errors.New("book not found"))

		// Execute
		response, err := svc.CheckBookAvailability(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.Contains(t, st.Message(), "book not found")
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())

		// Verify mock was called as expected
		mockBookRepo.AssertExpectations(t)
	})
}

// Test RegisterUser with mocks
func TestLibraryService_RegisterUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		password := "password123"

		req := &pb.RegisterUserRequest{
			Name:     name,
			Email:    email,
			Password: password,
		}

		// Set up mock expectation
		expectedUser := &pb.User{
			Id:    "user-id-123",
			Name:  name,
			Email: email,
		}
		mockUserRepo.On("Create", ctx, name, email, password).Return(expectedUser, nil)

		// Execute
		response, err := svc.RegisterUser(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedUser.Id, response.User.Id)
		assert.Equal(t, expectedUser.Name, response.User.Name)
		assert.Equal(t, expectedUser.Email, response.User.Email)

		// Verify mock was called as expected
		mockUserRepo.AssertExpectations(t)
	})
}

func TestLibraryService_LoginUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		email := "john@example.com"
		password := "password123"

		req := &pb.LoginUserRequest{
			Email:    email,
			Password: password,
		}

		// Set up mock expectation
		expectedUser := &pb.User{
			Id:    "user-id-123",
			Name:  "John Doe",
			Email: email,
		}
		mockUserRepo.On("VerifyCredentials", ctx, email, password).Return(expectedUser, nil)

		// Execute
		response, err := svc.LoginUser(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedUser.Id, response.User.Id)
		assert.Equal(t, expectedUser.Name, response.User.Name)
		assert.Equal(t, expectedUser.Email, response.User.Email)
		assert.NotEmpty(t, response.Token)

		// Verify mock was called as expected
		//mockUserRepo.AssertExpectations(
		// Verify mock was called as expected
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		email := "john@example.com"
		password := "wrongpassword"

		req := &pb.LoginUserRequest{
			Email:    email,
			Password: password,
		}

		// Set up mock expectation for failed authentication
		mockUserRepo.On("VerifyCredentials", ctx, email, password).Return(nil, errors.New("invalid credentials"))

		// Execute
		response, err := svc.LoginUser(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Contains(t, st.Message(), "invalid credentials")

		// Verify mock was called as expected
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Missing Fields", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test with missing email
		ctx := context.Background()
		req := &pb.LoginUserRequest{
			Email:    "",
			Password: "password123",
		}

		// Execute
		response, err := svc.LoginUser(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "email and password are required")

		// Test with missing password
		req = &pb.LoginUserRequest{
			Email:    "john@example.com",
			Password: "",
		}

		// Execute
		response, err = svc.LoginUser(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok = status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "email and password are required")
	})
}

// Test BorrowBook with mocks
func TestLibraryService_BorrowBook(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		userID := "user-id-123"
		bookID := "book-id-456"

		req := &pb.BorrowBookRequest{
			UserId: userID,
			BookId: bookID,
		}

		// We need to capture the due date to verify it later
		var capturedDueDate time.Time

		// Set up mock expectation
		mockBookRepo.On("BorrowBook", ctx, userID, bookID, mock.AnythingOfType("time.Time")).
			Run(func(args mock.Arguments) {
				capturedDueDate = args.Get(3).(time.Time)
			}).
			Return("borrow-id-789", nil)

		// Execute
		response, err := svc.BorrowBook(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "borrow-id-789", response.BorrowId)

		// Check that the due date is roughly 14 days from now (within a minute tolerance)
		expectedDueDate := time.Now().AddDate(0, 0, 14)
		dueTimeDiff := expectedDueDate.Sub(capturedDueDate)
		assert.Less(t, dueTimeDiff.Abs(), time.Minute)

		// Verify the due date format in the response
		assert.Equal(t, capturedDueDate.Format(time.RFC3339), response.DueDate)

		// Verify mock was called as expected
		mockBookRepo.AssertExpectations(t)
	})

	t.Run("Missing Fields", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test with missing user ID
		ctx := context.Background()
		req := &pb.BorrowBookRequest{
			UserId: "",
			BookId: "book-id-456",
		}

		// Execute
		response, err := svc.BorrowBook(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "user id and book id are required")

		// Test with missing book ID
		req = &pb.BorrowBookRequest{
			UserId: "user-id-123",
			BookId: "",
		}

		// Execute
		response, err = svc.BorrowBook(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok = status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "user id and book id are required")
	})

	t.Run("Book Borrowing Failed", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		userID := "user-id-123"
		bookID := "book-id-456"

		req := &pb.BorrowBookRequest{
			UserId: userID,
			BookId: bookID,
		}

		// Set up mock expectation for failure
		mockBookRepo.On("BorrowBook", ctx, userID, bookID, mock.AnythingOfType("time.Time")).
			Return("", errors.New("book is not available"))

		// Execute
		response, err := svc.BorrowBook(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to borrow book")

		// Verify mock was called as expected
		mockBookRepo.AssertExpectations(t)
	})
}

// Test ReturnBook with mocks
func TestLibraryService_ReturnBook(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		borrowID := "borrow-id-789"

		req := &pb.ReturnBookRequest{
			BorrowId: borrowID,
		}

		// Set up mock expectation
		mockBookRepo.On("ReturnBook", ctx, borrowID).Return(nil)

		// Execute
		response, err := svc.ReturnBook(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Success)

		// Verify mock was called as expected
		mockBookRepo.AssertExpectations(t)
	})

	t.Run("Missing BorrowID", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		req := &pb.ReturnBookRequest{
			BorrowId: "",
		}

		// Execute
		response, err := svc.ReturnBook(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "borrow id is required")
	})

	t.Run("Return Failed", func(t *testing.T) {
		// Create mock repositories
		mockUserRepo := new(mocks.MockUserRepository)
		mockBookRepo := new(mocks.MockBookRepository)

		// Create service
		svc := service.NewLibraryService(mockUserRepo, mockBookRepo)

		// Test data
		ctx := context.Background()
		borrowID := "borrow-id-789"

		req := &pb.ReturnBookRequest{
			BorrowId: borrowID,
		}

		// Set up mock expectation for failure
		mockBookRepo.On("ReturnBook", ctx, borrowID).Return(errors.New("borrow record not found"))

		// Execute
		response, err := svc.ReturnBook(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to return book")

		// Verify mock was called as expected
		mockBookRepo.AssertExpectations(t)
	})
}
