# EventBooker

Система бронирования мест на мероприятия с автоматической отменой неоплаченных броней через 15 минут.

## Описание

EventBooker - это веб-приложение для управления мероприятиями и бронированием мест. Система автоматически отменяет неоплаченные бронирования через 15 минут с использованием RabbitMQ и механизма Dead Letter Exchange (DLX).

## Основные возможности

- **Создание мероприятий** с указанием названия, описания, даты, количества мест и цены
- **Бронирование мест** с автоматическим уменьшением доступных билетов
- **Оплата бронирований** с подтверждением статуса
- **Автоматическая отмена** неоплаченных броней через 15 минут
- **Возврат мест** при отмене бронирования
- **Веб-интерфейс** для пользователей и администраторов
- **Таймер обратного отсчета** для оплаты бронирования

## Архитектура

### Технологический стек

- **Backend**: Go 1.25
- **Database**: PostgreSQL 15
- **Message Broker**: RabbitMQ 3.12
- **Web Server**: Gorilla Mux
- **Containerization**: Docker & Docker Compose

### Структура проекта

```
EventBooker/
├── cmd/                          # Точка входа приложения
│   └── main.go
├── config/                       # Конфигурация
│   └── config.go
├── internal/
│   ├── adapter/                  # Адаптеры внешних сервисов
│   │   ├── broker/              # RabbitMQ producer
│   │   │   └── rabbitmq.go
│   │   ├── consumer/            # RabbitMQ consumers
│   │   │   ├── cancellation_consumer.go
│   │   │   └── confirmation_consumer.go
│   │   └── repository/          # Работа с БД
│   │       └── postgres/
│   │           └── postgres.go
│   ├── app/                     # Инициализация приложения
│   │   └── app.go
│   ├── domain/                  # Доменные модели
│   │   └── event.go
│   ├── input/                   # HTTP handlers
│   │   └── http/
│   │       ├── handlers.go
│   │       └── server.go
│   ├── port/                    # Интерфейсы
│   │   ├── broker.go
│   │   ├── repository.go
│   │   └── usecases.go
│   └── usecases/                # Бизнес-логика
│       └── events.go
├── pkg/
│   └── migrations/              # Миграции БД
│       ├── 001_create_events_table.sql
│       └── migrations.go
├── web/                         # Веб-интерфейс
│   ├── admin.html              # Панель администратора
│   └── user.html               # Пользовательский интерфейс
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── readme.md
```

### Схема работы с RabbitMQ

#### 1. При бронировании места
- Система создает запись в БД со статусом `pending`
- Публикует сообщение в очередь `waiting_cancellations` с TTL = 15 минут
- Сообщение "засыпает" в очереди на 15 минут

#### 2. При оплате (если оплатили вовремя)
- Система меняет статус брони в БД на `confirmed`
- Публикует сообщение в очередь `confirmations` (без задержки)
- Consumer подтверждений обрабатывает сообщение (логирование, кэш и т.д.)

#### 3. Через 15 минут (если не оплатили)
- Сообщение из очереди `waiting_cancellations` через DLX попадает в `delayed_cancellations`
- Consumer отмен получает его и проверяет статус брони в БД
- Если статус `pending` → отменяет бронь и возвращает место
- Если статус `confirmed` → ничего не делает, удаляет сообщение

## Установка и запуск

### Требования

- Docker 20.10+
- Docker Compose 2.0+

### Быстрый старт

1. Клонируйте репозиторий:
```bash
git clone https://github.com/dontpanicw/EventBooker.git
cd EventBooker
```

2. Создайте файл `.env` на основе `.env.example`:
```bash
cp .env.example .env
```

3. Запустите приложение:
```bash
docker-compose up -d
```

4. Приложение будет доступно по адресам:
   - **Пользовательский интерфейс**: http://localhost:8080
   - **Панель администратора**: http://localhost:8080/admin
   - **RabbitMQ Management**: http://localhost:15672 (guest/guest)

### Переменные окружения

```env
# HTTP Server
HTTP_PORT=8080

# PostgreSQL Database
MASTER_DSN=postgres://eventbooker:eventbooker_pass@postgres:5432/eventbooker?sslmode=disable

# RabbitMQ
RABBITMQ_URL=amqp://eventbooker:eventbooker_pass@rabbitmq:5672/
```

## API Endpoints

### События

#### Создать мероприятие
```http
POST /api/events
Content-Type: application/json

{
  "name": "Концерт",
  "description": "Описание мероприятия",
  "date": "2026-03-01T19:00:00Z",
  "available_tickets": 100,
  "is_free": false,
  "price": 1500.00
}
```

#### Получить все мероприятия
```http
GET /api/events
```

#### Получить мероприятие по ID
```http
GET /api/events/{id}
```

### Бронирования

#### Забронировать место
```http
POST /api/events/{id}/book
Content-Type: application/json

{
  "user_id": "user_123"
}
```

#### Оплатить бронирование
```http
POST /api/bookings/{id}/confirm
```

#### Получить информацию о бронировании
```http
GET /api/bookings/{id}
```

## Веб-интерфейс

### Пользовательская страница (/)

- Просмотр доступных мероприятий
- Бронирование мест
- Таймер обратного отсчета (15 минут)
- Оплата бронирования
- Отображение статуса брони (pending/confirmed/cancelled)

### Административная панель (/admin)

- Создание новых мероприятий
- Просмотр всех мероприятий
- Мониторинг свободных мест
- Автообновление списка каждые 5 секунд

## Тестирование

### Запуск тестов

```bash
# Все тесты
make test

# Тесты с покрытием
make test-coverage

# Unit тесты
make test-unit

# Тесты с race detector
make test-race
```

### Структура тестов

```
internal/
├── usecases/
│   └── events_test.go          # Тесты бизнес-логики
├── input/http/
│   └── handlers_test.go        # Тесты HTTP handlers
├── adapter/consumer/
│   ├── cancellation_consumer_test.go
│   └── confirmation_consumer_test.go
└── domain/
    └── event_test.go           # Тесты доменных моделей
```

### Покрытие кода

Тесты покрывают:
- ✅ Usecases (бизнес-логика)
- ✅ HTTP Handlers
- ✅ Domain models
- ✅ RabbitMQ consumers
- ✅ Обработка ошибок
- ✅ Валидация данных

### CI/CD

GitHub Actions автоматически запускает тесты при каждом push и pull request:
- Запуск всех тестов
- Проверка race conditions
- Генерация отчета о покрытии
- Линтинг кода

## Мониторинг

### RabbitMQ Management UI

Доступен по адресу: http://localhost:15672

Логин: `eventbooker`  
Пароль: `eventbooker_pass`

Здесь можно отслеживать:
- Количество сообщений в очередях
- Скорость обработки сообщений
- Статус consumers
- Dead letter messages

## Контакты

- GitHub: [@dontpanicw](https://github.com/dontpanicw)
- Email: your-email@example.com

## Roadmap

- [ ] Добавить аутентификацию пользователей
- [ ] Реализовать email уведомления
- [ ] Интеграция с платежными системами, отдельный микросервис для оплаты
- [ ] Добавить отчеты и аналитику
- [ ] Добавить кеш
