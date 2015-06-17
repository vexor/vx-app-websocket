package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewConsumer(t *testing.T) {
	c := NewConsumer()
	assert.Equal(t, c.uri, "amqp://guest:guest@localhost")
}

func TestConsumerOpen(t *testing.T) {
	c := NewConsumer()

	assert.Nil(t, c.conn)

	c.Open()
	defer c.Close()

	assert.NotNil(t, c.conn)
}

func TestConsumerClose(t *testing.T) {
	c := NewConsumer()

	err := c.Open()
	assert.Nil(t, err)
	defer c.Close()

	assert.NotNil(t, c.conn)

	c.Close()
	assert.Nil(t, c.conn)
}

func TestConsumerNewSubscriber(t *testing.T) {
	c := NewConsumer()

	assert.Nil(t, c.Open())
	defer c.Close()

	handler := func(d amqp.Delivery) {
		log.Printf("GOT: %+v", string(d.Body))
	}

	err := c.NewSubscriber(
		handler,
		ConsumerSubscribeOptions{
			Queue:       "foo",
			Exchange:    "bar",
			Type:        "topic",
			QAutoDelete: true,
			EAutoDelete: true,
		},
	)
	assert.Nil(t, err)

	pub, err := c.Channel()
	assert.Nil(t, err)

	msg := amqp.Publishing{
		Timestamp:   time.Now(),
		ContentType: "text/plain",
		Body:        []byte("Ping"),
	}
	err = pub.Publish("bar", "", false, false, msg)
	assert.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	done := make(chan bool)
	defer close(done)

	go func() {
		c.GracefulShutdown()
		done <- true
	}()

	select {
	case <-done:
		log.Printf("Done")
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout")
	}
}

func TestPubSub(t *testing.T) {
	counter := uint32(0)
	c := NewConsumer()

	assert.Nil(t, c.Open())
	defer c.Close()

	pubCh, err := c.Channel()
	assert.Nil(t, err)

	pubHandler := func(n int) ConsumerHandler {
		return func(d amqp.Delivery) {
			msg := amqp.Publishing{
				Timestamp:   time.Now(),
				ContentType: "text/plain",
				Body:        d.Body,
			}
			err = pubCh.Publish("sub", "", false, false, msg)
			assert.Nil(t, err)
		}
	}

	subHandler := func(n int) ConsumerHandler {
		return func(d amqp.Delivery) {
			atomic.AddUint32(&counter, 1)
		}
	}

	for i := 0; i < 10; i++ {
		err = c.NewSubscriber(
			pubHandler(i),
			ConsumerSubscribeOptions{
				Queue:       "pub",
				Exchange:    "pub",
				Type:        "topic",
				QAutoDelete: true,
				EAutoDelete: true,
			},
		)
		assert.Nil(t, err)
	}

	for i := 0; i < 10; i++ {
		err = c.NewSubscriber(
			subHandler(i),
			ConsumerSubscribeOptions{
				Queue:       "sub",
				Exchange:    "sub",
				Type:        "topic",
				QAutoDelete: true,
				EAutoDelete: true,
			},
		)
		assert.Nil(t, err)
	}

	time.Sleep(100 * time.Millisecond)

	pub, err := c.Channel()
	assert.Nil(t, err)

	for i := 0; i < 100; i++ {
		msg := amqp.Publishing{
			Timestamp:   time.Now(),
			ContentType: "text/plain",
			Body:        []byte(fmt.Sprintf("Ping %n", i)),
		}
		err = pub.Publish("pub", "", false, false, msg)
		assert.Nil(t, err)
	}

	time.Sleep(200 * time.Millisecond)

	done := make(chan bool)

	go func() {
		c.GracefulShutdown()
		done <- true
	}()

	select {
	case <-done:
		log.Printf("Done")
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout")
	}

	assert.Equal(t, 100, int(counter))
}
