package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/apache/camel-k/pkg/util/test"
	"github.com/sco1237896/sco-backend/pkg/client"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

func TestGetPipes(t *testing.T) {
	clientset, err := test.NewFakeClient(
		&v1alpha1.KameletBindingList{
			Items: []v1alpha1.KameletBinding{
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
		})

	if err != nil {
		t.Fatal(err)
	}

	router := setupRouter(&client.Client{Client: clientset})
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/pipes", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	list := v1alpha1.KameletBindingList{}
	err = json.Unmarshal(
		w.Body.Bytes(),
		&list,
	)
	if err != nil {
		return
	}

	assert.Equal(t, "mykb1", list.Items[0].Name)
	assert.Equal(t, "mykb2", list.Items[1].Name)
}
