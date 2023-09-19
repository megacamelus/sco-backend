package serve

import (
	"github.com/sco1237896/sco-backend/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/logger"
	"github.com/sco1237896/sco-backend/pkg/server"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {

	logOpts := logger.Options{
		Development: false,
	}

	options := server.Options{
		Addr: ":8080",
	}

	cmd := cobra.Command{
		Use:   "serve",
		Short: "serve",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			logger.Init(logOpts)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := client.New(cmd.Context())
			if err != nil {
				return err
			}

			return server.Start(options, cl)
		},
	}

	cmd.Flags().StringVar(&options.Addr,
		"bind-address",
		options.Addr,
		"The address the server binds to.")

	cmd.Flags().BoolVar(
		&logOpts.Development,
		"dev",
		logOpts.Development,
		"Turn on/off development mode (dev: handler=text,level=debug, prod: handler=json,level=info")

	return &cmd
}
