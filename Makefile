.PHONY: test test-coverage test-unit test-integration build run docker-up docker-down clean

# Запуск всех тестов
test:
	go test -v ./...

# Запуск тестов с покрытием
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Запуск только unit тестов
test-unit:
	go test -v -short ./...

# Запуск интеграционных тестов
test-integration:
	go test -v -run Integration ./...

# Запуск тестов с race detector
test-race:
	go test -v -race ./...

# Сборка приложения
build:
	go build -o bin/eventbooker cmd/main.go

# Запуск приложения локально
run:
	go run cmd/main.go

# Запуск Docker Compose
docker-up:
	docker-compose up -d

# Остановка Docker Compose
docker-down:
	docker-compose down

# Остановка и удаление volumes
docker-clean:
	docker-compose down -v

# Просмотр логов
logs:
	docker-compose logs -f app

# Очистка артефактов
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Установка зависимостей
deps:
	go mod download
	go mod tidy

# Линтинг (требует golangci-lint)
lint:
	golangci-lint run

# Форматирование кода
fmt:
	go fmt ./...

# Проверка форматирования
fmt-check:
	@test -z $(shell gofmt -l .)

# Запуск миграций
migrate-up:
	docker-compose exec app goose -dir pkg/migrations postgres "$(MASTER_DSN)" up

# Откат миграций
migrate-down:
	docker-compose exec app goose -dir pkg/migrations postgres "$(MASTER_DSN)" down

# Помощь
help:
	@echo "Available targets:"
	@echo "  test              - Run all tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-integration  - Run integration tests only"
	@echo "  test-race         - Run tests with race detector"
	@echo "  build             - Build the application"
	@echo "  run               - Run the application locally"
	@echo "  docker-up         - Start Docker Compose services"
	@echo "  docker-down       - Stop Docker Compose services"
	@echo "  docker-clean      - Stop services and remove volumes"
	@echo "  logs              - View application logs"
	@echo "  clean             - Clean build artifacts"
	@echo "  deps              - Download and tidy dependencies"
	@echo "  lint              - Run linter"
	@echo "  fmt               - Format code"
	@echo "  fmt-check         - Check code formatting"
	@echo "  help              - Show this help message"
