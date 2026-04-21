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
)

func main() {

	// слой 3 (repository)
	testRepo := repository.NewTestRepository()

	// слой 2 (service)
	testService := service.NewTestService(testRepo)

	// слой 1 (delivery)
	testHandler := deliveryHTTP.NewTestHandler(testService)

	// Регистрация маршрутов
	mux := http.NewServeMux()
	testHandler.RegisterRoutes(mux)


	// Настройка http-сервера
	port := ":8080"
	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}


	// Graceful Shutdown

	// Канал для перехвата OS-сигналов
	quit := make(chan os.Signal, 1)

	// signal.Notify перенаправляет сигналы в канал quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине, чтобы не блокировать main
	go func() {
		fmt.Printf("Server is starting on port %s...\n", port)
		fmt.Printf("Test endpoint: http://localhost%s/test\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Блокируемся до получения сигнала завершения
	sig := <-quit
	fmt.Printf("\nReceived signal: %v. Shutting down gracefully...\n", sig)

	// 30 секунд на завершение активных запросов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server stopped gracefully")
}