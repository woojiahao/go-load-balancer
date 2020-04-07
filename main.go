package main

import (
  "context"
  "flag"
  "fmt"
  "log"
  "net"
  "net/http"
  "net/http/httputil"
  "net/url"
  "strings"
  "time"
)

const (
  Attempts int = iota
  Retry
)

func GetAttemptsFromContext(r *http.Request) int {
  if attempts, ok := r.Context().Value(Attempts).(int); ok {
    return attempts
  }
  return 0
}

func GetRetryFromContext(r *http.Request) int {
  if retry, ok := r.Context().Value(Retry).(int); ok {
    return retry
  }
  return 0
}

func lb(w http.ResponseWriter, r *http.Request) {
  attempts := GetAttemptsFromContext(r)
  if attempts > 3 {
    log.Printf("%s (%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path, )
    http.Error(w, "Service not available", http.StatusServiceUnavailable)
    return
  }

  peer := serverpool.GetNextPeer()
  if peer != nil {
    peer.ReverseProxy.ServeHTTP(w, r)
    return
  }

  http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func isBackendAlive(u *url.URL) bool {
  timeout := 2 * time.Second
  conn, err := net.DialTimeout("tcp", u.Host, timeout)
  if err != nil {
    log.Println("Site unreachable error: ", err)
    return false
  }
  _ = conn.Close()
  return true
}

func healthCheck() {
  t := time.NewTicker(time.Second * 20)
  for {
    select {
    case <-t.C:
      log.Println("Starting health check...")
      serverPool.HealthCheck()
      log.Println("Health check completed")
    }
  }
}

var serverPool ServerPool

func main() {
  var serverList string
  var port int
  flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
  flag.IntVar(&port, "port", 3030, "Port to serve")
  flag.Parse()

  if len(serverList) == 0 {
    log.Fatal("Please provide one or more backends to load balance")
  }

  tokens := strings.Split(serverList, ",")
  for _, tok := range tokens {
    serverUrl, err := url.Parse(tok)
    if err != nil {
      log.Fatal(err)
    }

    proxy := httputil.NewSingleHostReverseProxy(serverUrl)
    proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
      log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
      retries := GetRetryFromContext(request)
      if retries < 3 {
        select {
        case <-time.After(10 * time.Millisecond):
          ctx := context.WithValue(request.Context(), Retry, retries+1)
          proxy.ServeHTTP(writer, request.WithContext(ctx))
        }
        return
      }

      // after 3 retries, mark this backend as down
      serverPool.MarkBackendStatus(serverUrl, false)

      // if the same request routing for few attempts with different backends, increase the count
      attempts := GetAttemptsFromContext(request)
      log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
      ctx := context.WithValue(request.Context(), Attempts, attempts+1)
      lb(writer, request.WithContext(ctx))
    }

    serverPool.AddBackend(&Backend{
      URL:          serverUrl,
      Alive:        true,
      ReverseProxy: proxy,
    })
    log.Printf("Configured server: %s\n", serverUrl)
  }

  server := http.Server{
    Addr:    fmt.Sprintf(":%d", port),
    Handler: http.HandlerFunc(lb),
  
  }

  go healthCheck()

  log.Printf("Load Balancer started at :%d\n", port)
  if err := server.ListenAndServe(); err != nil {
    log.Fatal(err)
  }
}
