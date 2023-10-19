package client

import (
	"context"

	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/sco1237896/sco-backend/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestClient struct {
}

var _ client.Interface = &TestClient{}

func (cl TestClient) ListPipes(_ context.Context) (*camelv1.PipeList, error) {
	list := &camelv1.PipeList{
		Items: []camelv1.Pipe{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mykb1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mykb2",
				},
			},
		},
	}
	return list, nil
}

func (cl TestClient) Check(context.Context) error {
	return nil
}
