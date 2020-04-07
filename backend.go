package main

import (
  "net/http/httputil"
  "net/url"
  "sync"
)

type Backend struct {
  URL          *url.URL
  Alive        bool
  mux          sync.RWMutex
  ReverseProxy *httputil.ReverseProxy
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
