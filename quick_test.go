package main

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestAmqpConnection(t *testing.T) {
	t.Log("Now starting test")
	conn, err := amqp.Dial("amqp://guest:guest@192.168.1.102:30073/")

	assert.Nil(t, err, "Unexpected error when connecting to amqp server")
	defer conn.Close()

	ch, err := conn.Channel()
	assert.Nil(t, err, "Unexpected error when instantiating channel")
	assert.NotNil(t, ch, "Unexpected nil channel")

	q, err := ch.QueueDeclare(
		"aquapone.test", // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	assert.Nil(t, err, "Failed to declare queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	assert.Nil(t, err, "Failed to start consume channel")
	var forever chan struct{}
	go func() {
		t.Log("Now listening on messages")
		for d := range msgs {
			t.Logf("Received a message: %s", d.Body)
		}
	}()
	t.Log(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
