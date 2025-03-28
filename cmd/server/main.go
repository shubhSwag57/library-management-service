package main

import (
	"google.golang.org/grpc"
	"library-management-service/internal/database"
	"library-management-service/internal/repository"
	"library-management-service/internal/server"
	"library-management-service/internal/service"
	pb "library-management-service/proto/library/v1"
	"log"
	"net"
)

func main() {
	// Initialize database
	db, err := database.NewDB("postgres://postgres:password@localhost:5432/library")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup schema
	if err := db.SetupSchema(); err != nil {
		log.Fatalf("Failed to setup database schema: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	bookRepo := repository.NewBookRepository(db)

	// Initialize service
	libraryService := service.NewLibraryService(userRepo, bookRepo)

	// Start gRPC server in a goroutine
	go startGRPCServer(libraryService)

	// Start REST server
	startRESTServer(libraryService)
}

func startGRPCServer(libraryService *service.LibraryService) {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterLibraryServiceServer(grpcServer, libraryService)

	log.Println("gRPC server is running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func startRESTServer(libraryService *service.LibraryService) {
	restServer := server.NewRESTServer(libraryService)

	log.Println("REST server is running on :8086")
	if err := restServer.Start(":8086"); err != nil {
		log.Fatalf("Failed to serve REST: %v", err)
	}
}
