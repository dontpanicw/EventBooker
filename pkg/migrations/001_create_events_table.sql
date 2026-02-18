-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_free BOOLEAN NOT NULL DEFAULT FALSE,
    price DECIMAL(10, 2),
    available_tickets INT,
    date TIMESTAMP NOT NULL,

    CONSTRAINT check_price_for_free CHECK (
        (is_free = TRUE AND (price IS NULL OR price = 0))
        OR
        (is_free = FALSE AND price > 0)
    )
);

CREATE TABLE bookings (
      id VARCHAR(36) PRIMARY KEY,
      user_id VARCHAR(36) NOT NULL,
      event_id VARCHAR(36) NOT NULL,
      status VARCHAR(50) NOT NULL,
      date TIMESTAMP NOT NULL,

-- Внешние ключи для связи с другими таблицами
      CONSTRAINT fk_bookings_event
          FOREIGN KEY (event_id)
              REFERENCES events(id)
              ON DELETE RESTRICT
              ON UPDATE CASCADE
);
