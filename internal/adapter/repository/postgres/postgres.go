package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dontpanicw/EventBooker/config"
	"github.com/dontpanicw/EventBooker/internal/domain"
	"github.com/dontpanicw/EventBooker/internal/port"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"log"
	"time"
)

const (
	createEventQuery   = `INSERT INTO events (id, name, description, is_free, price, available_tickets, date) VALUES ($1, $2, $3, $4, $5, $6, $7);`
	getEventQuery      = `SELECT * FROM events WHERE id = $1;`
	getAllEventsQuery  = `SELECT id, name, description, is_free, price, available_tickets, date FROM events ORDER BY date ASC;`
	bookEventQuery     = `INSERT INTO bookings (id, user_id, event_id, status, date) VALUES ($1, $2, $3, $4, $5);`
	confirmBookQuery   = `UPDATE bookings SET status = $1 WHERE id = $2;`
	getBookingQuery    = `SELECT id, user_id, event_id, status, date FROM bookings WHERE id = $1;`
	cancelBookingQuery = `UPDATE bookings SET status = $1 WHERE id = $2;`
	updateEventQuery   = `UPDATE events 
						SET available_tickets = available_tickets - 1 
						WHERE id = $1 AND available_tickets > 0
						RETURNING available_tickets;`
	addAvailableTicketQuery = `UPDATE events
							   SET available_tickets = available_tickets + 1
							   WHERE id = $1
							   RETURNING available_tickets;`
)

type EventRepository struct {
	PostgresDB *dbpg.DB
}

func NewEventRepository(cfg *config.Config) port.Repository {
	opts := &dbpg.Options{MaxOpenConns: 10, MaxIdleConns: 5}
	db, err := dbpg.New(cfg.MasterDSN, cfg.SlaveDSNs, opts)
	if err != nil {
		panic(err)
	}

	return &EventRepository{
		PostgresDB: db,
	}
}

func (e *EventRepository) CreateEvent(ctx context.Context, event *domain.Event) (string, error) {
	_, err := e.PostgresDB.ExecWithRetry(ctx, createRetryStrategy(), createEventQuery,
		event.Id,
		event.Name,
		event.Description,
		event.IsFree,
		event.Price,
		event.AvailableTickets,
		event.Date,
	)
	if err != nil {
		return "", fmt.Errorf("error create event: %w", err)
	}

	return event.Id, nil
}

func (e *EventRepository) BookEvent(ctx context.Context, booking *domain.Booking) (string, error) {
	tx, err := e.PostgresDB.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}() // Будет отменен, если не закоммитим

	var newAvailableTickets int

	err = e.PostgresDB.QueryRowContext(ctx, updateEventQuery, booking.EventId).Scan(&newAvailableTickets)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Билетов нет или событие не найдено
			log.Printf("No tickets available or event not found for ID: %s", booking.EventId)
			return "", fmt.Errorf("no tickets available")
		}
		return "", fmt.Errorf("failed to update tickets: %w", err)
	}
	// Успешное обновление, newAvailableTickets содержит новое количество билетов
	log.Printf("Tickets updated successfully. Remaining tickets: %d", newAvailableTickets)

	if newAvailableTickets == 0 {
		log.Printf("Event %s is now sold out", booking.EventId)
	}

	_, err = e.PostgresDB.ExecWithRetry(ctx, createRetryStrategy(), bookEventQuery,
		booking.Id,
		booking.UserId,
		booking.EventId,
		booking.Status,
		booking.Date,
	)
	if err != nil {
		return "", fmt.Errorf("failed to book event: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Printf("Ticket booked successfully. Booking ID: %s, Remaining tickets: %d",
		booking.Id, newAvailableTickets)

	return booking.Id, nil
}

func (e *EventRepository) ConfirmBooking(ctx context.Context, bookingID string) error {
	log.Printf("Confirming booking %s", bookingID)
	_, err := e.PostgresDB.ExecWithRetry(ctx, createRetryStrategy(), confirmBookQuery,
		domain.ConfirmedStatus,
		bookingID,
	)
	if err != nil {
		return fmt.Errorf("error confirm booking: %w", err)
	}
	log.Printf("Confirmed booking %s", bookingID)
	return nil
}

func (e *EventRepository) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	var event domain.Event
	err := e.PostgresDB.QueryRowContext(ctx, getEventQuery, eventID).Scan(&event.Id, &event.Name, &event.Description, &event.IsFree, &event.Price, &event.AvailableTickets, &event.Date)
	if err != nil {
		return nil, fmt.Errorf("error get event: %w", err)
	}
	return &event, nil
}

func (e *EventRepository) GetAllEvents(ctx context.Context) ([]*domain.Event, error) {
	rows, err := e.PostgresDB.QueryContext(ctx, getAllEventsQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying all events: %w", err)
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var events []*domain.Event
	for rows.Next() {
		var event domain.Event
		err := rows.Scan(&event.Id, &event.Name, &event.Description, &event.IsFree, &event.Price, &event.AvailableTickets, &event.Date)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %w", err)
		}
		events = append(events, &event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

func (e *EventRepository) AddAvailableTickets(ctx context.Context, eventID string) error {
	var newAvailableTickets int

	err := e.PostgresDB.QueryRowContext(ctx, addAvailableTicketQuery, eventID).Scan(&newAvailableTickets)

	if err != nil {
		return fmt.Errorf("failed to update tickets: %w", err)
	}
	// Успешное обновление, newAvailableTickets содержит новое количество билетов
	log.Printf("Tickets updated successfully. Remaining tickets: %d", newAvailableTickets)

	return nil
}

func createRetryStrategy() retry.Strategy {
	return retry.Strategy{
		Attempts: 3,
		Delay:    5 * time.Second,
		Backoff:  2}
}

func (e *EventRepository) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	var booking domain.Booking
	err := e.PostgresDB.QueryRowContext(ctx, getBookingQuery, bookingID).Scan(
		&booking.Id,
		&booking.UserId,
		&booking.EventId,
		&booking.Status,
		&booking.Date,
	)
	if err != nil {
		return nil, fmt.Errorf("error get booking: %w", err)
	}
	return &booking, nil
}

func (e *EventRepository) CancelBooking(ctx context.Context, bookingID string) error {
	_, err := e.PostgresDB.ExecWithRetry(ctx, createRetryStrategy(), cancelBookingQuery,
		domain.CancelledStatus,
		bookingID,
	)
	if err != nil {
		return fmt.Errorf("error cancel booking: %w", err)
	}
	return nil
}

func (e *EventRepository) IncrementAvailableTickets(ctx context.Context, eventID string) error {
	var newAvailableTickets int
	err := e.PostgresDB.QueryRowContext(ctx, addAvailableTicketQuery, eventID).Scan(&newAvailableTickets)
	if err != nil {
		return fmt.Errorf("failed to increment tickets: %w", err)
	}
	log.Printf("Tickets incremented successfully. Available tickets: %d", newAvailableTickets)
	return nil
}
