package main

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewClientConsumer(t *testing.T) {

	sessions := NewClientSessions()
	consumer := NewConsumer()

	defer consumer.Close()

	c, err := NewClientConsumer(consumer, sessions, "test.websocket", "test.websocket.consumer")
	assert.Nil(t, err)

	time.Sleep(200 * time.Millisecond)

	msg := amqp.Publishing{
		Timestamp:   time.Now(),
		ContentType: "text/plain",
		Body:        []byte("Ping"),
	}

	ch, err := consumer.Channel()
	assert.Nil(t, err)

	for i := 0; i < 100; i++ {
		err := ch.Publish("test.websocket", "", false, false, msg)
		assert.Nil(t, err)
	}

	time.Sleep(300 * time.Millisecond)

	done := make(chan bool)

	go func() {
		consumer.GracefulShutdown()
		done <- true
	}()

	select {
	case <-done:
		assert.Equal(t, 100, int(c.counter))
	case <-time.After(300 * time.Millisecond):
		t.Errorf("Timeout")
	}

}
