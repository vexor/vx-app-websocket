package main

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	pool := NewConsumerPool()
	assert.Equal(t, pool.uri, "amqp://guest:guest@localhost")
}

func TestPoolOpen(t *testing.T) {
	pool := NewConsumerPool()

	assert.Nil(t, pool.conn)

	pool.Open()
	defer pool.Close()

	assert.NotNil(t, pool.conn)
}

func TestPoolClose(t *testing.T) {
	pool := NewConsumerPool()

	err := pool.Open()
	assert.Nil(t, err)
	defer pool.Close()

	assert.NotNil(t, pool.conn)

	pool.Close()
	assert.Nil(t, pool.conn)
}

func TestNewConsumer(t *testing.T) {
	p := NewConsumerPool()

	assert.Nil(t, p.Open())
	defer p.Close()

	c := p.NewConsumer("foo", "bar", "topic")
	assert.Equal(t, c.Exchange, "foo")
	assert.Equal(t, c.Queue, "bar")
	assert.Equal(t, c.Type, "topic")
}

func TestSubscribeConsumer(t *testing.T) {
	p := NewConsumerPool()

	assert.Nil(t, p.Open())
	defer p.Close()

	c := p.NewConsumer("foo", "bar", "topic")

	c.QAutoDelete = true
	c.EAutoDelete = true

	handler := func(d amqp.Delivery) {
		log.Printf("%+v", d)
		c.done <- nil
	}

	c.Subscribe(handler)

	msg := amqp.Publishing{
		Timestamp:   time.Now(),
		ContentType: "text/plain",
		Body:        []byte("Ping"),
	}

	err := c.Publish(msg)
	assert.Nil(t, err)

	select {
	case err := <-c.done:
		assert.Nil(t, err)
	case <-time.After(time.Second):
		t.Errorf("Timeout")
	}
}

func TestPubSub(t *testing.T) {
	var err error
	pubIdx := 0
	subIdx := 0

	p := NewConsumerPool()

	assert.Nil(t, p.Open())
	defer p.Close()

	c1 := p.NewConsumer("foo1", "bar1", "topic")

	c1.QAutoDelete = true
	c1.EAutoDelete = true

	c2 := p.NewConsumer("foo2", "bar2", "topic")

	c2.QAutoDelete = true
	c2.EAutoDelete = true

	pubHandler := func(d amqp.Delivery) {

		time.Sleep(100 * time.Millisecond)

		log.Printf("<-- [%n]", pubIdx)
		msg := amqp.Publishing{
			Timestamp:   time.Now(),
			ContentType: "text/plain",
			Body:        d.Body,
		}
		c2.Publish(msg)

		pubIdx++
		if pubIdx == 9 {
			c1.done <- nil
		}
	}

	subHandler := func(d amqp.Delivery) {
		log.Printf("--> [%n]", subIdx)
		subIdx++
		if subIdx == 9 {
			c2.done <- nil
		}
	}

	err = c1.Subscribe(pubHandler)
	assert.Nil(t, err)

	err = c2.Subscribe(subHandler)
	assert.Nil(t, err)

	msg := amqp.Publishing{
		Timestamp:   time.Now(),
		ContentType: "text/plain",
		Body:        []byte("Ping!"),
	}

	for i := 0; i < 10; i++ {
		err = c1.Publish(msg)
		assert.Nil(t, err)
	}

	select {
	case err = <-c1.done:
		assert.Nil(t, err)
	case <-time.After(2 * time.Second):
		t.Errorf("Sub Timeout")
	}

	select {
	case err = <-c2.done:
		assert.Nil(t, err)
	case <-time.After(2 * time.Second):
		t.Errorf("Pub Timeout")
	}
}
