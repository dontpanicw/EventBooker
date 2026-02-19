package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dontpanicw/EventBooker/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository - мок репозитория
type MockRepository struct {
	mock.Mock
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

func (m *MockRepository) AddAvailableTickets(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

// MockRabbitMQBroker - мок RabbitMQ брокера
type MockRabbitMQBroker struct {
	mock.Mock
}

func (m *MockRabbitMQBroker) PublishDelayedCancellation(ctx context.Context, booking *domain.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

//func (m *MockRabbitMQBroker) PublishConfirmation(ctx context.Context, bookingID, eventID string) error {
//	args := m.Called(ctx, bookingID, eventID)
//	return args.Error(0)
//}

func (m *MockRabbitMQBroker) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestCreateEvent_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	event := &domain.Event{
		Name:             "Test Event",
		Description:      "Test Description",
		IsFree:           false,
		Price:            100.0,
		AvailableTickets: 50,
	}

	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.Event")).Return("event-123", nil)

	eventID, err := usecase.CreateEvent(ctx, event)

	assert.NoError(t, err)
	assert.NotEmpty(t, eventID)
	assert.NotEmpty(t, event.Id)
	mockRepo.AssertExpectations(t)
}

func TestCreateEvent_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	event := &domain.Event{
		Name:             "Test Event",
		Description:      "Test Description",
		AvailableTickets: 50,
	}

	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.Event")).Return("", errors.New("database error"))

	eventID, err := usecase.CreateEvent(ctx, event)

	assert.Error(t, err)
	assert.Empty(t, eventID)
	assert.Contains(t, err.Error(), "failed to create event")
	mockRepo.AssertExpectations(t)
}

func TestBookEvent_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	booking := &domain.Booking{
		UserId:  "user-123",
		EventId: "event-123",
	}

	mockRepo.On("BookEvent", ctx, mock.AnythingOfType("*domain.Booking")).Return("booking-123", nil)
	mockBroker.On("PublishDelayedCancellation", ctx, mock.AnythingOfType("*domain.Booking")).Return(nil)

	bookingID, err := usecase.BookEvent(ctx, booking)

	assert.NoError(t, err)
	assert.NotEmpty(t, bookingID)
	assert.Equal(t, domain.PendingStatus, booking.Status)
	assert.NotEmpty(t, booking.Id)
	mockRepo.AssertExpectations(t)
	mockBroker.AssertExpectations(t)
}

func TestBookEvent_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	booking := &domain.Booking{
		UserId:  "user-123",
		EventId: "event-123",
	}

	mockRepo.On("BookEvent", ctx, mock.AnythingOfType("*domain.Booking")).Return("", errors.New("no tickets available"))

	bookingID, err := usecase.BookEvent(ctx, booking)

	assert.Error(t, err)
	assert.Empty(t, bookingID)
	assert.Contains(t, err.Error(), "failed to book event")
	mockRepo.AssertExpectations(t)
}

func TestBookEvent_BrokerError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	booking := &domain.Booking{
		UserId:  "user-123",
		EventId: "event-123",
	}

	mockRepo.On("BookEvent", ctx, mock.AnythingOfType("*domain.Booking")).Return("booking-123", nil)
	mockBroker.On("PublishDelayedCancellation", ctx, mock.AnythingOfType("*domain.Booking")).Return(errors.New("rabbitmq error"))

	bookingID, err := usecase.BookEvent(ctx, booking)

	assert.Error(t, err)
	assert.Empty(t, bookingID)
	assert.Contains(t, err.Error(), "failed to publish delayed cancellation")
	mockRepo.AssertExpectations(t)
	mockBroker.AssertExpectations(t)
}

func TestConfirmBooking_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	bookingID := "booking-123"
	eventID := "event-123"

	booking := &domain.Booking{
		Id:      bookingID,
		EventId: eventID,
		Status:  domain.PendingStatus,
	}

	mockRepo.On("GetBooking", ctx, bookingID).Return(booking, nil)
	mockRepo.On("ConfirmBooking", ctx, bookingID).Return(nil)

	err := usecase.ConfirmBooking(ctx, bookingID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockBroker.AssertExpectations(t)
}

func TestConfirmBooking_GetBookingError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	bookingID := "booking-123"

	mockRepo.On("GetBooking", ctx, bookingID).Return(nil, errors.New("booking not found"))

	err := usecase.ConfirmBooking(ctx, bookingID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get booking")
	mockRepo.AssertExpectations(t)
}

func TestGetEvent_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	eventID := "event-123"
	expectedEvent := &domain.Event{
		Id:               eventID,
		Name:             "Test Event",
		AvailableTickets: 50,
	}

	mockRepo.On("GetEvent", ctx, eventID).Return(expectedEvent, nil)

	event, err := usecase.GetEvent(ctx, eventID)

	assert.NoError(t, err)
	assert.Equal(t, expectedEvent, event)
	mockRepo.AssertExpectations(t)
}

func TestGetAllEvents_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	expectedEvents := []*domain.Event{
		{Id: "event-1", Name: "Event 1"},
		{Id: "event-2", Name: "Event 2"},
	}

	mockRepo.On("GetAllEvents", ctx).Return(expectedEvents, nil)

	events, err := usecase.GetAllEvents(ctx)

	assert.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, expectedEvents, events)
	mockRepo.AssertExpectations(t)
}

func TestGetBooking_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockBroker := new(MockRabbitMQBroker)

	usecase := &EventsUsecases{
		repo:   mockRepo,
		broker: mockBroker,
	}

	ctx := context.Background()
	bookingID := "booking-123"
	expectedBooking := &domain.Booking{
		Id:      bookingID,
		UserId:  "user-123",
		EventId: "event-123",
		Status:  domain.PendingStatus,
		Date:    time.Now(),
	}

	mockRepo.On("GetBooking", ctx, bookingID).Return(expectedBooking, nil)

	booking, err := usecase.GetBooking(ctx, bookingID)

	assert.NoError(t, err)
	assert.Equal(t, expectedBooking, booking)
	mockRepo.AssertExpectations(t)
}
