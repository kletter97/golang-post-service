package main

import (
	"context"
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

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("Application starting...")

	// DB connection
	dbURL := "postgres://postgres:postgres@localhost:5432/postgres"

	log.Println("Connecting to database...")

	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	log.Println("Database connected")

	// Init tables
	log.Println("Initializing database tables...")

	err = repository.InitTables(context.Background(), dbpool)
	if err != nil {
		log.Fatalf("Failed to init tables: %v", err)
	}

	log.Println("Tables initialized")

	// Repository
	testRepo := repository.NewPostgresRepository(dbpool)

	// Service
	testService := service.NewTestService(testRepo)

	// Handler
	testHandler := deliveryHTTP.NewTestHandler(testService)

	// Routes
	mux := http.NewServeMux()
	testHandler.RegisterRoutes(mux)

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
