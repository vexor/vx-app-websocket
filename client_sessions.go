package main

import (
  "sync"
  "sync/atomic"
)

type ClientSessions struct {
  sync.Mutex
  list    []*ClientSession
  counter uint64
}

func (c *ClientSessions) Add(s *ClientSession) {
  c.Lock()
  defer c.Unlock()
  c.list = append(c.list, s)
}

func (c *ClientSessions) Remove(s *ClientSession) {
  c.Lock()
  defer c.Unlock()

  idx := c.Index(s)
  if idx != -1 {
    c.list = append(c.list[:idx], c.list[idx+1:]...)
  }
}

func (c *ClientSessions) Len() int {
  return len(c.list)
}

func (c *ClientSessions) IncCounter () uint64 {
  return atomic.AddUint64(&c.counter, 1)
}

func (c *ClientSessions) ResetCounter () uint64 {
  return atomic.SwapUint64(&c.counter, 0)
}

func (c *ClientSessions) Index(s *ClientSession) int {
  for p, v := range c.list {
    if (v == s) {
      return p
    }
  }
  return -1
}

func (c *ClientSessions) Each(channelId string, fn func(*ClientSession)) {
  for _, s := range c.list {
    if s.channelId == channelId {
      fn(s)
    }
  }
}

func NewClientSessions() *ClientSessions {
  s := &ClientSessions{}
  return s
}
