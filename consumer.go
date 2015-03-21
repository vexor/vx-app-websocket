package main

import (
	"github.com/streadway/amqp"
	"log"
	"os"
	"sync"
)

type ConsumerHandler func(amqp.Delivery)

type Consumer struct {
	uri      string
	conn     *amqp.Connection
	connLock *sync.Mutex
	handLock *sync.Mutex
	wg       *sync.WaitGroup
	done     chan bool
	closes   []chan error
}

type ConsumerSubscribeOptions struct {
	Queue       string
	Exchange    string
	RoutingKey  string
	Type        string
	EDurable    bool
	EAutoDelete bool
	QDurable    bool
	QAutoDelete bool
	QExclusive  bool
	Ack         bool
}

func NewConsumer() *Consumer {

	uri := os.Getenv("RABBITMQ_URL")
	if uri == "" {
		uri = "amqp://guest:guest@localhost"
	}

	consumer := &Consumer{
		uri:      uri,
		connLock: &sync.Mutex{},
		handLock: &sync.Mutex{},
		wg:       new(sync.WaitGroup),
	}
	return consumer
}

func (self *Consumer) Open() error {
	self.connLock.Lock()
	defer self.connLock.Unlock()

	var err error

	if self.conn != nil {
		return nil
	}

	if self.conn, err = amqp.Dial(self.uri); err != nil {
		self.conn = nil
		return err
	}

	go func() {
		err := <-self.conn.NotifyClose(make(chan *amqp.Error))
		self.handleError(err)
	}()

	self.done = make(chan bool)

	log.Printf("[info ] [amqp] successfuly connected")
	return nil
}

func (self *Consumer) Close() error {
	defer self.connLock.Unlock()
	self.connLock.Lock()

	if self.conn == nil {
		return nil
	}

	if err := self.conn.Close(); err != nil {
		return err
	}
	self.conn = nil
	self.done = nil

	log.Printf("[info ] [amqp] connection closed")

	return nil
}

func (self *Consumer) GracefulShutdown() error {
	log.Printf("[warn ] [amqp] process graceful shutdown")

	close(self.done)

	self.wg.Wait()

	if err := self.Close(); err != nil {
		return err
	}

	log.Printf("[warn ] [amqp] shutdown complete")

	return nil
}

func (self *Consumer) Channel() (*amqp.Channel, error) {
	return self.conn.Channel()
}

func (self *Consumer) NewSubscriber(fn ConsumerHandler, o ConsumerSubscribeOptions) error {

	ch, err := self.Channel()
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(
		o.Exchange,    // name of the exchange
		o.Type,        // type
		o.EDurable,    // durable
		o.EAutoDelete, // delete when complete
		false,         // internal
		false,         // noWait
		nil,           // arguments
	)
	if err != nil {
		ch.Close()
		return err
	}

	q, err := ch.QueueDeclare(
		o.Queue,       // name of the queue
		o.QDurable,    // durable
		o.QAutoDelete, // delete when usused
		o.QExclusive,  // exclusive
		false,         // noWait
		nil,           // arguments
	)
	if err != nil {
		ch.Close()
		return err
	}

	log.Printf(
		"[info ] [amqp] binding %s to %s using [%s]",
		q.Name, o.Exchange, o.RoutingKey,
	)
	err = ch.QueueBind(
		q.Name,       // name of the queue
		o.RoutingKey, // bindingKey
		o.Exchange,   // sourceExchange
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		ch.Close()
		return err
	}

	log.Printf("[info ] [amqp] consume %s", q.Name)
	messages, err := ch.Consume(
		q.Name, // name
		"",     // consumerTag,
		!o.Ack, // auto ack
		false,  // exclusive
		false,  // noLocal
		false,  // noWait
		nil,    // arguments
	)
	if err != nil {
		ch.Close()
		return err
	}

	self.wg.Add(1)
	go self.handle(ch, fn, o.Ack, messages)

	return nil
}

func (self *Consumer) handleError(err *amqp.Error) {
	if err != nil {
		log.Printf("[error] [amqp] %s", err)
		for _, c := range self.closes {
			c <- err
		}
	}
}

func (self *Consumer) NotifyError(c chan error) chan error {
	self.handLock.Lock()
	defer self.handLock.Unlock()

	self.closes = append(self.closes, c)
	return c
}

func (self *Consumer) handle(ch *amqp.Channel, fn ConsumerHandler, ack bool, messages <-chan amqp.Delivery) {
	defer ch.Close()
	defer self.wg.Done()

	for {
		select {
		case message, ok := <-messages:
			// connection level error
			if !ok {
				return
			}

			fn(message)

			if ack {
				message.Ack(false)
			}
		case <-self.done:
			// gracefull shutdown
			return
		}
	}
}

/*
func (c *Consumer) Publish(msg amqp.Publishing) error {
	err := c.pubCh.Publish(c.Exchange, c.RoutingKey, false, false, msg)
	if err != nil {
		return err
	}
	return nil
}
*/
