package health

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/expvar"
	"github.com/gin-contrib/pprof"

	"github.com/gin-gonic/gin"
)

type Options struct {
	Enabled         bool
	Addr            string
	Prefix          string
	ShutdownTimeout time.Duration
}

type Service struct {
	l       *slog.Logger
	running atomic.Bool
	router  *gin.Engine
	srv     *http.Server
	opts    Options

	checksMutex     sync.RWMutex
	livenessChecks  map[string]Check
	readinessChecks map[string]Check
}

type Check func() error

func DefaultOptions() Options {
	return Options{
		Enabled:         false,
		Addr:            ":8081",
		Prefix:          "",
		ShutdownTimeout: 3 * time.Second,
	}
}

func New(opts Options, logger *slog.Logger) *Service {
	s := Service{
		opts: opts,
	}
	s.l = logger.WithGroup("health")

	s.router = gin.New()
	s.router.Use(s.log)
	s.router.Use(gin.Recovery())
	s.router.GET(path.Join(opts.Prefix, "/health", "/ready"), s.ready)
	s.router.GET(path.Join(opts.Prefix, "/health", "/live"), s.live)

	// register pprof middleware endpoints
	pprof.Register(s.router)

	// register expvar endpoints
	s.router.GET("/debug/vars", expvar.Handler())

	s.srv = &http.Server{
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Addr:              opts.Addr,
		Handler:           s.router,
	}

	s.readinessChecks = make(map[string]Check)
	s.livenessChecks = make(map[string]Check)

	return &s
}

func (s *Service) Start(context.Context) error {
	if s.running.CompareAndSwap(false, true) {
		err := s.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.running.CompareAndSwap(true, false)
			return err
		}
	}

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	if s.running.CompareAndSwap(true, false) {
		tctx, cancel := context.WithTimeout(ctx, s.opts.ShutdownTimeout)
		defer cancel()

		if err := s.srv.Shutdown(tctx); err != nil {
			s.srv.Close()
			return fmt.Errorf("could not stop health server gracefully: %w", err)
		}
	}

	return nil
}

func (s *Service) AddLivenessCheck(name string, check Check) {
	s.checksMutex.Lock()
	defer s.checksMutex.Unlock()

	s.livenessChecks[name] = check
}

func (s *Service) RemoveLivenessCheck(name string) {
	s.checksMutex.Lock()
	defer s.checksMutex.Unlock()

	delete(s.livenessChecks, name)
}

func (s *Service) AddReadinessCheck(name string, check Check) {
	s.checksMutex.Lock()
	defer s.checksMutex.Unlock()

	s.readinessChecks[name] = check
}

func (s *Service) RemoveReadinessCheck(name string) {
	s.checksMutex.Lock()
	defer s.checksMutex.Unlock()

	delete(s.readinessChecks, name)
}

func (s *Service) ready(c *gin.Context) {
	s.handle(c, s.readinessChecks)
}
func (s *Service) live(c *gin.Context) {
	s.handle(c, s.livenessChecks)
}

func (s *Service) handle(c *gin.Context, checks ...map[string]Check) {
	checkResults := make(map[string]string)
	status := http.StatusOK

	for _, checks := range checks {
		s.collectChecks(checks, checkResults, &status)
	}

	switch c.Query("full") {
	case "true":
		c.JSON(status, gin.H{
			"status": "OK",
			"data":   checkResults,
		})
	default:
		c.JSON(status, gin.H{
			"status": "OK",
		})
	}
}

func (s *Service) collectChecks(checks map[string]Check, resultsOut map[string]string, statusOut *int) {
	s.checksMutex.RLock()
	defer s.checksMutex.RUnlock()

	for name, check := range checks {
		if err := check(); err != nil {
			*statusOut = http.StatusServiceUnavailable
			resultsOut[name] = err.Error()
		} else {
			resultsOut[name] = "OK"
		}
	}
}

func (s *Service) log(c *gin.Context) {
	start := time.Now()

	// some evil middlewares modify this values
	urlPath := c.Request.URL.Path
	urlQuery := c.Request.URL.RawQuery

	c.Next()

	end := time.Now()
	latency := end.Sub(start)

	fields := []any{
		slog.Int("status", c.Writer.Status()),
		slog.String("method", c.Request.Method),
		slog.String("path", urlPath),
		slog.String("query", urlQuery),
		slog.String("ip", c.ClientIP()),
		slog.String("user-agent", c.Request.UserAgent()),
		slog.Duration("latency", latency),
	}

	if len(c.Errors) > 0 {
		for _, e := range c.Errors.Errors() {
			s.l.ErrorContext(c.Request.Context(), e, fields...)
		}
	} else {
		s.l.DebugContext(c.Request.Context(), urlPath, fields...)
	}
}
