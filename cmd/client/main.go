package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "library-management-service/proto/library/v1"
)

func main() {
	// Connect to the server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create service client
	client := pb.NewLibraryServiceClient(conn)

	// Set timeout for our operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("=== Library Service Test Client ===")

	// 1. Register a user
	fmt.Println("\n[1] Registering a user...")
	userResp, err := client.RegisterUser(ctx, &pb.RegisterUserRequest{
		Name:     "Test User4",
		Email:    "test@example4.com",
		Password: "password1234",
	})
	if err != nil {
		log.Fatalf("RegisterUser failed: %v", err)
	}
	fmt.Printf("User registered successfully! ID: %s\n", userResp.User.Id)
	userID := userResp.User.Id

	// 2. Login with the user
	fmt.Println("\n[2] Logging in...")
	loginResp, err := client.LoginUser(ctx, &pb.LoginUserRequest{
		Email:    "test@example1.com",
		Password: "password1234",
	})
	if err != nil {
		log.Fatalf("LoginUser failed: %v", err)
	}
	fmt.Printf("Login successful! Token: %s\n", loginResp.Token)

	// 3. Create a First book
	fmt.Println("\n[3] Creating a book...")
	bookResp, err := client.CreateBook(ctx, &pb.CreateBookRequest{
		Book: &pb.Book{
			Title:     "The Test Book3",
			Author:    "Test Author",
			Isbn:      "1234567893",
			Available: true,
		},
	})
	if err != nil {
		log.Fatalf("CreateBook failed: %v", err)
	}
	fmt.Printf("Book created! ID: %s\n", bookResp.Book.Id)
	bookID := bookResp.Book.Id

	// 4. Get the book
	fmt.Println("\n[4] Getting book details...")
	getBookResp, err := client.GetBook(ctx, &pb.GetBookRequest{
		Id: bookID,
	})
	if err != nil {
		log.Fatalf("GetBook failed: %v", err)
	}
	fmt.Printf("Book details: %s by %s (Available: %v)\n",
		getBookResp.Book.Title,
		getBookResp.Book.Author,
		getBookResp.Book.Available)

	// 5. List all books
	fmt.Println("\n[5] Listing all books...")
	listResp, err := client.ListBooks(ctx, &pb.ListBooksRequest{
		PageSize: 10,
	})
	if err != nil {
		log.Fatalf("ListBooks failed: %v", err)
	}
	fmt.Printf("Found %d books:\n", len(listResp.Books))
	for i, book := range listResp.Books {
		fmt.Printf("  %d. %s by %s (ID: %s)\n", i+1, book.Title, book.Author, book.Id)
	}

	// 6. Borrow the book
	fmt.Println("\n[6] Borrowing a book...")
	borrowResp, err := client.BorrowBook(ctx, &pb.BorrowBookRequest{
		UserId: userID,
		BookId: bookID,
	})
	if err != nil {
		log.Fatalf("BorrowBook failed: %v", err)
	}
	fmt.Printf("Book borrowed! Borrow ID: %s, Due date: %s\n",
		borrowResp.BorrowId, borrowResp.DueDate)
	borrowID := borrowResp.BorrowId
	fmt.Println("Book returned successfully: %v\n", borrowID)
	// 7. Return the book
	fmt.Println("\n[7] Returning the book...")
	returnResp, err := client.ReturnBook(ctx, &pb.ReturnBookRequest{
		BorrowId: borrowID,
	})
	if err != nil {
		log.Fatalf("ReturnBook failed: %v", err)
	}
	fmt.Printf("Book returned successfully: %v\n", returnResp.Success)

	fmt.Println("\nAll tests completed successfully!")
}
