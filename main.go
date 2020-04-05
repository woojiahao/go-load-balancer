package main

import (
  "net/http/httputil"
  "net/url"
  "sync"
  "sync/atomic"
)

type Backend struct {
  URL          *url.URL
  Alive        bool
  mux          sync.RWMutex
  ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
  backends []*Backend
  current  uint64
}

func (s *ServerPool) NextIndex() int {
  return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) GetNextPeer() *Backend {
  next := s.NextIndex()
  l := len(s.backends) + next
  for i := next; i < l; i++ {
    idx := i % len(s.backends)
    if s.backends[idx].IsAlive() {
      // But why only store when the value is not equal? This should work even if it is equal
      // and it's better to keep the behavior consistent
      if i != next {
        atomic.StoreUint64(&s.current, uint64(idx))
      }
      return s.backends[idx]
    }
  }
  return nil
}

func (b *Backend) SetAlive(alive bool) {
  b.mux.Lock()
  b.Alive = alive
  b.mux.Unlock()
}

func (b *Backend) IsAlive() (alive bool) {
  b.mux.RLock()
  alive = b.Alive
  b.mux.RUnlock()
  return
}
