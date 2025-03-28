package repository

import (
	"context"
	pb "library-management-service/proto/library/v1"
	"time"
)

type BookRepositoryInterface interface {
	Create(ctx context.Context, book *pb.Book) (*pb.Book, error)
	GetByID(ctx context.Context, id string) (*pb.Book, error)
	List(ctx context.Context, limit, offset int32) ([]*pb.Book, error)
	BorrowBook(ctx context.Context, userID, bookID string, dueDate time.Time) (string, error)
	ReturnBook(ctx context.Context, borrowID string) error
}

type UserRepositoryInterface interface {
	Create(ctx context.Context, name, email, password string) (*pb.User, error)
	VerifyCredentials(ctx context.Context, email, password string) (*pb.User, error)
	GetByID(ctx context.Context, id string) (*pb.User, error)
}
