package main

import (
	"errors"
	"sync"
)

type ClientSession struct {
	sock         ClientSessionTransport
	wg           *sync.WaitGroup
	channelId    string
	sendCh       chan string
	closed       bool
  sessions     *ClientSessions
}

func (session *ClientSession) handleMessage(msg string) {
  if len(msg) > 12 {
    c := msg[0:10]
    if c == "subscribe:" {
      msgLen := len(msg)
      if msgLen > 80 {
        msgLen = 80
      }
      channelId := msg[11:msgLen]
      session.channelId = channelId
    }
  }
}

func (session *ClientSession) Send(msg string) error {
	if session.closed {
		return errors.New("ClientSession closed")
	}
	session.sendCh <- msg
	return nil
}

func (session *ClientSession) reader() {
	for {
		msg, err := session.sock.Recv()
		if err != nil {
      session.Close()
			break
		}
		session.handleMessage(msg)
	}
	session.wg.Done()
}

func (session *ClientSession) writer() {

  for msg := range session.sendCh {
    if err := session.sock.Send(msg); err != nil {
      break
    }
    session.sessions.IncCounter()
  }
	session.wg.Done()
}

func (session * ClientSession) Wait () {
  session.wg.Wait()
}

func (session *ClientSession) Close() {
  if session.closed == false {
    session.sessions.Remove(session)
    session.closed = true
    close(session.sendCh)
  }
}

func (session *ClientSession) Shutdown() {
  session.Wait()
  session.Close()
}

func (s *ClientSession) Consume() {
  s.sessions.Add(s)

	s.wg.Add(1)
	go s.writer()

	s.wg.Add(1)
	go s.reader()
}

func NewClientSession(sessions *ClientSessions, sock ClientSessionTransport) *ClientSession {
	s := &ClientSession{
		sock:     sock,
		sendCh:   make(chan string, 1),
		wg:       new(sync.WaitGroup),
		closed:   false,
    sessions: sessions,
	}
  return s
}

func HandleClientSession(sessions *ClientSessions, sock ClientSessionTransport) {
  s := NewClientSession(sessions, sock)
  s.Consume()
  s.Shutdown()
}
