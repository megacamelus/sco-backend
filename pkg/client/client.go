package client

import (
	"context"
	"fmt"
	"log/slog"

	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelv1alpha "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	camelclient "github.com/apache/camel-k/pkg/client"
	"github.com/go-logr/logr/slogr"
	controllerlog "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = slog.Default().With(slog.String("component", "k8s-client"))

type Interface interface {
	ListPipes(c context.Context) (*camelv1alpha.KameletBindingList, error)
}

type defaultClient struct {
	camelCl camelclient.Client
}

var _ Interface = &defaultClient{}

func New(c context.Context) (Interface, error) {
	controllerlog.SetLogger(slogr.NewLogr(logger.Handler()))

	cl, err := camelclient.NewClient(false)
	if err != nil {
		return nil, fmt.Errorf("failed to create camel k8s client: %w", err)
	}

	ip := &camelv1.IntegrationPlatformList{}
	err = cl.List(c, ip)
	if err != nil {
		logger.Warn("Failed to find IntegrationPlatform.", err)
	}
	if len(ip.Items) == 0 {
		logger.Warn("Failed to find IntegrationPlatform. Is Camel K running?")
	}

	return &defaultClient{cl}, nil
}

func (cl *defaultClient) ListPipes(c context.Context) (*camelv1alpha.KameletBindingList, error) {
	list := &camelv1alpha.KameletBindingList{}
	err := cl.camelCl.List(c, list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
