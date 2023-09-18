package server_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"knative.dev/eventing/third_party/VENDOR-LICENSE/github.com/hashicorp/go-cleanhttp"

	"github.com/sco1237896/sco-backend/cmd/serve"
	"github.com/stretchr/testify/assert"
)

func TestServe(t *testing.T) {
	cmd := serve.NewServeCmd()
	cmd.SetArgs([]string{"--bind-address", "localhost:9090"})

	go func() {
		err := cmd.ExecuteContext(context.Background())
		if err != nil {
			t.Error(err)
		}
	}()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:9090/pipes", http.NoBody)
	if err != nil {
		t.Error(err)
	}

	assertEventually(t, func() bool {
		resp, err := cleanhttp.DefaultClient().Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		return http.StatusOK == resp.StatusCode
	})
}

// wrapper for assert.Eventually.
func assertEventually(t *testing.T, condition func() bool) {
	t.Helper()

	waitD, err := time.ParseDuration("2s")
	if err != nil {
		t.Error(err)
	}

	tickD, err := time.ParseDuration("200ms")
	if err != nil {
		t.Error(err)
	}

	assert.Eventually(t, condition, waitD, tickD)
}
