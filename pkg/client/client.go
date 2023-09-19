package client

import (
	"context"
	"fmt"
	"log/slog"

	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelv1alpha "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	camelclient "github.com/apache/camel-k/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/logger"
)

type Interface interface {
	ListPipes(c context.Context) (*camelv1alpha.KameletBindingList, error)
}

type defaultClient struct {
	camelCl camelclient.Client
	logger  *slog.Logger
}

var _ Interface = &defaultClient{}

func New(c context.Context) (Interface, error) {
	l := logger.With(slog.String("component", "k8s-client"))

	cl, err := camelclient.NewClient(false)
	if err != nil {
		return nil, fmt.Errorf("failed to create camel k8s client: %w", err)
	}

	ip := &camelv1.IntegrationPlatformList{}
	err = cl.List(c, ip)
	if err != nil {
		l.Warn("Failed to find IntegrationPlatform.", err)
	}
	if len(ip.Items) == 0 {
		l.Warn("Failed to find IntegrationPlatform. Is Camel K running?")
	}

	return &defaultClient{camelCl: cl, logger: l}, nil
}

func (cl *defaultClient) ListPipes(c context.Context) (*camelv1alpha.KameletBindingList, error) {
	list := &camelv1alpha.KameletBindingList{}
	err := cl.camelCl.List(c, list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
