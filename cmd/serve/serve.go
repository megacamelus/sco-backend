package serve

import (
	"context"

	"github.com/sco1237896/sco-backend/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/server"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {

	options := server.Options{
		Addr: ":8080",
	}

	cmd := cobra.Command{
		Use:   "serve",
		Short: "serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := client.NewClient(context.Background())
			if err != nil {
				return err
			}

			return server.Start(options, cl)
		},
	}

	cmd.Flags().StringVar(&options.Addr, "bind-address", options.Addr, "The address the server binds to.")

	return &cmd
}
