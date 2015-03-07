package main

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewClientConsumer(t *testing.T) {

	sessions := NewClientSessions()

	c, err := NewClientConsumer(sessions, 10, "test.vx.sockd", "test.vx.sockd.shipper")
	assert.Nil(t, err)

	time.Sleep(500 * time.Millisecond)

	msg := amqp.Publishing{
		Timestamp:   time.Now(),
		ContentType: "text/plain",
		Body:        []byte("Ping"),
	}

	for i := 0; i < 100; i++ {
		err := c.consumer.Publish(msg)
		assert.Nil(t, err)
	}

	time.Sleep(1000 * time.Millisecond)

	c.Close()
	c.Wait()

	assert.Equal(t, 100, c.counter)

}
