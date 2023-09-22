package serve

import (
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/sco1237896/sco-backend/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/health"
	"github.com/sco1237896/sco-backend/pkg/logger"
	"github.com/sco1237896/sco-backend/pkg/server"
	"github.com/spf13/cobra"
)

type Options struct {
	Development bool
}

func NewServeCmd() *cobra.Command {

	opts := Options{
		Development: false,
	}

	serverOpts := server.DefaultOptions()

	healthEnabled := true
	healthOpts := health.Options{
		Prefix: health.DefaultPrefix,
		Addr:   health.DefaultAddress,
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
			logger.L.Debug("Initializing shutdown channel")
			c, _ := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)

			var h *health.Service
			if healthEnabled {
				logger.L.Info("Initializing Health Check server")
				h = health.New(healthOpts, logger.L)
				go func() {
					if err := h.Start(c); err != nil {
						logger.L.ErrorContext(c, "Error in Health Check service", slog.Any("error", err))
					}
				}()
			}

			logger.L.Info("Initializing SCO client")
			cl, err := client.GetInstance()
			if err != nil {
				return err
			}

			logger.L.Info("Initializing main server")
			s := server.New(*serverOpts, cl, h, logger.L)
			go func() {
				if err := s.Start(c); err != nil {
					logger.L.ErrorContext(c, "Error starting main server", slog.Any("error", err))
				}
			}()

			logger.L.Info("Main thread running until shutdown signal")
			<-c.Done()
			logger.L.Info("Main thread is shutting down")

			logger.L.Info("Terminating main server")
			if s != nil {
				if err := s.Stop(c); err != nil {
					logger.L.ErrorContext(c, "Error stopping the main server", slog.Any("error", err))
				}
			}

			if h != nil {
				logger.L.Info("Terminating Health Check server")
				if err := h.Stop(c); err != nil {
					logger.L.ErrorContext(c, "Error stopping the health service", slog.Any("error", err))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverOpts.Addr, "bind-address", serverOpts.Addr, "The address the server binds to.")
	cmd.Flags().BoolVar(&opts.Development, "dev", opts.Development, "Turn on/off development mode")
	cmd.Flags().BoolVar(&healthEnabled, "health-check-enabled", healthEnabled, "health-check-enabled")
	cmd.Flags().StringVar(&healthOpts.Prefix, "health-check-prefix", healthOpts.Prefix, "health-check-prefix")
	cmd.Flags().StringVar(&healthOpts.Addr, "health-check-address", healthOpts.Addr, "health-check-address")

	return &cmd
}
