package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"library-management-service/internal/database"
	pb "library-management-service/proto/library/v1"
)

type BookRepository struct {
	db *database.DB
}

func NewBookRepository(db *database.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, book *pb.Book) (*pb.Book, error) {
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO books (title, author, isbn, available)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, author, isbn, available
	`, book.Title, book.Author, book.Isbn, book.Available).Scan(
		&book.Id, &book.Title, &book.Author, &book.Isbn, &book.Available)

	if err != nil {
		return nil, fmt.Errorf("failed to create book: %w", err)
	}

	return book, nil
}

func (r *BookRepository) GetByID(ctx context.Context, id string) (*pb.Book, error) {
	var book pb.Book

	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, title, author, isbn, available 
		FROM books 
		WHERE id = $1
	`, id).Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, &book.Available)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("book not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &book, nil
}

func (r *BookRepository) List(ctx context.Context, limit int32, offset int32) ([]*pb.Book, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, title, author, isbn, available 
		FROM books 
		ORDER BY title 
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list books: %w", err)
	}
	defer rows.Close()

	var books []*pb.Book
	for rows.Next() {
		var book pb.Book
		if err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, &book.Available); err != nil {
			return nil, fmt.Errorf("failed to scan book: %w", err)
		}
		books = append(books, &book)
	}

	return books, nil
}

func (r *BookRepository) BorrowBook(ctx context.Context, userID, bookID string, dueDate time.Time) (string, error) {
	// Since the Pool interface doesn't expose Begin directly, we need to implement
	// transaction logic without using that method directly

	// Check if book is available
	var available bool
	err := r.db.Pool.QueryRow(ctx, "SELECT available FROM books WHERE id = $1", bookID).Scan(&available)
	if err != nil {
		return "", fmt.Errorf("failed to check book availability: %w", err)
	}
	if !available {
		return "", fmt.Errorf("book is not available")
	}

	// Update book availability
	_, err = r.db.Pool.Exec(ctx, "UPDATE books SET available = false WHERE id = $1", bookID)
	if err != nil {
		return "", fmt.Errorf("failed to update book availability: %w", err)
	}

	// Create borrow record
	var borrowID string
	err = r.db.Pool.QueryRow(ctx, `
		INSERT INTO borrows (user_id, book_id, due_date)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, bookID, dueDate).Scan(&borrowID)
	if err != nil {
		// If there was an error, try to revert the book availability
		_, revertErr := r.db.Pool.Exec(ctx, "UPDATE books SET available = true WHERE id = $1", bookID)
		if revertErr != nil {
			// Log but continue with original error
			fmt.Printf("Failed to revert book availability: %v\n", revertErr)
		}
		return "", fmt.Errorf("failed to create borrow record: %w", err)
	}

	return borrowID, nil
}

func (r *BookRepository) ReturnBook(ctx context.Context, borrowID string) error {
	// Get book ID from borrow
	var bookID string
	err := r.db.Pool.QueryRow(ctx, "SELECT book_id FROM borrows WHERE id = $1", borrowID).Scan(&bookID)
	if err != nil {
		return fmt.Errorf("failed to get borrow: %w", err)
	}

	// Update book availability
	_, err = r.db.Pool.Exec(ctx, "UPDATE books SET available = true WHERE id = $1", bookID)
	if err != nil {
		return fmt.Errorf("failed to update book availability: %w", err)
	}

	// Update borrow record
	_, err = r.db.Pool.Exec(ctx, "UPDATE borrows SET return_date = NOW() WHERE id = $1", borrowID)
	if err != nil {
		// If this fails, try to revert the book availability
		_, revertErr := r.db.Pool.Exec(ctx, "UPDATE books SET available = false WHERE id = $1", bookID)
		if revertErr != nil {
			// Log but continue with original error
			fmt.Printf("Failed to revert book availability: %v\n", revertErr)
		}
		return fmt.Errorf("failed to update borrow record: %w", err)
	}

	return nil
}
