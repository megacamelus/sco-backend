// Package expvar manages the publishing of metrics to stdout.
package expvar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/dimfeld/httptreemux/v5"
)

// Expvar provide our basic publishing.
type Expvar struct {
	log    *slog.Logger
	server http.Server
	data   map[string]any
	mu     sync.Mutex
}

// New starts a service for consuming the raw expvar stats.
func New(log *slog.Logger, host string, route string, readTimeout, writeTimeout time.Duration, idleTimeout time.Duration) *Expvar {
	mux := httptreemux.New()
	exp := Expvar{
		log: log,
		server: http.Server{
			Addr:         host,
			Handler:      mux,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}

	mux.Handle("GET", route, exp.handler)

	go func() {
		ctx := context.Background()

		log.InfoContext(ctx, "expvar", "status", "API listening", "host", host)
		if err := exp.server.ListenAndServe(); err != nil {
			log.ErrorContext(ctx, "expvar", "msg", err)
		}
	}()

	return &exp
}

// Stop shuts down the service.
func (exp *Expvar) Stop(shutdownTimeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	exp.log.InfoContext(ctx, "expvar", "status", "start shutdown...")
	defer exp.log.InfoContext(ctx, "expvar: Completed")

	if err := exp.server.Shutdown(ctx); err != nil {
		exp.log.ErrorContext(ctx, "expvar", "status", "graceful shutdown did not complete", "msg", err, "shutdownTimeout", shutdownTimeout)
		if err := exp.server.Close(); err != nil {
			exp.log.ErrorContext(ctx, "expvar", "status", "could not stop http server", "msg", err)
		}
	}
}

// Publish is called by the publisher goroutine and saves the raw stats.
func (exp *Expvar) Publish(data map[string]any) {
	exp.mu.Lock()
	{
		exp.data = data
	}
	exp.mu.Unlock()
}

// handler is what consumers call to get the raw stats.
func (exp *Expvar) handler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	ctx := context.Background()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var data map[string]any
	exp.mu.Lock()
	{
		data = exp.data
	}
	exp.mu.Unlock()

	if err := json.NewEncoder(w).Encode(data); err != nil {
		exp.log.ErrorContext(ctx, "expvar", "status", "encoding data", "msg", err)
	}

	exp.log.InfoContext(ctx, "expvar", "metrics", fmt.Sprintf("(%d) : %s %s -> %s", http.StatusOK, r.Method, r.URL.Path, r.RemoteAddr))
}
