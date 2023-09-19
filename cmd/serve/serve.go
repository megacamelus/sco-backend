package serve

import (
	"github.com/gin-gonic/gin"
	"github.com/sco1237896/sco-backend/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/health"
	"github.com/sco1237896/sco-backend/pkg/logger"
	"github.com/sco1237896/sco-backend/pkg/server"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func NewServeCmd() *cobra.Command {

	logOpts := logger.Options{
		Development: false,
	}

	options := server.Options{
		Addr: ":8080",
	}

	ho := struct {
		Health        bool
		HealthAddress string
		HealthPrefix  string
	}{
		Health:        true,
		HealthPrefix:  health.DefaultPrefix,
		HealthAddress: health.DefaultAddress,
	}

	cmd := cobra.Command{
		Use:   "serve",
		Short: "serve",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			logger.Init(logOpts)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			done := make(chan os.Signal, 1)
			signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

			var h *health.Service

			if ho.Health {
				h = health.New(ho.HealthAddress, ho.HealthPrefix, logger.L)

				if err := h.Start(cmd.Context()); err != nil {
					return err
				}
			}

			cl, err := client.New(cmd.Context())
			if err != nil {
				return err
			}

			go func() {
				if err := server.Start(options, cl); err != nil {
					panic(err)
				}
			}()

			<-done

			if h != nil {
				if err := h.Stop(cmd.Context()); err != nil {
					logger.L.Error("error stopping the health service", slog.Any("error", err))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&options.Addr, "bind-address", options.Addr, "The address the server binds to.")
	cmd.Flags().BoolVar(&logOpts.Development, "dev", logOpts.Development, "Turn on/off development mode")
	cmd.Flags().BoolVar(&ho.Health, "health-check-enabled", ho.Health, "health-check-enabled")
	cmd.Flags().StringVar(&ho.HealthPrefix, "health-check-prefix", ho.HealthPrefix, "health-check-prefix")
	cmd.Flags().StringVar(&ho.HealthAddress, "health-check-address", ho.HealthAddress, "health-check-address")

	return &cmd
}
