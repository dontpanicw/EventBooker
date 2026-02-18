package port

import (
	"context"
	"github.com/dontpanicw/EventBooker/internal/domain"
)

type Repository interface {
	CreateEvent(ctx context.Context, event *domain.Event) (string, error)
	BookEvent(ctx context.Context, booking *domain.Booking) (string, error)
	ConfirmBooking(ctx context.Context, bookingID string) error
	GetEvent(ctx context.Context, eventID string) (*domain.Event, error)
	GetAllEvents(ctx context.Context) ([]*domain.Event, error)
	GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error)
	CancelBooking(ctx context.Context, bookingID string) error
	IncrementAvailableTickets(ctx context.Context, eventID string) error
	AddAvailableTickets(ctx context.Context, eventID string) error
}

//встроенные HTTP-методы:
//– POST /events — создание мероприятия;
//– POST /events/{id}/book — бронирование места;
//– POST /events/{id}/confirm — оплата брони (если мероприятие требует этого);
//– GET /events/{id} — получение информации о мероприятии и свободных местах.
