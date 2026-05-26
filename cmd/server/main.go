package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryHTTP "post-service/internal/delivery/http"
	"post-service/internal/repository"
	"post-service/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("Application starting...")

	// DB connection
	dbURL := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

	log.Println("Connecting to database...")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	log.Println("Database connected")

	// Init tables
	log.Println("Initializing database tables...")

	err = repository.InitTables(context.Background(), db)
	if err != nil {
		log.Fatalf("Failed to init tables: %v", err)
	}

	log.Println("Tables initialized")

	// Repository
	testRepo := repository.NewPostgresRepository(db)
	userRepo := repository.NewPostgresUserRepository(db)

	// Service
	testService := service.NewTestService(testRepo)
	passwordHasher := service.NewBcryptHasher()
	tokenGenerator := service.NewJWTTokenGenerator("your-secret-key", 24*time.Hour)
	userService := service.NewUserService(userRepo, "your-secret-key", passwordHasher, tokenGenerator)

	// Handler
	testHandler := deliveryHTTP.NewTestHandler(testService)
	userHandler := deliveryHTTP.NewUserHandler(userService)

	// Routes
	mux := http.NewServeMux()
	testHandler.RegisterRoutes(mux)
	userHandler.RegisterRoutes(mux)

	log.Println("Routes registered")

	// Server config
	port := ":8090"

	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	go func() {
		log.Printf("Server is running on http://localhost%s", port)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait signal
	sig := <-quit
	log.Printf("Received signal: %v", sig)

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
	fmt.Println("Bye")
}
