package main

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rabbitConn *amqp.Connection
	rabbitCh   *amqp.Channel
)

func InitRabbit(cfg Config) error {
	var err error

	for i := 0; i < 10; i++ {
		rabbitConn, err = amqp.Dial(cfg.RabbitURL)
		if err == nil {
			log.Println("Connected to RabbitMQ")
			break
		}
		log.Printf("RabbitMQ not ready, retrying in 2s... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if rabbitConn == nil {
		return fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
	}

	rabbitCh, err = rabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	return nil
}

func PublishMessage(queue string, body []byte) error {
	_, err := rabbitCh.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("queue declare failed: %w", err)
	}

	err = rabbitCh.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	return err
}

func ConsumeQueue(queue string, handler func([]byte) error) error {
	deliveries, err := rabbitCh.Consume(
		queue,
		"node", // consumer tag
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range deliveries {
			if err := handler(d.Body); err != nil {
				log.Printf("handler error, NACKing: %v", err)
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}
		log.Println("delivery channel closed")
	}()
	return nil
}

func CloseRabbit() {
	if rabbitCh != nil {
		_ = rabbitCh.Close()
	}
	if rabbitConn != nil {
		_ = rabbitConn.Close()
	}
}
