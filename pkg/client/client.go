package client

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	"github.com/pkg/errors"

	camelv1alpha "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	camelclient "github.com/apache/camel-k/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/logger"
)

var (
	once func() (Interface, error)
)

type Interface interface {
	Check(c context.Context) error
	ListPipes(c context.Context) (*camelv1alpha.KameletBindingList, error)
}

type defaultClient struct {
	camelCl camelclient.Client
	logger  *slog.Logger
}

var _ Interface = &defaultClient{}

func init() {
	once = sync.OnceValues(func() (Interface, error) {
		l := logger.With(slog.String("component", "k8s-client"))

		cl, err := camelclient.NewClient(false)
		if err != nil {
			return nil, fmt.Errorf("failed to create camel k8s client: %w", err)
		}

		return &defaultClient{camelCl: cl, logger: l}, nil
	})
}

func GetInstance() (Interface, error) {
	return once()
}

func (cl *defaultClient) Check(c context.Context) error {
	ip := &camelv1.IntegrationPlatformList{}
	err := cl.camelCl.List(c, ip)
	if err != nil {
		return errors.Wrap(err, "failed to find IntegrationPlatform")
	}
	if len(ip.Items) == 0 {
		return errors.Wrap(err, "failed to find IntegrationPlatform. Is Camel K running?")
	}
	return nil
}

func (cl *defaultClient) ListPipes(c context.Context) (*camelv1alpha.KameletBindingList, error) {
	list := &camelv1alpha.KameletBindingList{}
	err := cl.camelCl.List(c, list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
