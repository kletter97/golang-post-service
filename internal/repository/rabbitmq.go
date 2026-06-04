package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQService struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewRabbitMQService создает подключение и объявляет очередь
func NewRabbitMQService(url string) (*RabbitMQService, error) {
	var conn *amqp.Connection
	var err error

	// Пробуем подключиться 10 раз с паузой в 2 секунды
	for i := 1; i <= 10; i++ {
		log.Printf("Trying to connect to RabbitMQ (Attempt %d/10)...", i)
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}

		log.Printf("RabbitMQ not ready yet, waiting 2 seconds... (Error: %v)", err)
		time.Sleep(2 * time.Second)
	}

	// если не подключились с 10 попыток
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq after multiple attempts: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		"post_audit_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &RabbitMQService{conn: conn, ch: ch}, nil
}

// PublishAuditLog вызывается в хэндлере, чтобы кинуть сообщение в очередь
func (r *RabbitMQService) PublishAuditLog(ctx context.Context, message string) error {
	return r.ch.PublishWithContext(ctx,
		"",                 // exchange
		"post_audit_queue", // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

// StartAuditWorker — наш воркер, слушает очередь в фоне.
func (r *RabbitMQService) StartAuditWorker(ctx context.Context, db *pgxpool.Pool) {
	msgs, err := r.ch.Consume(
		"post_audit_queue", // имя очереди
		"",                 // consumer tag
		true,               // auto-ack (сообщение удаляется из очереди сразу после вычитки)
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,
	)
	if err != nil {
		log.Printf("Failed to start consuming: %v", err)
		return
	}

	// Запускаем бесконечный цикл в отдельной горутине (go func)
	go func() {
		log.Println("RabbitMQ Audit Worker successfully started...")
		for d := range msgs {
			log.Printf("Received a message from RabbitMQ: %s", d.Body)

			// Записываем полученный из очереди текст прямо в таблицу messages
			query := `INSERT INTO messages (message) VALUES ($1)`
			_, err := db.Exec(ctx, query, string(d.Body))
			if err != nil {
				log.Printf("Worker failed to insert message to DB: %v", err)
			} else {
				log.Printf("Worker successfully saved audit log to Postgres!")
			}
		}
	}()
}

// Close закрывает соединения при выходе из приложения
func (r *RabbitMQService) Close() {
	if r.ch != nil {
		r.ch.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
