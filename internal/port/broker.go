package port

import (
	"context"
	"github.com/dontpanicw/EventBooker/internal/domain"
)

type Broker interface {
	PublishDelayedCancellation(ctx context.Context, booking *domain.Booking) error
}
