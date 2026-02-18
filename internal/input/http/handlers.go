package http

import (
	"encoding/json"
	"github.com/dontpanicw/EventBooker/internal/domain"
	"github.com/dontpanicw/EventBooker/internal/port"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type Handler struct {
	usecases port.Usecases
}

func NewHandler(usecases port.Usecases) *Handler {
	return &Handler{
		usecases: usecases,
	}
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	event := &domain.Event{
		Name:             req.Name,
		Description:      req.Description,
		IsFree:           req.IsFree,
		Price:            req.Price,
		AvailableTickets: req.AvailableTickets,
		Date:             req.Date,
	}

	eventID, err := h.usecases.CreateEvent(r.Context(), event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"event_id": eventID})
}

func (h *Handler) BookEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	var req BookEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	booking := &domain.Booking{
		UserId:  req.UserId,
		EventId: eventID,
		Status:  domain.PendingStatus,
		Date:    time.Now(),
	}

	bookingID, err := h.usecases.BookEvent(r.Context(), booking)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"booking_id": bookingID, "event_id": eventID})
}

func (h *Handler) ConfirmBooking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookingID := vars["id"]

	if err := h.usecases.ConfirmBooking(r.Context(), bookingID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "confirmed"})
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	event, err := h.usecases.GetEvent(r.Context(), eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(event)
}

func (h *Handler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.usecases.GetAllEvents(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}


func (h *Handler) GetBooking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookingID := vars["id"]

	booking, err := h.usecases.GetBooking(r.Context(), bookingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(booking)
}
