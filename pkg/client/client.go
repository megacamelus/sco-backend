package client

import (
	"context"
	"fmt"
	"log/slog"

	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	camelclient "github.com/apache/camel-k/v2/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/logger"
)

type Interface interface {
	Check(c context.Context) error
	ListPipes(c context.Context) (*camelv1.PipeList, error)
}

func New() (Interface, error) {
	l := logger.With(slog.String("component", "k8s-client"))

	cl, err := camelclient.NewClient(false)
	if err != nil {
		return nil, fmt.Errorf("failed to create camel k8s client: %w", err)
	}

	return &defaultClient{camelCl: cl, logger: l}, nil
}
