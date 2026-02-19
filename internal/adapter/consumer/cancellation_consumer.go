package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/dontpanicw/EventBooker/internal/adapter/broker"
	"github.com/dontpanicw/EventBooker/internal/domain"
	"github.com/dontpanicw/EventBooker/internal/port"
	amqp "github.com/rabbitmq/amqp091-go"
)

type CancellationConsumer struct {
	channel *amqp.Channel
	repo    port.Repository
}

func NewCancellationConsumer(channel *amqp.Channel, repo port.Repository) *CancellationConsumer {
	return &CancellationConsumer{
		channel: channel,
		repo:    repo,
	}
}

func (c *CancellationConsumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		broker.DelayedCancellationsQueue,
		"cancellation_consumer",
		false, // auto-ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("Cancellation consumer started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Cancellation consumer stopped")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Channel closed")
					return
				}
				c.handleMessage(ctx, msg)
			}
		}
	}()

	return nil
}

func (c *CancellationConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var bookingMsg broker.BookingMessage
	if err := json.Unmarshal(msg.Body, &bookingMsg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false)
		return
	}

	log.Printf("Processing cancellation check for booking %s", bookingMsg.BookingID)

	// Получаем бронь из БД
	booking, err := c.repo.GetBooking(ctx, bookingMsg.BookingID)
	if err != nil {
		log.Printf("Failed to get booking %s: %v", bookingMsg.BookingID, err)
		msg.Nack(false, true) // Requeue
		return
	}

	// Проверяем статус
	if booking.Status == domain.PendingStatus {
		// Отменяем бронь и возвращаем место
		log.Printf("Cancelling booking %s (not paid in time)", bookingMsg.BookingID)

		if err := c.repo.CancelBooking(ctx, bookingMsg.BookingID); err != nil {
			log.Printf("Failed to cancel booking %s: %v", bookingMsg.BookingID, err)
			msg.Nack(false, true)
			return
		}

		// Возвращаем место
		if err := c.repo.IncrementAvailableTickets(ctx, bookingMsg.EventID); err != nil {
			log.Printf("Failed to increment tickets for event %s: %v", bookingMsg.EventID, err)
			msg.Nack(false, true)
			return
		}

		log.Printf("Booking %s cancelled and ticket returned", bookingMsg.BookingID)
	} else {
		log.Printf("Booking %s already confirmed, skipping cancellation", bookingMsg.BookingID)
	}

	msg.Ack(false)
}
