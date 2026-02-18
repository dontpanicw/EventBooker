package consumer

//
//import (
//	"context"
//	"encoding/json"
//	"log"
//
//	"github.com/dontpanicw/EventBooker/internal/adapter/broker"
//	amqp "github.com/rabbitmq/amqp091-go"
//)
//
//type ConfirmationConsumer struct {
//	channel *amqp.Channel
//}
//
//func NewConfirmationConsumer(channel *amqp.Channel) *ConfirmationConsumer {
//	return &ConfirmationConsumer{
//		channel: channel,
//	}
//}
//
//func (c *ConfirmationConsumer) Start(ctx context.Context) error {
//	msgs, err := c.channel.Consume(
//		broker.ConfirmationsQueue,
//		"confirmation_consumer",
//		false, // auto-ack
//		false,
//		false,
//		false,
//		nil,
//	)
//	if err != nil {
//		return err
//	}
//
//	log.Println("Confirmation consumer started")
//
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				log.Println("Confirmation consumer stopped")
//				return
//			case msg, ok := <-msgs:
//				if !ok {
//					log.Println("Channel closed")
//					return
//				}
//				c.handleMessage(msg)
//			}
//		}
//	}()
//
//	return nil
//}
//
//func (c *ConfirmationConsumer) handleMessage(msg amqp.Delivery) {
//	var bookingMsg broker.BookingMessage
//	if err := json.Unmarshal(msg.Body, &bookingMsg); err != nil {
//		log.Printf("Failed to unmarshal message: %v", err)
//		msg.Nack(false, false)
//		return
//	}
//
//	log.Printf("Processing confirmation for booking %s", bookingMsg.BookingID)
//
//	// Здесь можно добавить дополнительную логику:
//	// - Обновление кэша
//	// - Отправка уведомлений
//	// - Логирование в аналитику
//	// - Обновление статистики
//
//	log.Printf("Booking %s confirmed successfully", bookingMsg.BookingID)
//
//	msg.Ack(false)
//}
