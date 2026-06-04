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
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Application starting...")
	ctx := context.Background()

	// 1. Получаем хост базы данных из переменной окружения.
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	// 2. Подставляем dbHost в URL подключения
	dbURL := fmt.Sprintf("postgres://postgres:postgres@%s:5432/postgres?sslmode=disable", dbHost)

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

	// 3. Подключаемся к RabbitMQ (в докере хост — это имя сервиса)
	log.Println("Connecting to RabbitMQ...")
	rabbitService, err := repository.NewRabbitMQService("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("RabbitMQ init failed: %v", err)
	}
	defer rabbitService.Close()
	log.Println("RabbitMQ connected")

	// 4. Запускаем фоновый воркер! Он уходит в фон и будет слушать очередь 24/7
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

	// Service
	testService := service.NewTestService(testRepo)
	passwordHasher := service.NewBcryptHasher()
	tokenGenerator := service.NewJWTTokenGenerator("some-secret-key", 24*time.Hour)
	userService := service.NewUserService(userRepo, "some-secret-key", passwordHasher, tokenGenerator)
	postService := service.NewPostService(postRepo)

	// Handler (Передаем rabbitService вторым аргументом)
	testHandler := deliveryHTTP.NewTestHandler(testService)
	userHandler := deliveryHTTP.NewUserHandler(userService)
	postHandler := deliveryHTTP.NewPostHandler(postService, rabbitService)

	// Routes
	mux := http.NewServeMux()
	testHandler.RegisterRoutes(mux)
	userHandler.RegisterRoutes(mux)
	postHandler.RegisterRoutes(mux)

	mux.Handle("/metrics", promhttp.Handler())

	authMiddleware := deliveryHTTP.AuthMiddleware("some-secret-key")
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