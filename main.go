/*
 * WebX - Under Development
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
  "flag"
  "embed"
  "bytes"
  "io/fs"
  "strings"
  "context"
  "syscall"
  "net/http"
  "os/signal"
  "html/template"
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
  if statusCode != http.StatusOK {
    for _, k := range []string {
      "Cache-Control",
      "ETag",
    } {
      w.ResponseWriter.Header().Del(k)
    }
  }
  w.statusCode = statusCode
  w.ResponseWriter.WriteHeader(w.statusCode)
}

func logRequest(h http.Handler, xffPtr bool) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    _w := responseWriter(w)
    h.ServeHTTP(_w, r)

    host, _, _ := net.SplitHostPort(r.RemoteAddr)
    if xffPtr && r.Header.Get("X-Forwarded-For") != "" {
      host = r.Header.Get("X-Forwarded-For")
    }
    log.Printf("[%s] {%d} %s %s %s\n", host, _w.statusCode, r.Method, r.URL.Path, r.Proto)
  })
}

func wwwHandler(h http.Handler, tmpl *template.Template) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
      r.URL.Path = "/index.html"
    }
    if r.Header.Get("If-None-Match") == Version {
      w.WriteHeader(http.StatusNotModified)

    } else {
      if strings.HasPrefix(r.URL.Path, fmt.Sprintf("/%s/", Version)) {
        r.URL.Path = strings.TrimPrefix(r.URL.Path[1:], Version)
        w.Header().Set("Cache-Control", "max-age=31536000, immutable")

      } else {
        w.Header().Set("Cache-Control", "max-age=0, must-revalidate")
        w.Header().Set("ETag", Version)
      }

      if t := tmpl.Lookup(r.URL.Path[1:]); t != nil {
        var buf bytes.Buffer

        data := map[string]string {
          "Version": Version,
        }

        if err := t.Execute(&buf, data); err == nil {
          w.Write(buf.Bytes())

        } else {
          http.Error(w, err.Error(), http.StatusInternalServerError)
        }
      } else {
        h.ServeHTTP(w, r)
      }
    }
  })
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
  if r.Header.Get("Content-Type") == "application/json" {
    if r.URL.Path == "/api/create" {
      w.Header().Set("Content-Type", "application/json")
      fmt.Fprintf(w, "{}\n")

    } else {
      http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
    }
  } else {
    http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
  }
}

func main() {
  fmt.Fprintf(os.Stdout, "WebX v%s\n", Version)
  fmt.Fprintf(os.Stdout, "Copyright (c) 2025 Chris Mason <chris@netnix.org>\n\n")

  lPtr := flag.String("l", "127.0.0.1", "Listen Address")
  pPtr := flag.Int("p", 8080, "Listen Port")
  xffPtr := flag.Bool("xff", false, "Use X-Forwarded-For")
  flag.Parse()

  sCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
  defer stop()

  mux := http.NewServeMux()
  subFS, _ := fs.Sub(www, "www")
  if tmpl, err := template.ParseFS(subFS, "*.html"); err == nil {
    mux.Handle("GET /", wwwHandler(http.FileServer(http.FS(subFS)), tmpl))
    mux.HandleFunc("POST /api/", apiHandler)

    s := &http.Server {
      Addr: fmt.Sprintf("%s:%d", *lPtr, *pPtr),
      Handler: logRequest(mux, *xffPtr),
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

  } else {
    log.Fatalf("Error: %v\n", err)
  }
}

