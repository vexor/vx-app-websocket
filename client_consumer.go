package main

import (
	"github.com/streadway/amqp"
	"log"
  "sync"
  "sync/atomic"
)

type ClientConsumer struct {
	numberOfWorker int
	queueName      string
	exchangeName   string
	consumer       *Consumer
	pool           *ConsumerPool
	delivery       chan amqp.Delivery
  wg             *sync.WaitGroup

  counter        uint64
}

func (c *ClientConsumer) handler(delivery amqp.Delivery) {
  c.delivery <- delivery
}

func (c *ClientConsumer) worker (id int, sessions *ClientSessions) {
  log.Printf("[amqp] [worker.%d] started", id)

  for m := range c.delivery {
    atomic.AddUint64(&c.counter, 1)
    sessions.Each(m.Type, func(c *ClientSession) { c.Send(string(m.Body)) })
  }

  log.Printf("[amqp] [worker.%d] done", id)
  c.wg.Done()
}

func (c *ClientConsumer) ResetCounter() uint64 {
  return atomic.SwapUint64(&c.counter, 0)
}

func (c *ClientConsumer) Wait() {
  c.wg.Wait()
}

func (c *ClientConsumer) Close() {
	c.pool.Close()
	close(c.delivery)
}

func (c *ClientConsumer) Shutdown() {
  log.Printf("[amqp] process shutdown")
	c.consumer.GracefulShutdown()
	c.consumer.WaitSubscribersDone()
	c.Close()
  log.Printf("[amqp] shutdown complete")
}

func NewClientConsumer(sessions *ClientSessions, numberOfWorker int, exchangeName string, queueName string) (*ClientConsumer, error) {

	clientConsumer := &ClientConsumer{
		numberOfWorker: numberOfWorker,
		queueName:      queueName,
		exchangeName:   exchangeName,
		delivery:       make(chan amqp.Delivery, numberOfWorker),
    wg:             new(sync.WaitGroup),
	}

	clientConsumer.pool = NewConsumerPool()

	if err := clientConsumer.pool.Open(); err != nil {
		return nil, err
	}

	clientConsumer.consumer = clientConsumer.pool.NewConsumer(
		exchangeName, queueName, "fanout",
	)

	clientConsumer.consumer.EDurable    = false
	clientConsumer.consumer.EAutoDelete = false
	clientConsumer.consumer.QDurable    = false
	clientConsumer.consumer.QAutoDelete = true
	clientConsumer.consumer.QExclusive  = true
	clientConsumer.consumer.Ack         = true

  for i := 0 ; i < numberOfWorker ; i++ {
    clientConsumer.wg.Add(1)
    go clientConsumer.worker(i, sessions)
  }

	if err := clientConsumer.consumer.Subscribe(clientConsumer.handler); err != nil {
		clientConsumer.Close()
		return nil, err
	}

	return clientConsumer, nil
}

