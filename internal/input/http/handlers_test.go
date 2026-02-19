package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dontpanicw/EventBooker/internal/domain"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUsecases - мок для usecases
type MockUsecases struct {
	mock.Mock
}

func (m *MockUsecases) CreateEvent(ctx context.Context, event *domain.Event) (string, error) {
	args := m.Called(ctx, event)
	return args.String(0), args.Error(1)
}

func (m *MockUsecases) BookEvent(ctx context.Context, booking *domain.Booking) (string, error) {
	args := m.Called(ctx, booking)
	return args.String(0), args.Error(1)
}

func (m *MockUsecases) ConfirmBooking(ctx context.Context, bookingID string) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

func (m *MockUsecases) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *MockUsecases) GetAllEvents(ctx context.Context) ([]*domain.Event, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *MockUsecases) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func TestCreateEvent_Success(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	reqBody := CreateEventRequest{
		Name:             "Test Event",
		Description:      "Test Description",
		IsFree:           false,
		Price:            100.0,
		AvailableTickets: 50,
		Date:             time.Now(),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockUsecases.On("CreateEvent", mock.Anything, mock.AnythingOfType("*domain.Event")).Return("event-123", nil)

	handler.CreateEvent(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "event-123", response["event_id"])
	mockUsecases.AssertExpectations(t)
}

func TestCreateEvent_InvalidJSON(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateEvent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEvent_UsecaseError(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	reqBody := CreateEventRequest{
		Name:             "Test Event",
		AvailableTickets: 50,
		Date:             time.Now(),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockUsecases.On("CreateEvent", mock.Anything, mock.AnythingOfType("*domain.Event")).Return("", errors.New("database error"))

	handler.CreateEvent(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUsecases.AssertExpectations(t)
}

func TestBookEvent_Success(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	reqBody := BookEventRequest{
		UserId: "user-123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/events/event-123/book", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "event-123"})
	w := httptest.NewRecorder()

	mockUsecases.On("BookEvent", mock.Anything, mock.AnythingOfType("*domain.Booking")).Return("booking-123", nil)

	handler.BookEvent(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "booking-123", response["booking_id"])
	assert.Equal(t, "event-123", response["event_id"])
	mockUsecases.AssertExpectations(t)
}

func TestBookEvent_InvalidJSON(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	req := httptest.NewRequest(http.MethodPost, "/api/events/event-123/book", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "event-123"})
	w := httptest.NewRecorder()

	handler.BookEvent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfirmBooking_Success(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	req := httptest.NewRequest(http.MethodPost, "/api/bookings/booking-123/confirm", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "booking-123"})
	w := httptest.NewRecorder()

	mockUsecases.On("ConfirmBooking", mock.Anything, "booking-123").Return(nil)

	handler.ConfirmBooking(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "confirmed", response["status"])
	mockUsecases.AssertExpectations(t)
}

func TestConfirmBooking_Error(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	req := httptest.NewRequest(http.MethodPost, "/api/bookings/booking-123/confirm", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "booking-123"})
	w := httptest.NewRecorder()

	mockUsecases.On("ConfirmBooking", mock.Anything, "booking-123").Return(errors.New("booking not found"))

	handler.ConfirmBooking(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUsecases.AssertExpectations(t)
}

func TestGetEvent_Success(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	expectedEvent := &domain.Event{
		Id:               "event-123",
		Name:             "Test Event",
		AvailableTickets: 50,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/events/event-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "event-123"})
	w := httptest.NewRecorder()

	mockUsecases.On("GetEvent", mock.Anything, "event-123").Return(expectedEvent, nil)

	handler.GetEvent(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var event domain.Event
	json.Unmarshal(w.Body.Bytes(), &event)
	assert.Equal(t, "event-123", event.Id)
	assert.Equal(t, "Test Event", event.Name)
	mockUsecases.AssertExpectations(t)
}

func TestGetEvent_NotFound(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	req := httptest.NewRequest(http.MethodGet, "/api/events/event-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "event-123"})
	w := httptest.NewRecorder()

	mockUsecases.On("GetEvent", mock.Anything, "event-123").Return(nil, errors.New("not found"))

	handler.GetEvent(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUsecases.AssertExpectations(t)
}

func TestGetAllEvents_Success(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	expectedEvents := []*domain.Event{
		{Id: "event-1", Name: "Event 1"},
		{Id: "event-2", Name: "Event 2"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	w := httptest.NewRecorder()

	mockUsecases.On("GetAllEvents", mock.Anything).Return(expectedEvents, nil)

	handler.GetAllEvents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var events []*domain.Event
	json.Unmarshal(w.Body.Bytes(), &events)
	assert.Len(t, events, 2)
	mockUsecases.AssertExpectations(t)
}

func TestGetBooking_Success(t *testing.T) {
	mockUsecases := new(MockUsecases)
	handler := NewHandler(mockUsecases)

	expectedBooking := &domain.Booking{
		Id:      "booking-123",
		UserId:  "user-123",
		EventId: "event-123",
		Status:  domain.PendingStatus,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/bookings/booking-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "booking-123"})
	w := httptest.NewRecorder()

	mockUsecases.On("GetBooking", mock.Anything, "booking-123").Return(expectedBooking, nil)

	handler.GetBooking(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var booking domain.Booking
	json.Unmarshal(w.Body.Bytes(), &booking)
	assert.Equal(t, "booking-123", booking.Id)
	assert.Equal(t, domain.PendingStatus, booking.Status)
	mockUsecases.AssertExpectations(t)
}
