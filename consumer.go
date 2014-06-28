package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
)

type ConsumerHandler func(amqp.Delivery)

type ConsumerPool struct {
	uri  string
	conn *amqp.Connection
}

type Consumer struct {
	Queue      string
	Exchange   string
	RoutingKey string
	Type       string

	EDurable    bool
	EAutoDelete bool

	QDurable    bool
	QAutoDelete bool
	QExclusive  bool

  Ack         bool

	pubCh *amqp.Channel
	subCh *amqp.Channel
	done  chan error

	pool *ConsumerPool
}

func rabbitmqURL() string {
	uri := os.Getenv("RABBITMQ_URL")
	if uri == "" {
		uri = "amqp://guest:guest@localhost"
	}
	return uri
}

func NewConsumerPool() *ConsumerPool {
	uri := rabbitmqURL()
	pool := &ConsumerPool{
		uri: uri,
	}

	return pool
}

func (p *ConsumerPool) Open() error {
	var err error

	if p.conn == nil {
		log.Printf("[amqp] connecting to %s", p.uri)
		p.conn, err = amqp.Dial(p.uri)
		return err
	} else {
		return nil
	}
}

func (p *ConsumerPool) Close() error {
	if p.conn != nil {
		log.Printf("[amqp] closing connection")
		if err := p.conn.Close(); err != nil {
			return err
		}
		p.conn = nil
	}
	return nil
}

func (p *ConsumerPool) NewConsumer(exch string, queue string, exchType string) *Consumer {

	c := &Consumer{
		Queue:    queue,
		Exchange: exch,
		Type:     exchType,
		done:     make(chan error),
		pool:     p,
	}

	return c
}

func (c *Consumer) allocateChannels() error {
	var err error

	log.Println("[amqp] allocating pub/sub channels")

	c.pubCh, err = c.pool.conn.Channel()
	if err != nil {
		return err
	}

	c.subCh, err = c.pool.conn.Channel()
	if err != nil {
		c.pubCh.Close()
		c.pubCh = nil
		return err
	}

	go func() {
		c.done <- fmt.Errorf("[amqp] connection closed: %s\n", <-c.pool.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	return err
}

func (c *Consumer) declareExchange() error {
	err := c.subCh.ExchangeDeclare(
		c.Exchange,    // name of the exchange
		c.Type,        // type
		c.EDurable,    // durable
		c.EAutoDelete, // delete when complete
		false,         // internal
		false,         // noWait
		nil,           // arguments
	)
	return err
}

func (c *Consumer) Publish(msg amqp.Publishing) error {
	err := c.pubCh.Publish(c.Exchange, c.RoutingKey, false, false, msg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Consumer) declareAndBindQueue() error {
	log.Println("[amqp] declaring queue:", c.Queue)

	q, err := c.subCh.QueueDeclare(
		c.Queue,       // name of the queue
		c.QDurable,    // durable
		c.QAutoDelete, // delete when usused
		c.QExclusive,  // exclusive
		false,         // noWait
		nil,           // arguments
	)

	if err != nil {
		return err
	}

	c.Queue = q.Name

	log.Printf("[amqp] binding %s to %s using [%s]", q.Name, c.Exchange, c.RoutingKey)
	err = c.subCh.QueueBind(
		q.Name,       // name of the queue
		c.RoutingKey, // bindingKey
		c.Exchange,   // sourceExchange
		false,        // noWait
		nil,          // arguments
	)

	if err != nil {
		return err
	}

	return nil
}

func (c *Consumer) Subscribe(fn ConsumerHandler) error {

	if err := c.allocateChannels(); err != nil {
		return err
	}

	if err := c.declareExchange(); err != nil {
		return err
	}

	if err := c.declareAndBindQueue(); err != nil {
		return err
	}

	log.Printf("[amqp] subscribing to %s", c.Queue)
	messages, err := c.subCh.Consume(
		c.Queue, // name
		"",      // consumerTag,
		!c.Ack,  // noAck
		false,   // exclusive
		false,   // noLocal
		false,   // noWait
		nil,     // arguments
	)

	if err != nil {
		return err
	}

	go func() {
		for m := range messages {
			fn(m)
      if c.Ack {
        m.Ack(false)
      }
		}
		c.done <- nil
	}()

	return nil
}

func (c *Consumer) GracefulShutdown() {
  log.Printf("[amqp] processing graceful shutdown")
	c.done <- nil
}

func (c *Consumer) WaitSubscribersDone() error {
  err := <-c.done
  log.Printf("[amqp] shutdown")

  return err
}
