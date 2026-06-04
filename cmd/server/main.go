package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	deliveryHTTP "post-service/internal/delivery/http"
	"post-service/internal/repository"
	"post-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Application starting...")
	ctx := context.Background()

	// получаем хост базы данных из переменной окружения.
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	log.Println("Connecting to database...")

	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	// Проверяем связь с базой данных
	if err := dbpool.Ping(ctx); err != nil {
		log.Fatalf("Database connection ping failed: %v", err)
	}
	log.Println("Database connected")

	// подключаемся к RabbitMQ (в докере хост — это имя сервиса)
	log.Println("Connecting to RabbitMQ...")

	rabbitURL := os.Getenv("RABBITMQ_URL")
	rabbitService, err := repository.NewRabbitMQService(rabbitURL)

	if err != nil {
		log.Fatalf("RabbitMQ init failed: %v", err)
	}
	defer rabbitService.Close()
	log.Println("RabbitMQ connected")

	// запускаем фоновый воркер (слушает очередь 24/7)
	rabbitService.StartAuditWorker(ctx, dbpool)

	// Init tables
	log.Println("Initializing database tables...")
	err = repository.InitTables(ctx, dbpool)
	if err != nil {
		log.Fatalf("Failed to init tables: %v", err)
	}
	log.Println("Tables initialized")

	// Repository
	testRepo := repository.NewPostgresRepository(dbpool)
	userRepo := repository.NewPostgresUserRepository(dbpool)
	postRepo := repository.NewPostgresPostRepository(dbpool)

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtHours := 24
	parsed, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
	if err == nil {
		jwtHours = parsed
	}

	// Service
	testService := service.NewTestService(testRepo)
	passwordHasher := service.NewBcryptHasher()
	tokenGenerator := service.NewJWTTokenGenerator(jwtSecret, time.Duration(jwtHours)*time.Hour)
	userService := service.NewUserService(userRepo, jwtSecret, passwordHasher, tokenGenerator)
	postService := service.NewPostService(postRepo)

	// Handler
	testHandler := deliveryHTTP.NewTestHandler(testService)
	userHandler := deliveryHTTP.NewUserHandler(userService)
	postHandler := deliveryHTTP.NewPostHandler(postService, rabbitService)

	// Routes
	mux := http.NewServeMux()
	testHandler.RegisterRoutes(mux)
	userHandler.RegisterRoutes(mux)
	postHandler.RegisterRoutes(mux)

	mux.Handle("/metrics", promhttp.Handler())

	authMiddleware := deliveryHTTP.AuthMiddleware(jwtSecret)
	mux.Handle("/posts/my", authMiddleware(http.HandlerFunc(postHandler.GetMyPosts)))

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
	fmt.Println("Bye")
}
