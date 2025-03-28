package service

import (
	"context"
	"regexp"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"library-management-service/internal/repository"
	pb "library-management-service/proto/library/v1"
)

type LibraryService struct {
	pb.UnimplementedLibraryServiceServer
	userRepo repository.UserRepositoryInterface
	bookRepo repository.BookRepositoryInterface
}

//	func NewLibraryService(userRepo *repository.UserRepository, bookRepo *repository.BookRepository) *LibraryService {
//		return &LibraryService{
//			userRepo: userRepo,
//			bookRepo: bookRepo,
//		}
//	}
func NewLibraryService(userRepo repository.UserRepositoryInterface, bookRepo repository.BookRepositoryInterface) *LibraryService {
	return &LibraryService{
		userRepo: userRepo,
		bookRepo: bookRepo,
	}
}

// User-related methods
func (s *LibraryService) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	// Validate inputs
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "name, email, and password are required")
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}

	// Validate password strength
	if len(req.Password) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}

	user, err := s.userRepo.Create(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.RegisterUserResponse{User: user}, nil
}

func (s *LibraryService) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.userRepo.VerifyCredentials(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// In a real implementation, you'd generate a JWT token here
	token := "sample-jwt-token"

	return &pb.LoginUserResponse{
		User:  user,
		Token: token,
	}, nil
}

// Book-related methods
func (s *LibraryService) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.CreateBookResponse, error) {
	if req.Book == nil {
		return nil, status.Error(codes.InvalidArgument, "book is required")
	}

	if req.Book.Title == "" || req.Book.Author == "" {
		return nil, status.Error(codes.InvalidArgument, "title and author are required")
	}

	book, err := s.bookRepo.Create(ctx, req.Book)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create book: %v", err)
	}

	return &pb.CreateBookResponse{Book: book}, nil
}

func (s *LibraryService) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "book id is required")
	}

	book, err := s.bookRepo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "book not found: %v", err)
	}

	return &pb.GetBookResponse{Book: book}, nil
}

func (s *LibraryService) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	pageSize := int32(10) // Default page size
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	// In a real application, you'd implement proper pagination with tokens
	// For simplicity, we'll just use an offset of 0
	books, err := s.bookRepo.List(ctx, pageSize, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list books: %v", err)
	}

	return &pb.ListBooksResponse{
		Books: books,
	}, nil
}

func (s *LibraryService) BorrowBook(ctx context.Context, req *pb.BorrowBookRequest) (*pb.BorrowBookResponse, error) {
	if req.UserId == "" || req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id and book id are required")
	}

	// Set due date to 14 days from now
	dueDate := time.Now().AddDate(0, 0, 14)

	borrowID, err := s.bookRepo.BorrowBook(ctx, req.UserId, req.BookId, dueDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to borrow book: %v", err)
	}

	return &pb.BorrowBookResponse{
		BorrowId: borrowID,
		DueDate:  dueDate.Format(time.RFC3339),
	}, nil
}

func (s *LibraryService) ReturnBook(ctx context.Context, req *pb.ReturnBookRequest) (*pb.ReturnBookResponse, error) {
	if req.BorrowId == "" {
		return nil, status.Error(codes.InvalidArgument, "borrow id is required")
	}

	err := s.bookRepo.ReturnBook(ctx, req.BorrowId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to return book: %v", err)
	}

	return &pb.ReturnBookResponse{
		Success: true,
	}, nil
}

func (s *LibraryService) CheckBookAvailability(ctx context.Context, req *pb.CheckBookAvailabilityRequest) (*pb.CheckBookAvailabilityResponse, error) {
	if req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "book id is required")
	}

	book, err := s.bookRepo.GetByID(ctx, req.BookId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "book not found: %v", err)
	}

	statusMsg := "Borrowed"
	if book.Available {
		statusMsg = "Available"
	}

	return &pb.CheckBookAvailabilityResponse{
		Available: book.Available,
		Status:    statusMsg,
	}, nil
}
