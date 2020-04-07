package main

import (
  "log"
  "net/url"
  "sync/atomic"
)

type ServerPool struct {
  backends []*Backend
  current  uint64
}

func (s *ServerPool) AddBackend(backend *Backend) {
  s.backends = append(s.backends, backend)
}

func (s *ServerPool) NextIndex() int {
  return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
  for _, b := range s.backends {
    if b.URL.String() == backendUrl.String() {
      b.SetAlive(alive)
      break
    }
  }
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

func (s *ServerPool) HealthCheck() {
  for _, b := range s.backends {
    status := "up"
    alive := isBackendAlive(b.URL)
    b.SetAlive(alive)
    if !alive {
      status = "down"
    }
    log.Printf("%s [%s]\n", b.URL, status)
  }
}

