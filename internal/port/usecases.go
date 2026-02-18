package port

import (
	"context"
	"github.com/dontpanicw/EventBooker/internal/domain"
)

type Usecases interface {
	CreateEvent(ctx context.Context, event *domain.Event) (string, error)
	BookEvent(ctx context.Context, booking *domain.Booking) (string, error)
	ConfirmBooking(ctx context.Context, bookingID string) error
	GetEvent(ctx context.Context, eventID string) (*domain.Event, error)
	GetAllEvents(ctx context.Context) ([]*domain.Event, error)
	GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error)
}
