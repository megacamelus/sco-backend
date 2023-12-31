// Package prometheus provides suppoert for sending metrics to prometheus.
package prometheus

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"github.com/sco1237896/sco-backend/pkg/logger"
)

// Exporter implements the prometheus exporter support.
type Exporter struct {
	log    *slog.Logger
	Server http.Server
	data   map[string]any
	mu     sync.RWMutex
}

// New constructs an Exporter for use.
func New(log *slog.Logger, host string, route string, readTimeout, writeTimeout time.Duration, idleTimeout time.Duration) *Exporter {
	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(sloggin.New(logger.L.WithGroup("prometheus")))

	exp := Exporter{
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
		exp.handler(c.Writer, c.Request)
	})

	return &exp
}

// Publish stores a deep copy of the data for publishing.
func (exp *Exporter) Publish(data map[string]any) error {
	exp.mu.Lock()
	defer exp.mu.Unlock()

	exp.data = deepCopyMap(data)

	return nil
}

// Stop turns off all the prometheus support.
func (exp *Exporter) Stop(shutdownTimeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	exp.log.InfoContext(ctx, "prometheus", "status", "start shutdown...")
	defer exp.log.InfoContext(ctx, "prometheus: Completed")

	if err := exp.Server.Shutdown(ctx); err != nil {
		exp.log.ErrorContext(ctx, "prometheus", "status", "graceful shutdown did not complete", "msg", err, "shutdownTimeout", shutdownTimeout)

		if err := exp.Server.Close(); err != nil {
			exp.log.ErrorContext(ctx, "prometheus", "status", "could not stop http server", "msg", err)
		}
	}
}

// =============================================================================

func (exp *Exporter) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)

	var data map[string]any
	exp.mu.Lock()
	{
		data = deepCopyMap(exp.data)
	}
	exp.mu.Unlock()

	out(w, "", data)

	exp.log.Info("prometheus", "metrics", fmt.Sprintf("expvar : (%d) : %s %s -> %s", http.StatusOK, r.Method, r.URL.Path, r.RemoteAddr))
}

// =============================================================================

func deepCopyMap(source map[string]any) map[string]any {
	result := make(map[string]any)

	for k, v := range source {
		switch vm := v.(type) {
		case map[string]any:
			result[k] = deepCopyMap(vm)

		case int64:
			result[k] = float64(vm)

		case float64:
			result[k] = vm

		case bool:
			result[k] = 0.0
			if vm {
				result[k] = 1.0
			}
		default:
			// do nothing
		}
	}

	return result
}

func out(w io.Writer, prefix string, data map[string]any) {
	if prefix != "" {
		prefix += "_"
	}

	for k, v := range data {
		writeKey := fmt.Sprintf("%s%s", prefix, k)

		switch vm := v.(type) {
		case float64:
			fmt.Fprintf(w, "%s %.f\n", writeKey, vm)

		case map[string]any:
			out(w, writeKey, vm)

		default:
			// Discard this value.
		}
	}
}
