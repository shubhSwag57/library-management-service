package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"library-management-service/internal/database"
	pb "library-management-service/proto/library/v1"

	"testing"
	"time"
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
type MockRows struct {
	mock.Mock
	index int
	data  [][5]string // [id, title, author, isbn, available]
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

// TestBookRepository_Create tests the Create method
func TestBookRepository_Create(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := NewBookRepository(db)
	ctx := context.Background()

	// Test data
	book := &pb.Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Isbn:      "1234567890",
		Available: true,
	}

	// Expectations
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		// Simulate filling the book fields
		dests := args.Get(0).([]interface{})
		*(dests[0].(*string)) = "book-id-123"
		*(dests[1].(*string)) = book.Title
		*(dests[2].(*string)) = book.Author
		*(dests[3].(*string)) = book.Isbn
		*(dests[4].(*bool)) = book.Available
	}).Return(nil)

	// Execute
	result, err := repo.Create(ctx, book)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "book-id-123", result.Id)
	assert.Equal(t, book.Title, result.Title)
	assert.Equal(t, book.Author, result.Author)
	assert.Equal(t, book.Isbn, result.Isbn)
	assert.Equal(t, book.Available, result.Available)

	// Verify correct parameters were passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, book.Title, argsSlice[0])
	assert.Equal(t, book.Author, argsSlice[1])
	assert.Equal(t, book.Isbn, argsSlice[2])
	assert.Equal(t, book.Available, argsSlice[3])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

// TestBookRepository_Create_Error tests the Create method with a database error
func TestBookRepository_Create_Error(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := NewBookRepository(db)
	ctx := context.Background()

	// Test data
	book := &pb.Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Isbn:      "1234567890",
		Available: true,
	}

	// Expectations
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(errors.New("database error"))

	// Execute
	result, err := repo.Create(ctx, book)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create book")

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

// TestBookRepository_GetByID tests the GetByID method
func TestBookRepository_GetByID(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := NewBookRepository(db)
	ctx := context.Background()

	// Test data
	bookID := "book-id-123"
	expectedBook := &pb.Book{
		Id:        bookID,
		Title:     "Test Book",
		Author:    "Test Author",
		Isbn:      "1234567890",
		Available: true,
	}

	// Expectations
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		// Simulate filling the book fields
		dests := args.Get(0).([]interface{})
		*(dests[0].(*string)) = expectedBook.Id
		*(dests[1].(*string)) = expectedBook.Title
		*(dests[2].(*string)) = expectedBook.Author
		*(dests[3].(*string)) = expectedBook.Isbn
		*(dests[4].(*bool)) = expectedBook.Available
	}).Return(nil)

	// Execute
	book, err := repo.GetByID(ctx, bookID)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, expectedBook.Id, book.Id)
	assert.Equal(t, expectedBook.Title, book.Title)
	assert.Equal(t, expectedBook.Author, book.Author)
	assert.Equal(t, expectedBook.Isbn, book.Isbn)
	assert.Equal(t, expectedBook.Available, book.Available)

	// Verify correct ID was passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, bookID, argsSlice[0])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

// TestBookRepository_GetByID_NotFound tests GetByID with nonexistent book
func TestBookRepository_GetByID_NotFound(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRow := new(MockRow)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := NewBookRepository(db)
	ctx := context.Background()

	// Test data
	bookID := "nonexistent-id"

	// Expectations
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

	// Execute
	book, err := repo.GetByID(ctx, bookID)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, book)
	assert.Contains(t, err.Error(), "book not found")

	// Verify correct ID was passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, bookID, argsSlice[0])

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

// TestBookRepository_List tests the List method
func TestBookRepository_List(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockRows := &MockRows{
		data: [][5]string{
			{"book-id-1", "Book 1", "Author 1", "ISBN1", "true"},
			{"book-id-2", "Book 2", "Author 2", "ISBN2", "false"},
		},
	}

	db := &database.DB{
		Pool: mockPool,
	}

	repo := NewBookRepository(db)
	ctx := context.Background()

	// Test data
	limit := int32(10)
	offset := int32(0)

	// Expectations
	mockPool.On("Query", ctx, mock.Anything, mock.Anything).Return(mockRows, nil)
	mockRows.On("Close").Return()

	// Execute
	books, err := repo.List(ctx, limit, offset)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, books)
	assert.Len(t, books, 2)

	// Verify first book
	assert.Equal(t, "book-id-1", books[0].Id)
	assert.Equal(t, "Book 1", books[0].Title)
	assert.Equal(t, "Author 1", books[0].Author)
	assert.Equal(t, "ISBN1", books[0].Isbn)
	assert.True(t, books[0].Available)

	// Verify second book
	assert.Equal(t, "book-id-2", books[1].Id)
	assert.Equal(t, "Book 2", books[1].Title)
	assert.Equal(t, "Author 2", books[1].Author)
	assert.Equal(t, "ISBN2", books[1].Isbn)
	assert.False(t, books[1].Available)

	// Verify correct parameters were passed
	argsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, limit, argsSlice[0])
	assert.Equal(t, offset, argsSlice[1])

	mockPool.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

// TestBookRepository_List_QueryError tests List with a database query error
func TestBookRepository_List_QueryError(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := NewBookRepository(db)
	ctx := context.Background()

	// Test data
	limit := int32(10)
	offset := int32(0)

	// Expectations
	mockPool.On("Query", ctx, mock.Anything, mock.Anything).Return((*MockRows)(nil), errors.New("query error"))

	// Execute
	books, err := repo.List(ctx, limit, offset)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.Contains(t, err.Error(), "failed to list books")

	mockPool.AssertExpectations(t)
}

// TestBookRepository_BorrowBook tests the BorrowBook method
func TestBookRepository_BorrowBook(t *testing.T) {
	// Setup
	mockPool := new(MockPgxPool)
	mockAvailableRow := new(MockRow)
	mockBorrowRow := new(MockRow)
	mockCommandTag := new(MockPgxPool)

	db := &database.DB{
		Pool: mockPool,
	}

	repo := NewBookRepository(db)
	ctx := context.Background()

	// Test data
	userID := "user-id-123"
	bookID := "book-id-123"
	dueDate := time.Now().AddDate(0, 0, 14)
	borrowID := "borrow-id-123"

	// Expectations
	// 1. Check if book is available
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockAvailableRow).Once()
	mockAvailableRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		// Book is available
		*(args.Get(0).([]interface{})[0].(*bool)) = true
	}).Return(nil)

	// 2. Update book availability
	mockPool.On("Exec", ctx, mock.Anything, mock.Anything).Return(mockCommandTag, nil).Once()
	mockCommandTag.On("RowsAffected").Return(int64(1))

	// 3. Create borrow record
	mockPool.On("QueryRow", ctx, mock.Anything, mock.Anything).Return(mockBorrowRow).Once()
	mockBorrowRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args.Get(0).([]interface{})[0].(*string)) = borrowID
	}).Return(nil)

	// Execute
	result, err := repo.BorrowBook(ctx, userID, bookID, dueDate)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, borrowID, result)

	// Verify correct parameters were passed for book availability check
	availableArgsSlice := mockPool.Calls[0].Arguments[2].([]interface{})
	assert.Equal(t, bookID, availableArgsSlice[0])

}
