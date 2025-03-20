package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
	"library-management-service/internal/database"
	pb "library-management-service/proto/library/v1"
)
	
type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, name, email, password string) (*pb.User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var user pb.User
	err = r.db.Pool.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email
	`, name, email, string(hashedPassword)).Scan(&user.Id, &user.Name, &user.Email)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) VerifyCredentials(ctx context.Context, email, password string) (*pb.User, error) {
	var user pb.User
	var passwordHash string

	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, email, password_hash 
		FROM users 
		WHERE email = $1
	`, email).Scan(&user.Id, &user.Name, &user.Email, &passwordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Compare hashed password with provided password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*pb.User, error) {
	var user pb.User

	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, email 
		FROM users 
		WHERE id = $1
	`, id).Scan(&user.Id, &user.Name, &user.Email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}
