package client

import (
	"context"
	"log/slog"

	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	camelclient "github.com/apache/camel-k/v2/pkg/client"
	"github.com/pkg/errors"
)

type defaultClient struct {
	camelCl camelclient.Client
	logger  *slog.Logger
}

var _ Interface = &defaultClient{}

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

func (cl *defaultClient) ListPipes(c context.Context) (*camelv1.PipeList, error) {
	list := &camelv1.PipeList{}
	err := cl.camelCl.List(c, list)
	if err != nil {
		return nil, err
	}

	return list, nil
}
