package main

import (
	"github.com/igm/sockjs-go/sockjs"
)

type ClientSessionTransport interface {
  Send(string) error
  Recv() (string, error)
}

type SockjsClientSessionTransport struct {
  ClientSessionTransport
  sock sockjs.Session
}

func (s *SockjsClientSessionTransport) Send(str string) error {
  return s.sock.Send(str)
}

func (s *SockjsClientSessionTransport) Recv() (string, error) {
  return s.sock.Recv()
}
