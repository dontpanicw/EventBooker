package consumer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dontpanicw/EventBooker/internal/adapter/broker"
	"github.com/dontpanicw/EventBooker/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository для consumer тестов
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockRepository) CancelBooking(ctx context.Context, bookingID string) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

func (m *MockRepository) IncrementAvailableTickets(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func (m *MockRepository) CreateEvent(ctx context.Context, event *domain.Event) (string, error) {
	args := m.Called(ctx, event)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) BookEvent(ctx context.Context, booking *domain.Booking) (string, error) {
	args := m.Called(ctx, booking)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) ConfirmBooking(ctx context.Context, bookingID string) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

func (m *MockRepository) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *MockRepository) GetAllEvents(ctx context.Context) ([]*domain.Event, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *MockRepository) AddAvailableTickets(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

// MockDelivery для тестирования обработки сообщений
type MockDelivery struct {
	body        []byte
	acked       bool
	nacked      bool
	requeued    bool
}

func (m *MockDelivery) Ack(multiple bool) error {
	m.acked = true
	return nil
}

func (m *MockDelivery) Nack(multiple, requeue bool) error {
	m.nacked = true
	m.requeued = requeue
	return nil
}

func TestCancellationConsumer_PendingBooking(t *testing.T) {
	mockRepo := new(MockRepository)
	consumer := &CancellationConsumer{
		repo: mockRepo,
	}

	ctx := context.Background()
	bookingMsg := broker.BookingMessage{
		BookingID: "booking-123",
		EventID:   "event-123",
		UserID:    "user-123",
		Timestamp: time.Now(),
	}

	body, _ := json.Marshal(bookingMsg)
	
	// Создаем реальный Delivery
	delivery := amqp.Delivery{
		Body: body,
	}

	booking := &domain.Booking{
		Id:      "booking-123",
		EventId: "event-123",
		Status:  domain.PendingStatus,
	}

	mockRepo.On("GetBooking", ctx, "booking-123").Return(booking, nil)
	mockRepo.On("CancelBooking", ctx, "booking-123").Return(nil)
	mockRepo.On("IncrementAvailableTickets", ctx, "event-123").Return(nil)

	consumer.handleMessage(ctx, delivery)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertCalled(t, "CancelBooking", ctx, "booking-123")
	mockRepo.AssertCalled(t, "IncrementAvailableTickets", ctx, "event-123")
}

func TestCancellationConsumer_ConfirmedBooking(t *testing.T) {
	mockRepo := new(MockRepository)
	consumer := &CancellationConsumer{
		repo: mockRepo,
	}

	ctx := context.Background()
	bookingMsg := broker.BookingMessage{
		BookingID: "booking-123",
		EventID:   "event-123",
		UserID:    "user-123",
		Timestamp: time.Now(),
	}

	body, _ := json.Marshal(bookingMsg)
	
	delivery := amqp.Delivery{
		Body: body,
	}

	booking := &domain.Booking{
		Id:      "booking-123",
		EventId: "event-123",
		Status:  domain.ConfirmedStatus,
	}

	mockRepo.On("GetBooking", ctx, "booking-123").Return(booking, nil)

	consumer.handleMessage(ctx, delivery)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "CancelBooking", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "IncrementAvailableTickets", mock.Anything, mock.Anything)
}

func TestBookingMessage_Marshaling(t *testing.T) {
	msg := broker.BookingMessage{
		BookingID: "booking-123",
		EventID:   "event-123",
		UserID:    "user-123",
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded broker.BookingMessage
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, msg.BookingID, decoded.BookingID)
	assert.Equal(t, msg.EventID, decoded.EventID)
	assert.Equal(t, msg.UserID, decoded.UserID)
}
