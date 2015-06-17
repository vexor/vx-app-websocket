package main

import (
	"flag"
	"github.com/igm/sockjs-go/sockjs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

import _ "net/http/pprof"

func watch(s *ClientSessions, c *ClientConsumer) {
	for {
		select {
		case <-time.After(60 * time.Second):
			log.Printf(
				"[info ] [stat] %d clients, %d delivered, %d received",
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
	flag.StringVar(&queueName, "queue", "websocket.%H", "name of amqp queue")
	flag.StringVar(&exchangeName, "exch", "websocket", "name of amqp amqp exchange")
	flag.StringVar(&httpBind, "bind", ":3003", "http address and port to use")
	flag.Parse()
}

func httpHandler(clientSessions *ClientSessions) func(sock sockjs.Session) {
	return func(sock sockjs.Session) {
		HandleClientSession(clientSessions, &SockjsClientSessionTransport{sock: sock})
	}
}

func expandHostName(queueName string) string {

	hostName := "localhost"

	osHostName, err := os.Hostname()
	if err != nil {
		hostName = osHostName
	} else {
		osHostName = os.Getenv("HOSTNAME")
		if osHostName != "" {
			hostName = osHostName
		}
	}
	return strings.Replace(queueName, "%H", hostName, 1)
}

func installSignalHandler(consumer *Consumer) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		log.Printf("[warn ] Got signal: %s", s)
		consumer.GracefulShutdown()
		os.Exit(0)
	}()
}

func catchAmqpError(consumer *Consumer) {
	<-consumer.NotifyError(make(chan error))

	consumer.Close()
	os.Exit(1)
}

func main() {
	log.SetFlags(0)

	sessions := NewClientSessions()
	consumer := NewConsumer()

	client, err := NewClientConsumer(consumer, sessions, exchangeName, expandHostName(queueName))
	if err != nil {
		log.Fatal("[fatal] [amqp] ", err)
	}

	go watch(sessions, client)
	go catchAmqpError(consumer)

	installSignalHandler(consumer)

	http.Handle("/echo/", sockjs.NewHandler("/echo", sockjs.DefaultOptions, httpHandler(sessions)))
	log.Printf("[info ] [http] listing on %s", httpBind)

	err = http.ListenAndServe(httpBind, nil)
	if err != nil {
		log.Fatal("[fatal] [http] ", err)
	}
}
