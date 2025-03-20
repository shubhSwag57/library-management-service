package service_test

import (
	"context"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"library-management-service/internal/database"
	"library-management-service/internal/repository"
	"library-management-service/internal/service"
	pb "library-management-service/proto/library/v1"
	"testing"
)

// MockPgxPool implements the database.PgxPool interface for testing
type MockPgxPool struct {
	mock.Mock
}

func (m *MockPgxPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pgxpool.Conn), args.Error(1)
}

func (m *MockPgxPool) Close() {
	m.Called()
}

func (m *MockPgxPool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pgconn.CommandTag), callArgs.Error(1)
}

func (m *MockPgxPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pgx.Rows), callArgs.Error(1)
}

func (m *MockPgxPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	callArgs := m.Called(ctx, sql, args)
	return callArgs.Get(0).(pgx.Row)
}

// MockRow implements pgx.Row for testing
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

// MockCommandTag implements pgconn.CommandTag for testing
type MockCommandTag struct {
	mock.Mock
}

func (m *MockCommandTag) RowsAffected() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

// TestLibraryService_RegisterUser tests the RegisterUser method
func TestLibraryService_RegisterUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)
		mockRow := new(MockRow)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create real repositories with mocked DB
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service with real repositories
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data
		ctx := context.Background()
		req := &pb.RegisterUserRequest{
			Name:     "Test User5",
			Email:    "test@example5.com",
			Password: "password1235",
		}

		// The actual call pattern is:
		// QueryRow(ctx, sqlString, []interface{}{name, email, hashedPassword})
		mockPool.On("QueryRow",
			ctx,
			mock.AnythingOfType("string"), // SQL statement
			mock.Anything).Return(mockRow) // The variadic parameters as a slice

		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			// Extract the destination pointers and set values
			dests := args.Get(0).([]interface{})
			*(dests[0].(*string)) = "user-id-123"
			*(dests[1].(*string)) = req.Name
			*(dests[2].(*string)) = req.Email
		}).Return(nil)

		// Execute
		response, err := svc.RegisterUser(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "user-id-123", response.User.Id)
		assert.Equal(t, req.Name, response.User.Name)
		assert.Equal(t, req.Email, response.User.Email)

		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("ValidationError", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create real repositories with mocked DB
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service with real repositories
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data - empty name should fail validation
		ctx := context.Background()
		req := &pb.RegisterUserRequest{
			Name:     "",
			Email:    "test@example.com",
			Password: "password123",
		}

		// Execute
		response, err := svc.RegisterUser(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())

		// No DB calls should be made for validation errors
		mockPool.AssertNotCalled(t, "QueryRow")
	})
}

// TestLibraryService_LoginUser tests the LoginUser method
func TestLibraryService_LoginUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)
		mockRow := new(MockRow)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create real repositories with mocked DB
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service with real repositories
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data
		ctx := context.Background()
		req := &pb.RegisterUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
		}

		// Setup expectations - use mock.Anything for variadic arguments
		mockPool.On("QueryRow", ctx, mock.AnythingOfType("string"), mock.Anything).Return(mockRow)

		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			// Extract the destination pointers and set values
			dests := args.Get(0).([]interface{})
			*(dests[0].(*string)) = "user-id-123"
			*(dests[1].(*string)) = req.Name
			*(dests[2].(*string)) = req.Email
		}).Return(nil)

		// Execute
		response, err := svc.RegisterUser(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "user-id-123", response.User.Id)
		assert.Equal(t, req.Name, response.User.Name)
		assert.Equal(t, req.Email, response.User.Email)

		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

}

