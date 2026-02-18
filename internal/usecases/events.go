package usecases

import (
	"context"
	"fmt"
	"github.com/dontpanicw/EventBooker/internal/adapter/broker"
	"github.com/dontpanicw/EventBooker/internal/domain"
	"github.com/dontpanicw/EventBooker/internal/port"
	"github.com/google/uuid"
	"time"
)

type EventsUsecases struct {
	repo   port.Repository
	broker *broker.RabbitMQBroker
}

func NewEventsUsecases(repo port.Repository, rabbitBroker *broker.RabbitMQBroker) port.Usecases {
	return &EventsUsecases{
		repo:   repo,
		broker: rabbitBroker,
	}
}

func (e *EventsUsecases) CreateEvent(ctx context.Context, event *domain.Event) (string, error) {
	id := uuid.New().String()
	event.Id = id
	date := time.Now()
	event.Date = date

	id, err := e.repo.CreateEvent(ctx, event)
	if err != nil {
		return "", fmt.Errorf("failed to create event: %w", err)
	}
	return id, nil
}

func (e *EventsUsecases) BookEvent(ctx context.Context, booking *domain.Booking) (string, error) {
	id := uuid.New().String()
	booking.Id = id
	booking.Date = time.Now()
	booking.Status = domain.PendingStatus

	id, err := e.repo.BookEvent(ctx, booking)
	if err != nil {
		return "", fmt.Errorf("failed to book event: %w", err)
	}

	// Публикуем сообщение в очередь отложенных отмен
	if err := e.broker.PublishDelayedCancellation(ctx, booking); err != nil {
		return "", fmt.Errorf("failed to publish delayed cancellation: %w", err)
	}

	return id, nil
}

func (e *EventsUsecases) ConfirmEvent(ctx context.Context, eventID string) error {
	err := e.repo.ConfirmBooking(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to confirm event: %w", err)
	}

	// Публикуем сообщение о подтверждении
	//if err := e.broker.PublishConfirmation(ctx, eventID, eventID); err != nil {
	//	return fmt.Errorf("failed to publish confirmation: %w", err)
	//}

	return nil
}

func (e *EventsUsecases) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	event, err := e.repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm event: %w", err)
	}
	return event, nil
}

func (e *EventsUsecases) GetAllEvents(ctx context.Context) ([]*domain.Event, error) {
	events, err := e.repo.GetAllEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all events: %w", err)
	}
	return events, nil
}

func (e *EventsUsecases) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	booking, err := e.repo.GetBooking(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	return booking, nil
}

func (e *EventsUsecases) ConfirmBooking(ctx context.Context, bookingID string) error {
	// Сначала получаем бронь, чтобы узнать eventID
	_, err := e.repo.GetBooking(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}

	// Обновляем статус брони
	err = e.repo.ConfirmBooking(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("failed to confirm booking: %w", err)
	}
	//
	//// Публикуем сообщение о подтверждении
	//if err := e.broker.PublishConfirmation(ctx, bookingID, booking.EventId); err != nil {
	//	return fmt.Errorf("failed to publish confirmation: %w", err)
	//}

	return nil
}
