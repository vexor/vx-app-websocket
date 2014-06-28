package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
  "errors"
  "time"
)

type TestClientSessionTransport struct {
  ClientSessionTransport

  sendError bool
  sended    []string

  ch        chan string
  recvError chan error
}

func (t *TestClientSessionTransport) Send(str string) error {
  if t.sendError {
    return errors.New("sendError")
  } else {
    t.sended = append(t.sended, str)
    return nil
  }
}

func (t *TestClientSessionTransport) Recv() (string, error) {
  select {
    case m := <- t.ch:
      return m, nil
    case e := <- t.recvError:
      return "", e
  }
}

func NewTestClientSessionTransport() *TestClientSessionTransport {
  transport := &TestClientSessionTransport{
    ch:        make(chan string, 1),
    recvError: make(chan error, 1),
  }
  return transport
}

func TestClientSessionAddChannel(t *testing.T) {
  sessions  := NewClientSessions()
  transport := NewTestClientSessionTransport()

  s := NewClientSession(sessions, transport)
  assert.NotNil(t, s)

  assert.Equal(t, sessions.Index(s), -1)
  s.Consume()

  // should be in sessions
  time.Sleep(100 * time.Millisecond)
  assert.Equal(t, sessions.Index(s), 0)

  // should be subscribed to channel
  assert.Equal(t, len(transport.sended), 0)
  transport.ch <- "subscribe: foo"
  time.Sleep(100 * time.Millisecond)
  sessions.Each("foo", func(c *ClientSession) { c.Send("str") })
  time.Sleep(100 * time.Millisecond)
  assert.Equal(t, transport.sended, []string{ "str" })

  s.Close()
}

func TestClientSessionReadError(t *testing.T) {
  sessions  := NewClientSessions()
  transport := NewTestClientSessionTransport()

  s := NewClientSession(sessions, transport)
  assert.NotNil(t, s)

  assert.Equal(t, sessions.Index(s), -1)
  s.Consume()
  time.Sleep(100 * time.Millisecond)

  transport.recvError <- errors.New("close")

  s.Wait()
  s.Close()
}

func TestClientSessionWriteError(t *testing.T) {
  sessions  := NewClientSessions()
  transport := NewTestClientSessionTransport()

  s := NewClientSession(sessions, transport)
  assert.NotNil(t, s)

  assert.Equal(t, sessions.Index(s), -1)
  s.Consume()
  time.Sleep(100 * time.Millisecond)

  s.Send("str")
  time.Sleep(100 * time.Millisecond)
  assert.Equal(t, transport.sended, []string{ "str" })

  transport.sendError = true
  s.Send("str2")
  time.Sleep(100 * time.Millisecond)
  assert.Equal(t, transport.sended, []string{ "str" })

  transport.recvError <- errors.New("close")

  s.Wait()
  s.Close()
}
