package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventStatuses(t *testing.T) {
	assert.Equal(t, "pending", PendingStatus)
	assert.Equal(t, "confirmed", ConfirmedStatus)
	assert.Equal(t, "cancelled", CancelledStatus)
}

func TestEventCreation(t *testing.T) {
	event := Event{
		Id:               "event-123",
		Name:             "Test Event",
		Description:      "Test Description",
		IsFree:           false,
		Price:            100.0,
		AvailableTickets: 50,
		Date:             time.Now(),
	}

	assert.Equal(t, "event-123", event.Id)
	assert.Equal(t, "Test Event", event.Name)
	assert.Equal(t, "Test Description", event.Description)
	assert.False(t, event.IsFree)
	assert.Equal(t, 100.0, event.Price)
	assert.Equal(t, uint32(50), event.AvailableTickets)
	assert.NotZero(t, event.Date)
}

func TestFreeEvent(t *testing.T) {
	event := Event{
		Id:               "event-123",
		Name:             "Free Event",
		IsFree:           true,
		Price:            0,
		AvailableTickets: 100,
		Date:             time.Now(),
	}

	assert.True(t, event.IsFree)
	assert.Equal(t, 0.0, event.Price)
}

func TestBookingCreation(t *testing.T) {
	booking := Booking{
		Id:      "booking-123",
		UserId:  "user-123",
		EventId: "event-123",
		Status:  PendingStatus,
		Date:    time.Now(),
	}

	assert.Equal(t, "booking-123", booking.Id)
	assert.Equal(t, "user-123", booking.UserId)
	assert.Equal(t, "event-123", booking.EventId)
	assert.Equal(t, PendingStatus, booking.Status)
	assert.NotZero(t, booking.Date)
}

func TestBookingStatusTransitions(t *testing.T) {
	booking := Booking{
		Id:      "booking-123",
		UserId:  "user-123",
		EventId: "event-123",
		Status:  PendingStatus,
		Date:    time.Now(),
	}

	// Pending -> Confirmed
	booking.Status = ConfirmedStatus
	assert.Equal(t, ConfirmedStatus, booking.Status)

	// Reset to pending
	booking.Status = PendingStatus

	// Pending -> Cancelled
	booking.Status = CancelledStatus
	assert.Equal(t, CancelledStatus, booking.Status)
}
