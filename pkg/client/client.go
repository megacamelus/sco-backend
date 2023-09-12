package client

import (
	"context"
	"log/slog"
	"os"

	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelclient "github.com/apache/camel-k/pkg/client"
	"github.com/go-logr/logr/slogr"
	controllerlog "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = slog.Default()

type Client struct {
	camelclient.Client
}

func NewClient() (*Client, error) {
	controllerlog.SetLogger(slogr.NewLogr(logger.Handler()))

	cl, err := camelclient.NewClient(false)
	if err != nil {
		logger.Error("failed to create k8s client for camel k", err)
		os.Exit(1)
	}

	ip := &camelv1.IntegrationPlatformList{}
	err = cl.List(context.Background(), ip)
	if err != nil {
		logger.Warn("Failed to find IntegrationPlatform.", err)
	}
	if len(ip.Items) == 0 {
		logger.Warn("Failed to find IntegrationPlatform. Is Camel K running?")
	}

	return &Client{cl}, nil
}
