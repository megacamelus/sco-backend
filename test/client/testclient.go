package client

import (
	"context"

	camelv1alpha "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/sco1237896/sco-backend/pkg/client"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestClient struct {
}

var _ client.Interface = &TestClient{}

func (cl TestClient) ListPipes(_ context.Context) (*camelv1alpha.KameletBindingList, error) {
	list := &camelv1alpha.KameletBindingList{
		Items: []camelv1alpha.KameletBinding{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "mykb1",
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "mykb2",
				},
			},
		},
	}
	return list, nil
}
