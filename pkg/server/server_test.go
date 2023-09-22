package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sco1237896/sco-backend/pkg/logger"

	"github.com/sco1237896/sco-backend/test/client"

	camelv1alpha "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/json"
)

func TestGetPipes(t *testing.T) {
	logger.Init(true)

	serverOpts := DefaultOptions()
	server := New(*serverOpts, &client.TestClient{}, nil, logger.L)

	router := server.svr.Handler
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/pipes", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	list := camelv1alpha.KameletBindingList{}
	err := json.Unmarshal(
		w.Body.Bytes(),
		&list,
	)
	if err != nil {
		return
	}

	assert.Equal(t, "mykb1", list.Items[0].Name)
	assert.Equal(t, "mykb2", list.Items[1].Name)
}
