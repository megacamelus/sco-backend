package serve

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sloggin "github.com/samber/slog-gin"

	"github.com/gin-contrib/expvar"
	"github.com/gin-contrib/pprof"

	"github.com/gin-gonic/gin"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/sco1237896/sco-backend/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/health"
	"github.com/sco1237896/sco-backend/pkg/logger"
	"github.com/sco1237896/sco-backend/pkg/server"
	"github.com/spf13/cobra"
)

type Options struct {
	Development bool
}

type ServerOptions struct {
	Addr              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

func NewServeCmd() *cobra.Command {
	opts := Options{
		Development: false,
	}

	serverOpts := server.DefaultOptions()
	healthOpts := health.DefaultOptions()
	debugOpts := ServerOptions{
		Addr:              ":8082",
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	}

	cmd := cobra.Command{
		Use:   "serve",
		Short: "serve",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			logger.Init(opts.Development)
			if !opts.Development {
				gin.SetMode(gin.ReleaseMode)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// -------------------------------------------------------------------------
			// GOMAXPROCS
			_, err := maxprocs.Set(maxprocs.Logger(func(f string, a ...interface{}) { logger.L.Info(fmt.Sprintf(f, a)) }))
			if err != nil {
				logger.L.ErrorContext(ctx, "failed to set GOMAXPROCS from cgroups")
			}

			// -------------------------------------------------------------------------
			// Print config to stdout
			logger.L.Info("startup", "server config", serverOpts, "health config", healthOpts)

			shutdown := make(chan os.Signal, 1)
			signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

			// -------------------------------------------------------------------------
			// Start Debug Service
			router := gin.Default()
			router.Use(gin.Recovery())
			router.Use(sloggin.New(logger.L.WithGroup("debug")))

			// register pprof middleware endpoints
			pprof.Register(router)

			// register expvar endpoints
			router.GET("/debug/vars", expvar.Handler())

			go func() {
				logger.L.InfoContext(ctx, "startup", "status", "debug v1 router started", "host", debugOpts.Addr)

				srv := http.Server{
					Addr:         debugOpts.Addr,
					Handler:      router,
					ReadTimeout:  debugOpts.ReadTimeout,
					IdleTimeout:  debugOpts.IdleTimeout,
					WriteTimeout: debugOpts.WriteTimeout,
				}

				err = srv.ListenAndServe()
				if err != nil {
					logger.L.ErrorContext(ctx, "shutdown", "status", "debug v1 router closed", "host", debugOpts.Addr, "msg", err)
				}
			}()

			// -------------------------------------------------------------------------
			// Initialize health service
			var h *health.Service
			if healthOpts.Enabled {
				logger.L.Info("Initializing Health Check server")
				h = health.New(healthOpts, logger.L)
				go func() {
					if err := h.Start(ctx); err != nil {
						logger.L.ErrorContext(ctx, "error in Health Check service", slog.Any("error", err))
					}
				}()
			}

			// -------------------------------------------------------------------------
			// Initialize client
			logger.L.Info("Initializing SCO client")
			cl, err := client.GetInstance()
			if err != nil {
				return err
			}

			// -------------------------------------------------------------------------
			// Initialize backend service
			logger.L.Info("Initializing main server")
			s := server.New(serverOpts, cl, h, logger.L)
			go func() {
				if err = s.Start(ctx); err != nil {
					logger.L.ErrorContext(ctx, "error starting main server", slog.Any("error", err))
				}
			}()

			logger.L.Info("Main thread running until shutdown signal")

			// -------------------------------------------------------------------------
			// Start shutdown sequence
			sig := <-shutdown
			logger.L.Info("Main thread is shutting down")
			defer logger.L.Info("Main thread shutdown", "status", "shutdown complete", "signal", sig)

			logger.L.Info("Terminating main server")
			if s != nil {
				if err := s.Stop(ctx); err != nil {
					logger.L.ErrorContext(ctx, "error stopping the main server", slog.Any("error", err))
				}
			}

			if h != nil {
				logger.L.Info("Terminating Health Check server")
				if err := h.Stop(ctx); err != nil {
					logger.L.ErrorContext(ctx, "error stopping the health service", slog.Any("error", err))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverOpts.Addr, "bind-address", serverOpts.Addr, "The address the server binds to.")
	cmd.Flags().BoolVar(&opts.Development, "dev", opts.Development, "Turn on/off development mode")
	cmd.Flags().BoolVar(&healthOpts.Enabled, "health-check-enabled", healthOpts.Enabled, "health-check-enabled")
	cmd.Flags().StringVar(&healthOpts.Prefix, "health-check-prefix", healthOpts.Prefix, "health-check-prefix")
	cmd.Flags().StringVar(&healthOpts.Addr, "health-check-address", healthOpts.Addr, "health-check-address")
	cmd.Flags().StringVar(&debugOpts.Addr, "debug-bind-address", debugOpts.Addr, "debug-bind-address")

	return &cmd
}
