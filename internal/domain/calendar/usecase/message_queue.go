package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"mail2calendar/internal/domain/calendar/service"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// MessageQueueService defines the interface for async message processing
type MessageQueueService interface {
	PublishEmailEvent(ctx context.Context, emailContent string, userID string) error
	ProcessMessages(ctx context.Context) error
	Close() error
}

// QueueConfig holds RabbitMQ configuration
type QueueConfig struct {
	URI               string
	EmailQueueName    string
	DeadLetterQueue   string
	MaxRetries        int
	RetryDelaySeconds int
}

type messagingService struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	config   QueueConfig
	calendar service.CalendarService // Changed to use the correct interface
	tracer   trace.Tracer
	logger   *logrus.Logger
}

// EmailMessage represents a message in the queue
type EmailMessage struct {
	EmailContent string    `json:"email_content"`
	UserID       string    `json:"user_id"`
	RetryCount   int       `json:"retry_count"`
	Timestamp    time.Time `json:"timestamp"`
}

// NewMessageQueueService creates a new instance of MessageQueueService
func NewMessageQueueService(config QueueConfig, calendar service.CalendarService) (MessageQueueService, error) { // Updated parameter type
	conn, err := amqp.Dial(config.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// Declare queues
	if err := declareQueues(ch, config); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queues: %v", err)
	}

	return &messagingService{
		conn:     conn,
		channel:  ch,
		config:   config,
		calendar: calendar,
		tracer:   otel.Tracer("message-queue-service"),
		logger:   logrus.New(),
	}, nil
}

func (s *messagingService) PublishEmailEvent(ctx context.Context, emailContent string, userID string) error {
	ctx, span := s.tracer.Start(ctx, "PublishEmailEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.Int("email_content_length", len(emailContent)),
	)

	msg := EmailMessage{
		EmailContent: emailContent,
		UserID:       userID,
		RetryCount:   0,
		Timestamp:    time.Now(),
	}

	body, err := json.Marshal(msg)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	err = s.channel.PublishWithContext(ctx,
		"",                      // exchange
		s.config.EmailQueueName, // routing key
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

func (s *messagingService) ProcessMessages(ctx context.Context) error {
	msgs, err := s.channel.Consume(
		s.config.EmailQueueName, // queue
		"",                      // consumer
		false,                   // auto-ack
		false,                   // exclusive
		false,                   // no-local
		false,                   // no-wait
		nil,                     // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			processCtx, span := s.tracer.Start(ctx, "ProcessMessage")

			var emailMsg EmailMessage
			if err := json.Unmarshal(msg.Body, &emailMsg); err != nil {
				span.RecordError(err)
				if err := s.moveToDeadLetter(processCtx, msg); err != nil {
					s.logger.Error("Failed to move message to dead letter queue", zap.Error(err))
				}
				span.End()
				continue
			}

			span.SetAttributes(
				attribute.String("user_id", emailMsg.UserID),
				attribute.Int("retry_count", emailMsg.RetryCount),
			)

			_, err := s.calendar.ProcessEmailToCalendar(processCtx, emailMsg.EmailContent) // Updated to match interface
			if err != nil {
				span.RecordError(err)
				if emailMsg.RetryCount < s.config.MaxRetries {
					if err := s.retryMessage(processCtx, emailMsg); err != nil {
						s.logger.Error("Failed to retry message", zap.Error(err))
					}
				} else {
					if err := s.moveToDeadLetter(processCtx, msg); err != nil {
						s.logger.Error("Failed to move message to dead letter queue", zap.Error(err))
					}
				}
			} else {
				if err := msg.Ack(false); err != nil {
					s.logger.Error("Failed to acknowledge message", zap.Error(err))
				}
			}

			span.End()
		}
	}()

	return nil
}

func (s *messagingService) Close() error {
	if err := s.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %v", err)
	}
	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}
	return nil
}

func (s *messagingService) retryMessage(ctx context.Context, msg EmailMessage) error {
	msg.RetryCount++
	msg.Timestamp = time.Now()

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Publish with delay
	time.Sleep(time.Duration(s.config.RetryDelaySeconds) * time.Second)
	return s.channel.PublishWithContext(ctx,
		"",
		s.config.EmailQueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (s *messagingService) moveToDeadLetter(ctx context.Context, msg amqp.Delivery) error {
	return s.channel.PublishWithContext(ctx,
		"",
		s.config.DeadLetterQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: msg.ContentType,
			Body:        msg.Body,
			Headers:     msg.Headers,
		},
	)
}

func declareQueues(ch *amqp.Channel, config QueueConfig) error {
	// Declare main queue
	_, err := ch.QueueDeclare(
		config.EmailQueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// Declare dead letter queue
	_, err = ch.QueueDeclare(
		config.DeadLetterQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}
