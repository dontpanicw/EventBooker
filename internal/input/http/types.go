package http

import "time"

type CreateEventRequest struct {
	UserId           string    `json:"user_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	IsFree           bool      `json:"is_free"`
	Price            float64   `json:"price"`
	AvailableTickets uint32    `json:"available_tickets"`
	Date             time.Time `json:"date"`
}

type BookEventRequest struct {
	UserId string `json:"user_id"`
}
