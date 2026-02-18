package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dontpanicw/EventBooker/config"
	"github.com/dontpanicw/EventBooker/internal/adapter/broker"
	"github.com/dontpanicw/EventBooker/internal/adapter/consumer"
	"github.com/dontpanicw/EventBooker/internal/adapter/repository/postgres"
	"github.com/dontpanicw/EventBooker/internal/input/http"
	"github.com/dontpanicw/EventBooker/internal/usecases"
	"github.com/dontpanicw/EventBooker/pkg/migrations"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func Start(cfg *config.Config) error {
	ctx := context.Background()

	// Retry подключения к PostgreSQL
	var db *sql.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", cfg.MasterDSN)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Waiting for PostgreSQL... (attempt %d/10): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database after 10 attempts: %w", err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()
	log.Print("Connected to PostgreSQL")

	if err := migrations.Migrate(db); err != nil {
		fmt.Println(err)
		return err
	}
	log.Print("Migrations applied successfully")

	// Инициализируем RabbitMQ
	rabbitBroker, err := broker.NewRabbitMQBroker(cfg.RabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to initialize RabbitMQ: %w", err)
	}
	defer rabbitBroker.Close()
	log.Print("RabbitMQ initialized")

	imageRepo := postgres.NewEventRepository(cfg)

	// Запускаем consumers
	cancellationConsumer := consumer.NewCancellationConsumer(rabbitBroker.GetChannel(), imageRepo)
	if err := cancellationConsumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start cancellation consumer: %w", err)
	}
	log.Print("Cancellation consumer started")

	//confirmationConsumer := consumer.NewConfirmationConsumer(rabbitBroker.GetChannel())
	//if err := confirmationConsumer.Start(ctx); err != nil {
	//	return fmt.Errorf("failed to start confirmation consumer: %w", err)
	//}
	//log.Print("Confirmation consumer started")

	imageUsecase := usecases.NewEventsUsecases(imageRepo, rabbitBroker)

	srv := http.NewServer(cfg.HTTPPort, imageUsecase)

	return srv.Start()
}
