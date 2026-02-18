package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dontpanicw/EventBooker/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DelayedCancellationsQueue = "delayed_cancellations"
	WaitingQueue              = "waiting_cancellations"
	ConfirmationsQueue        = "confirmations"
	DelayedExchange           = "delayed_exchange"
	WaitingExchange           = "waiting_exchange"
	ConfirmationsExchange     = "confirmations_exchange"
	BookingTTL                = 1 * 60 * 1000 // 15 минут в миллисекундах
)

type RabbitMQBroker struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type BookingMessage struct {
	BookingID string    `json:"booking_id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

func NewRabbitMQBroker(url string) (*RabbitMQBroker, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	broker := &RabbitMQBroker{
		conn:    conn,
		channel: channel,
	}

	if err := broker.setupQueues(); err != nil {
		broker.Close()
		return nil, fmt.Errorf("failed to setup queues: %w", err)
	}

	return broker, nil
}

func (b *RabbitMQBroker) setupQueues() error {
	// Declare delayed exchange (куда попадут сообщения после TTL)
	err := b.channel.ExchangeDeclare(
		DelayedExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare delayed exchange: %w", err)
	}

	// Declare waiting exchange (куда публикуем изначально)
	err = b.channel.ExchangeDeclare(
		WaitingExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare waiting exchange: %w", err)
	}

	// Declare confirmations exchange
	err = b.channel.ExchangeDeclare(
		ConfirmationsExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare confirmations exchange: %w", err)
	}

	// Declare waiting queue с TTL и DLX (сообщения "спят" здесь 15 минут)
	waitingArgs := amqp.Table{
		"x-message-ttl":             BookingTTL,
		"x-dead-letter-exchange":    DelayedExchange,
		"x-dead-letter-routing-key": DelayedCancellationsQueue,
	}
	_, err = b.channel.QueueDeclare(
		WaitingQueue,
		true,
		false,
		false,
		false,
		waitingArgs,
	)
	if err != nil {
		return fmt.Errorf("failed to declare waiting queue: %w", err)
	}

	// Bind waiting queue to waiting exchange
	err = b.channel.QueueBind(
		WaitingQueue,
		WaitingQueue,
		WaitingExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind waiting queue: %w", err)
	}

	// Declare delayed cancellations queue (сюда попадают сообщения после TTL)
	_, err = b.channel.QueueDeclare(
		DelayedCancellationsQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare delayed cancellations queue: %w", err)
	}

	// Bind delayed queue to delayed exchange
	err = b.channel.QueueBind(
		DelayedCancellationsQueue,
		DelayedCancellationsQueue,
		DelayedExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind delayed queue: %w", err)
	}

	// Declare confirmations queue
	//_, err = b.channel.QueueDeclare(
	//	ConfirmationsQueue,
	//	true,
	//	false,
	//	false,
	//	false,
	//	nil,
	//)
	//if err != nil {
	//	return fmt.Errorf("failed to declare confirmations queue: %w", err)
	//}
	//
	//// Bind confirmations queue to exchange
	//err = b.channel.QueueBind(
	//	ConfirmationsQueue,
	//	ConfirmationsQueue,
	//	ConfirmationsExchange,
	//	false,
	//	nil,
	//)
	//if err != nil {
	//	return fmt.Errorf("failed to bind confirmations queue: %w", err)
	//}

	return nil
}

func (b *RabbitMQBroker) PublishDelayedCancellation(ctx context.Context, booking *domain.Booking) error {
	msg := BookingMessage{
		BookingID: booking.Id,
		EventID:   booking.EventId,
		UserID:    booking.UserId,
		Timestamp: booking.Date,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Публикуем в waiting queue, откуда сообщение попадет в delayed queue через 15 минут
	err = b.channel.PublishWithContext(
		ctx,
		WaitingExchange,
		WaitingQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish delayed cancellation: %w", err)
	}

	log.Printf("Published delayed cancellation for booking %s (will be processed in 15 minutes)", booking.Id)
	return nil
}

//func (b *RabbitMQBroker) PublishConfirmation(ctx context.Context, bookingID, eventID string) error {
//	msg := BookingMessage{
//		BookingID: bookingID,
//		EventID:   eventID,
//		Timestamp: time.Now(),
//	}
//
//	body, err := json.Marshal(msg)
//	if err != nil {
//		return fmt.Errorf("failed to marshal message: %w", err)
//	}
//
//	err = b.channel.PublishWithContext(
//		ctx,
//		ConfirmationsExchange,
//		ConfirmationsQueue,
//		false,
//		false,
//		amqp.Publishing{
//			ContentType: "application/json",
//			Body:        body,
//		},
//	)
//	if err != nil {
//		return fmt.Errorf("failed to publish confirmation: %w", err)
//	}
//
//	log.Printf("Published confirmation for booking %s", bookingID)
//	return nil
//}

func (b *RabbitMQBroker) Close() error {
	if b.channel != nil {
		b.channel.Close()
	}
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *RabbitMQBroker) GetChannel() *amqp.Channel {
	return b.channel
}
