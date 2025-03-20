package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"library-management-service/internal/database"
	"library-management-service/internal/repository"
	"library-management-service/internal/service"
	pb "library-management-service/proto/library/v1"
)

func main() {
	// Get database connection string from environment or use default
	dbConnString := os.Getenv("DATABASE_URL")
	if dbConnString == "" {
		dbConnString = "postgresql://postgres:postgres@localhost:5432/library?sslmode=disable"
	}

	// Set up gRPC server port
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	// Connect to database
	db, err := database.NewDB(dbConnString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Set up database schema
	if err := db.SetupSchema(); err != nil {
		log.Fatalf("Failed to set up database schema: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	bookRepo := repository.NewBookRepository(db)

	// Initialize service
	libraryService := service.NewLibraryService(*userRepo, *bookRepo)

	// Create gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterLibraryServiceServer(grpcServer, libraryService)

	log.Printf("Starting gRPC server on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
