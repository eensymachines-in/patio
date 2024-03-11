package main

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestAmqpConnection(t *testing.T) {
	t.Log("Now starting test")
	conn, err := amqp.Dial("amqp://guest:guest@192.168.1.102:30073/")
	assert.Nil(t, err, "Unexpected error when connecting to amqp server")

	ch, err := conn.Channel()
	assert.Nil(t, err, "Unexpected error when instantiating channel")
	assert.NotNil(t, ch, "Unexpected nil channel")

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := "Hello World!"
	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	assert.Nil(t, err, "Unexpected nil error when publishing with context")
	defer conn.Close()
}
