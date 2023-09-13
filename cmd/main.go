package main

import (
	"log/slog"
	"os"

	"github.com/sco1237896/sco-backend/cmd/serve"

	"github.com/spf13/cobra"
)

var (
	logger = slog.Default().With(slog.String("component", "main"))
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "sco",
		Short: "sco",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	rootCmd.AddCommand(serve.NewServeCmd())

	if err := rootCmd.Execute(); err != nil {
		logger.Error("problem running command", err)
		os.Exit(1)
	}
}
