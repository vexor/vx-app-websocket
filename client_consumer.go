package main

import (
	"github.com/streadway/amqp"
	"sync/atomic"
)

type ClientConsumer struct {
	counter uint64
}

func (self *ClientConsumer) handler(sessions *ClientSessions) ConsumerHandler {
	return func(m amqp.Delivery) {
		atomic.AddUint64(&self.counter, 1)
		sessions.Each(m.Type, func(c *ClientSession) {
			c.Send(string(m.Body))
		})
	}
}

func (self *ClientConsumer) ResetCounter() uint64 {
	return atomic.SwapUint64(&self.counter, 0)
}

func NewClientConsumer(consumer *Consumer, sessions *ClientSessions, exchangeName string, queueName string) (*ClientConsumer, error) {

	c := &ClientConsumer{}

	if err := consumer.Open(); err != nil {
		return nil, err
	}

	err := consumer.NewSubscriber(
		c.handler(sessions),
		ConsumerSubscribeOptions{
			Queue:       queueName,
			Exchange:    exchangeName,
			Type:        "fanout",
			EDurable:    false,
			EAutoDelete: false,
			QDurable:    false,
			QAutoDelete: true,
			QExclusive:  true,
			Ack:         true,
		},
	)
	if err != nil {
		consumer.Close()
		return nil, err
	}

	return c, nil
}
