package server

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"library-management-service/internal/service"
	pb "library-management-service/proto/library/v1"
	"net/http"
)

type RESTServer struct {
	libraryService *service.LibraryService
	router         *gin.Engine
}

func NewRESTServer(libraryService *service.LibraryService) *RESTServer {
	server := &RESTServer{
		libraryService: libraryService,
		router:         gin.Default(),
	}
	server.setupRoutes()
	return server
}

func (s *RESTServer) setupRoutes() {
	// User routes
	s.router.POST("/api/users/registerUser", s.registerUser)
	s.router.POST("/api/users/loginUser", s.loginUser)

	// Book routes
	s.router.POST("/api/books", s.createBook)
	s.router.GET("/api/books/:id", s.getBook)
	s.router.GET("/api/books", s.listBooks)
	s.router.POST("/api/books/:id/borrowBook", s.borrowBook)
	s.router.POST("/api/books/returnBook", s.returnBook)
	s.router.GET("/api/books/:id/availability", s.checkBookAvailability)

}

func (s *RESTServer) Start(addr string) error {
	return s.router.Run(addr)
}

// Handler implementations
func (s *RESTServer) registerUser(c *gin.Context) {
	var request struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	grpcReq := &pb.RegisterUserRequest{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	}

	response, err := s.libraryService.RegisterUser(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    response.User.Id,
		"name":  response.User.Name,
		"email": response.User.Email,
	})
}

func (s *RESTServer) loginUser(c *gin.Context) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	grpcReq := &pb.LoginUserRequest{
		Email:    request.Email,
		Password: request.Password,
	}

	response, err := s.libraryService.LoginUser(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    response.User.Id,
			"name":  response.User.Name,
			"email": response.User.Email,
		},
		"token": response.Token,
	})
}

func (s *RESTServer) createBook(c *gin.Context) {
	var request struct {
		Title     string `json:"title"`
		Author    string `json:"author"`
		Isbn      string `json:"isbn"`
		Available bool   `json:"available"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	grpcReq := &pb.CreateBookRequest{
		Book: &pb.Book{
			Title:     request.Title,
			Author:    request.Author,
			Isbn:      request.Isbn,
			Available: request.Available,
		},
	}

	response, err := s.libraryService.CreateBook(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        response.Book.Id,
		"title":     response.Book.Title,
		"author":    response.Book.Author,
		"isbn":      response.Book.Isbn,
		"available": response.Book.Available,
	})
}

func (s *RESTServer) getBook(c *gin.Context) {
	bookID := c.Param("id")

	grpcReq := &pb.GetBookRequest{
		Id: bookID,
	}

	response, err := s.libraryService.GetBook(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        response.Book.Id,
		"title":     response.Book.Title,
		"author":    response.Book.Author,
		"isbn":      response.Book.Isbn,
		"available": response.Book.Available,
	})
}

func (s *RESTServer) listBooks(c *gin.Context) {
	pageSize := 10 // Default page size
	if pageSizeParam := c.Query("page_size"); pageSizeParam != "" {
		if size, err := parseInt32(pageSizeParam); err == nil {
			pageSize = int(size)
		}
	}

	grpcReq := &pb.ListBooksRequest{
		PageSize: int32(pageSize),
	}

	response, err := s.libraryService.ListBooks(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	books := make([]map[string]interface{}, 0, len(response.Books))
	for _, book := range response.Books {
		books = append(books, map[string]interface{}{
			"id":        book.Id,
			"title":     book.Title,
			"author":    book.Author,
			"isbn":      book.Isbn,
			"available": book.Available,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"books": books,
	})
}

func (s *RESTServer) borrowBook(c *gin.Context) {
	bookID := c.Param("id")

	var request struct {
		UserID string `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	grpcReq := &pb.BorrowBookRequest{
		UserId: request.UserID,
		BookId: bookID,
	}

	response, err := s.libraryService.BorrowBook(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"borrow_id": response.BorrowId,
		"due_date":  response.DueDate,
	})
}

func (s *RESTServer) returnBook(c *gin.Context) {
	var request struct {
		BorrowID string `json:"borrow_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	grpcReq := &pb.ReturnBookRequest{
		BorrowId: request.BorrowID,
	}

	response, err := s.libraryService.ReturnBook(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
	})
}

// Helper function to parse int32
func parseInt32(s string) (int32, error) {
	var result int
	err := json.Unmarshal([]byte(s), &result)
	return int32(result), err
}

func (s *RESTServer) checkBookAvailability(c *gin.Context) {
	bookID := c.Param("id")

	grpcReq := &pb.CheckBookAvailabilityRequest{
		BookId: bookID,
	}

	response, err := s.libraryService.CheckBookAvailability(c.Request.Context(), grpcReq)
	if err != nil {
		status, ok := status.FromError(err)
		if ok && status.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available": response.Available,
		"status":    response.Status,
	})
}
