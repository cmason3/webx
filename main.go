/*
 * WebX
 * Copyright (c) 2025 Chris Mason <chris@netnix.org>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package main

import (
  "os"
  "fmt"
  "net"
  "log"
  "time"
  "embed"
  "io/fs"
  "context"
  "syscall"
  "net/http"
  "os/signal"
)

var Version = "0.0.1"
//go:embed www
var www embed.FS

type httpWriter struct {
  http.ResponseWriter
  statusCode int
}

func responseWriter(w http.ResponseWriter) *httpWriter {
  return &httpWriter { w, http.StatusOK }
}

func (w *httpWriter) WriteHeader(statusCode int) {
  w.statusCode = statusCode
  w.ResponseWriter.WriteHeader(statusCode)
}

func logRequest(h http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    _w := responseWriter(w)
    h.ServeHTTP(_w, r)
    log.Printf("[%s] %s %s %s {%d}\n", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, _w.statusCode)
  })
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path == "/api/create" {
    fmt.Fprintf(w, "Hello World\r\n")

  } else {
    http.NotFound(w, r)
  }
}

func main() {
  fmt.Fprintf(os.Stdout, "WebX v%s\n", Version)
  fmt.Fprintf(os.Stdout, "Copyright (c) 2025 Chris Mason <chris@netnix.org>\n\n")

  sCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
  defer stop()

  mux := http.NewServeMux()
  subFS, _ := fs.Sub(www, "www")
  mux.Handle("GET /", http.FileServer(http.FS(subFS)))
  mux.HandleFunc("POST /api/", apiHandler)

  s := &http.Server {
    Addr: "0.0.0.0:8080",
    Handler: logRequest(mux),
    BaseContext: func(net.Listener) context.Context {
      return sCtx 
    },
  }

  go func() {
    log.Printf("Starting WebX (PID is %d) on http://%v...\n", os.Getpid(), s.Addr)

    if err := s.ListenAndServe(); err != http.ErrServerClosed {
      log.Fatalf("Error: %v\n", err)
    }
  }()

  <-sCtx.Done()
  log.Printf("Caught Signal... Terminating...\n")
  cCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
  defer cancel()

  s.Shutdown(cCtx)
}

