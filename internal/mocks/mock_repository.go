package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"library-management-service/internal/repository"
	pb "library-management-service/proto/library/v1"
)

// Ensure type safety by verifying that MockUserRepository implements UserRepositoryInterface
var _ repository.UserRepositoryInterface = (*MockUserRepository)(nil)

// MockUserRepository is a mock implementation of UserRepositoryInterface for testing
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

// Ensure type safety by verifying that MockBookRepository implements BookRepositoryInterface
var _ repository.BookRepositoryInterface = (*MockBookRepository)(nil)

// MockBookRepository is a mock implementation of BookRepositoryInterface for testing
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
