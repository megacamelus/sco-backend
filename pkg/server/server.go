package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

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
}

type Service struct {
	opts    *Options
	l       *slog.Logger
	cl      client.Interface
	health  *health.Service
	svr     *http.Server
	running atomic.Bool
}

func DefaultOptions() *Options {
	return &Options{
		Addr:              ":8080",
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
}

func New(opts Options, cl client.Interface, health *health.Service, l *slog.Logger) *Service {
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
		l:      logger.With(slog.String("component", "server")),
		cl:     cl,
		health: health,
		opts:   &opts,
		svr:    svr,
	}

	r.GET("/pipes", s.getPipes)

	return s
}

func (s *Service) Start(c context.Context) error {
	s.l.InfoContext(c, "starting server")

	if s.running.CompareAndSwap(false, true) {
		err := s.svr.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	if s.running.CompareAndSwap(true, false) {
		return s.svr.Shutdown(ctx)
	}

	return nil
}

func (s *Service) getPipes(c *gin.Context) {
	list, err := s.cl.ListPipes(context.Background())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, list)
}
