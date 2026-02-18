package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	HTTPPort    string
	MasterDSN   string
	SlaveDSNs   []string
	RabbitMQURL string
}

const (
	DefaultHTTPPort      = ":8080"
	DefaultMinioEndpoint = ":9000"
)

func NewConfig() (*Config, error) {
	cfg := Config{}

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		cfg.HTTPPort = DefaultHTTPPort
	} else {
		// Ensure port starts with ':' if not already present
		if len(httpPort) > 0 && httpPort[0] != ':' {
			cfg.HTTPPort = ":" + httpPort
		} else {
			cfg.HTTPPort = httpPort
		}
	}

	masterDSN := os.Getenv("MASTER_DSN")
	if masterDSN != "" {
		cfg.MasterDSN = masterDSN
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL != "" {
		cfg.RabbitMQURL = rabbitMQURL
	} else {
		cfg.RabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	return &cfg, nil
}
