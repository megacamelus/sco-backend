// Package expvar manages the publishing of metrics to stdout.
package expvar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"github.com/sco1237896/sco-backend/pkg/logger"
)

// Expvar provide our basic publishing.
type Expvar struct {
	log    *slog.Logger
	Server http.Server
	data   map[string]any
	mu     sync.RWMutex
}

// New starts a service for consuming the raw expvar stats.
func New(log *slog.Logger, host string, route string, readTimeout, writeTimeout time.Duration, idleTimeout time.Duration) *Expvar {
	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(sloggin.New(logger.L.WithGroup("expvar")))

	exp := Expvar{
		log: log,
		Server: http.Server{
			Addr:         host,
			Handler:      router,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}

	router.GET(route, func(c *gin.Context) {
		exp.handler(c.Writer, c.Request, nil)
	})

	return &exp
}

// Stop shuts down the service.
func (exp *Expvar) Stop(shutdownTimeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	exp.log.InfoContext(ctx, "expvar", "status", "start shutdown...")
	defer exp.log.InfoContext(ctx, "expvar: Completed")

	if err := exp.Server.Shutdown(ctx); err != nil {
		exp.log.ErrorContext(ctx, "expvar", "status", "graceful shutdown did not complete", "msg", err, "shutdownTimeout", shutdownTimeout)
		if err := exp.Server.Close(); err != nil {
			exp.log.ErrorContext(ctx, "expvar", "status", "could not stop http Server", "msg", err)
		}
	}
}

// Publish is called by the publisher goroutine and saves the raw stats.
func (exp *Expvar) Publish(data map[string]any) error {
	exp.mu.Lock()
	{
		exp.data = maps.Clone(data)
	}
	exp.mu.Unlock()

	return nil
}

// handler is what consumers call to get the raw stats.
func (exp *Expvar) handler(w http.ResponseWriter, r *http.Request, _ map[string]string) {
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
