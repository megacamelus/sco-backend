package serve

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sco1237896/sco-backend/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/health"
	"github.com/sco1237896/sco-backend/pkg/logger"
	"github.com/sco1237896/sco-backend/pkg/server"
	"github.com/spf13/cobra"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func NewServeCmd() *cobra.Command {

	logOpts := logger.Options{
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
			logger.Init(logOpts)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.L.Debug("Initializing shutdown channel")
			shutdown := make(chan os.Signal, 1)
			signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

			logger.L.Info("Initializing Health Check server")
			var h *health.Service
			if healthEnabled {
				h := health.New(healthOpts, logger.L)
				go func() {
					if err := h.Start(cmd.Context()); err != nil {
						logger.L.ErrorContext(cmd.Context(), "Error in Health Check service", slog.Any("error", err))
						shutdown <- syscall.SIGTERM
					}
				}()
			}

			logger.L.Info("Initializing SCO client")
			cl, err := client.New(cmd.Context())
			if err != nil {
				return err
			}

			logger.L.Info("Initializing main server")
			s := server.New(*serverOpts, cl, h, logger.L)
			go func() {
				if err := s.Start(cmd.Context()); err != nil {
					logger.L.ErrorContext(cmd.Context(), "Error starting main server", slog.Any("error", err))
					shutdown <- syscall.SIGTERM
				}
			}()

			logger.L.Info("Main thread running until shutdown signal")
			<-shutdown
			logger.L.Info("Main thread is shutting down")

			logger.L.Info("Terminating main server")
			if s != nil {
				if err := s.Stop(cmd.Context()); err != nil {
					logger.L.ErrorContext(cmd.Context(), "Error stopping the main server", slog.Any("error", err))
				}
			}

			logger.L.Info("Terminating Health Check server")
			if h != nil {
				if err := h.Stop(cmd.Context()); err != nil {
					logger.L.ErrorContext(cmd.Context(), "Error stopping the health service", slog.Any("error", err))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverOpts.Addr, "bind-address", serverOpts.Addr, "The address the server binds to.")
	cmd.Flags().BoolVar(&logOpts.Development, "dev", logOpts.Development, "Turn on/off development mode")
	cmd.Flags().BoolVar(&healthEnabled, "health-check-enabled", healthEnabled, "health-check-enabled")
	cmd.Flags().StringVar(&healthOpts.Prefix, "health-check-prefix", healthOpts.Prefix, "health-check-prefix")
	cmd.Flags().StringVar(&healthOpts.Addr, "health-check-address", healthOpts.Addr, "health-check-address")

	return &cmd
}
