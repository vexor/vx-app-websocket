package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClientSessionsEach(t *testing.T) {
  s := NewClientSessions()
  assert.NotNil(t, s)

  c1 := &ClientSession{}
  c1.channelId = "c1"
  s.Add(c1)

  c2 := &ClientSession{}
  c2.channelId = "c2"
  s.Add(c2)

  var q []*ClientSession

  fn := func(c *ClientSession) {
    q = append(q, c)
  }

  s.Each("c1", fn)
  assert.Equal(t, len(q), 1)
  assert.Equal(t, []*ClientSession{c1}, q)

  q = q[0:0]
  s.Each("bar", fn)
  assert.Equal(t, len(q), 0)
}

func TestClientSessionsAddRemove(t *testing.T) {
  s := NewClientSessions()
  assert.NotNil(t, s)
  assert.Equal(t, s.Index(nil), -1)

  c1 := &ClientSession{}
  s.Add(c1)

  assert.Equal(t, s.Len(), 1)
  assert.Equal(t, s.Index(c1), 0)

  c2 := &ClientSession{}
  s.Add(c2)

  assert.Equal(t, s.Len(), 2)
  assert.Equal(t, s.Index(c2), 1)

  c3 := &ClientSession{}
  s.Add(c3)

  assert.Equal(t, s.Len(), 3)
  assert.Equal(t, s.Index(c3), 2)

  s.Remove(c2)
  assert.Equal(t, s.Len(), 2)
  assert.Equal(t, s.Index(c2), -1)
  assert.Equal(t, s.Index(c1), 0)
  assert.Equal(t, s.Index(c3), 1)

  s.Remove(c1)
  assert.Equal(t, s.Len(), 1)
  assert.Equal(t, s.Index(c2), -1)
  assert.Equal(t, s.Index(c1), -1)
  assert.Equal(t, s.Index(c3), 0)

  s.Remove(c3)
  assert.Equal(t, s.Len(), 0)
  assert.Equal(t, s.Index(c2), -1)
  assert.Equal(t, s.Index(c1), -1)
  assert.Equal(t, s.Index(c3), -1)
}
