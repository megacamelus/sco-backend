package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sco1237896/sco-backend/pkg/connectors"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"github.com/sco1237896/sco-backend/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/health"
	"github.com/sco1237896/sco-backend/pkg/logger"
)

type Options struct {
	Addr              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

type Service struct {
	opts    *Options
	l       *slog.Logger
	cl      client.Interface
	catalog *connectors.Catalog
	health  *health.Service
	svr     *http.Server
	running atomic.Bool
}

func DefaultOptions() Options {
	return Options{
		Addr:              ":8080",
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	}
}

func New(opts Options, cl client.Interface, catalog *connectors.Catalog, health *health.Service, l *slog.Logger) *Service {
	l = l.WithGroup("server")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(sloggin.New(l))

	svr := &http.Server{
		ReadTimeout:       opts.ReadTimeout,
		WriteTimeout:      opts.WriteTimeout,
		IdleTimeout:       opts.IdleTimeout,
		ReadHeaderTimeout: opts.ReadHeaderTimeout,
		Addr:              opts.Addr,
		Handler:           r,
		ErrorLog:          slog.NewLogLogger(l.Handler(), slog.LevelError),
	}

	s := &Service{
		l:       logger.With(slog.String("component", "server")),
		cl:      cl,
		catalog: catalog,
		health:  health,
		opts:    &opts,
		svr:     svr,
	}

	s.routes(r)

	return s
}

func (s *Service) Start(c context.Context) error {
	s.l.InfoContext(c, "Starting server")

	if s.health != nil {
		s.health.AddReadinessCheck(s.serverName()+"1", func() error {
			if s.running.Load() {
				return nil

			}
			return errors.New(s.serverName() + " is not running")
		})
		s.health.AddReadinessCheck(s.serverName()+"2", func() error {
			return s.cl.Check(c)
		})
	}

	if s.running.CompareAndSwap(false, true) {
		err := s.svr.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.running.CompareAndSwap(true, false)
			return err
		}
	}

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	if s.health != nil {
		s.health.RemoveReadinessCheck(s.serverName())
	}

	if s.running.CompareAndSwap(true, false) {
		tctx, cancel := context.WithTimeout(ctx, s.opts.ShutdownTimeout)
		defer cancel()

		if err := s.svr.Shutdown(tctx); err != nil {
			s.svr.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func (s *Service) serverName() string {
	return "server at " + s.opts.Addr
}
