package client

import (
	"context"
	"fmt"
	"log/slog"

	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelclient "github.com/apache/camel-k/pkg/client"
	"github.com/go-logr/logr/slogr"
	controllerlog "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = slog.Default().With(slog.String("component", "k8s-client"))

type Client struct {
	camelclient.Client
}

func NewClient() (*Client, error) {
	controllerlog.SetLogger(slogr.NewLogr(logger.Handler()))

	cl, err := camelclient.NewClient(false)
	if err != nil {
		return nil, fmt.Errorf("failed to create camel k8s client: %w", err)
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
