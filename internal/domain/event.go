package domain

import "time"

const (
	PendingStatus   = "pending"
	ConfirmedStatus = "confirmed"
	CancelledStatus = "cancelled"
)

type Event struct {
	Id               string
	Name             string
	Description      string
	IsFree           bool
	Price            float64
	AvailableTickets uint32
	Date             time.Time
}

type Booking struct {
	Id      string
	UserId  string
	EventId string
	Status  string
	Date    time.Time
}

// TaskMessage - структура сообщения для Kafka
type TaskMessage struct {
	EventID  string    `json:"image_id"`
	Deadline time.Time `json:"deadline"`
}