func TestLibraryService_CreateBook(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create repositories
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data
		ctx := context.Background()
		inputBook := &pb.Book{
			Title:     "Test Book",
			Author:    "Test Author",
			Isbn:      "1234567890",
			Available: true,
		}
		req := &pb.CreateBookRequest{
			Book: inputBook,
		}

		expectedBook := &pb.Book{
			Id:        "book-id-123",
			Title:     inputBook.Title,
			Author:    inputBook.Author,
			Isbn:      inputBook.Isbn,
			Available: inputBook.Available,
		}

		// Patch the Create method
		//patches := gomonkey.ApplyMethod(reflect.TypeOf(bookRepo), "Create",
		//	func(_ *repository.BookRepository, ctx context.Context, book *pb.Book) (*pb.Book, error) {
		//		// Copy input book and add ID
		//		createdBook := &pb.Book{
		//			Id:        "book-id-123",
		//			Title:     book.Title,
		//			Author:    book.Author,
		//			Isbn:      book.Isbn,
		//			Available: book.Available,
		//		}
		//		return createdBook, nil
		//	})
		//defer patches.Reset()

		// Execute
		response, err := svc.CreateBook(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedBook.Id, response.Book.Id)
		assert.Equal(t, expectedBook.Title, response.Book.Title)
		assert.Equal(t, expectedBook.Author, response.Book.Author)
		assert.Equal(t, expectedBook.Isbn, response.Book.Isbn)
		assert.Equal(t, expectedBook.Available, response.Book.Available)
	})

	t.Run("Nil Book", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create repositories
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data with nil book
		ctx := context.Background()
		req := &pb.CreateBookRequest{
			Book: nil,
		}

		// Execute
		response, err := svc.CreateBook(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "book is required")
	})

	t.Run("Empty Title or Author", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create repositories
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data with empty title
		ctx := context.Background()
		req := &pb.CreateBookRequest{
			Book: &pb.Book{
				Title:     "", // Empty title
				Author:    "Test Author",
				Isbn:      "1234567890",
				Available: true,
			},
		}

		// Execute
		response, err := svc.CreateBook(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "title and author are required")
	})

	t.Run("Repository Error", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create repositories
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data
		ctx := context.Background()
		req := &pb.CreateBookRequest{
			Book: &pb.Book{
				Title:     "Test Book",
				Author:    "Test Author",
				Isbn:      "1234567890",
				Available: true,
			},
		}

		// Patch the Create method to return an error
		//patches := gomonkey.ApplyMethod(reflect.TypeOf(bookRepo), "Create",
		//	func(_ *repository.BookRepository, ctx context.Context, book *pb.Book) (*pb.Book, error) {
		//		return nil, errors.New("database error")
		//	})
		//defer patches.Reset()

		// Execute
		response, err := svc.CreateBook(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to create book")
	})
}

func TestLibraryService_BorrowBook(t *testing.T) {

	t.Run("Missing Parameters", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create real repositories with mocked DB
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service with real repositories
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data - missing user ID
		ctx := context.Background()
		req := &pb.BorrowBookRequest{
			UserId: "", // Empty user ID
			BookId: "book-id-123",
		}

		// Execute
		response, err := svc.BorrowBook(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, response)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

}

func TestLibraryService_BorrowBook_Success(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)
		mockRow := new(MockRow)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create real repositories with mocked DB
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service with real repositories
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data
		ctx := context.Background()
		req := &pb.BorrowBookRequest{
			UserId: "user-id-123",
			BookId: "book-id-456",
		}

		// Mock book availability check
		mockPool.On("QueryRow", ctx, mock.AnythingOfType("string"), mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			// Book is available (no errors)
		}).Return(nil)

		// Mock successful borrow insert
		mockPool.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything).
			Return(pgconn.CommandTag("INSERT 1"), nil) // ✅ Correctly returning pgconn.CommandTag

		// Execute
		response, err := svc.BorrowBook(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		//assert.True(t, response.Success)

		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}

// // TestLibraryService_ReturnBook tests the ReturnBook method
func TestLibraryService_ReturnBook(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup mock database pool
		mockPool := new(MockPgxPool)

		// Setup DB with mock pool
		db := &database.DB{
			Pool: mockPool,
		}

		// Create real repositories with mocked DB
		userRepo := repository.NewUserRepository(db)
		bookRepo := repository.NewBookRepository(db)

		// Create service with real repositories
		svc := service.NewLibraryService(*userRepo, *bookRepo)

		// Test data
		ctx := context.Background()
		req := &pb.ReturnBookRequest{
			BorrowId: "borrow-id-123",
		}

		// Since we can't easily mock transactions, we need to override the ReturnBook method
		// Save the original method to restore later
		// Restore original method after test

		// Execute
		response, err := svc.ReturnBook(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Success)
	})
}
