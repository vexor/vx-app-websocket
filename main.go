package main

import (
	"github.com/igm/sockjs-go/sockjs"
	"log"
	"net/http"
  "time"
  "flag"
)

import _ "net/http/pprof"

func watch(s *ClientSessions, c *ClientConsumer) {
  for {
    select {
    case <- time.After(10 * time.Second):
      log.Printf(
        "[stats] %d clients, %d delivered, %d received",
        s.Len(),
        s.ResetCounter(),
        c.ResetCounter(),
      )
    }
  }
}

var (
  numberOfWorkers int
  queueName       string
  exchangeName    string
  httpBind        string
)

func init() {
  flag.IntVar(&numberOfWorkers, "workers", 10, "number of amqp workers")
  flag.StringVar(&queueName, "queue", "vx.sockd.shipper", "name of amqp queue")
  flag.StringVar(&exchangeName, "exch", "vx.sockd", "name of amqp amqp exchange")
  flag.StringVar(&httpBind, "bind", ":8081", "http address and port to use")
  flag.Parse()
}

func httpHandler(clientSessions *ClientSessions) (func(sock sockjs.Session)) {
  return func(sock sockjs.Session) {
    HandleClientSession(clientSessions, &SockjsClientSessionTransport{sock: sock})
  }
}

func main() {

  clientSessions := NewClientSessions()

  consumer, err := NewClientConsumer(clientSessions, numberOfWorkers, exchangeName, queueName)
  if err != nil {
    log.Fatal(err)
  }

  go watch(clientSessions, consumer)

  http.Handle("/echo/", sockjs.NewHandler("/echo", sockjs.DefaultOptions, httpHandler(clientSessions)))
	log.Printf("[http] listing on %s", httpBind)
  log.Fatal(http.ListenAndServe(httpBind, nil))

  consumer.Shutdown()
}
