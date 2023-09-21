package server_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/go-cleanhttp"

	. "github.com/onsi/gomega"

	"github.com/sco1237896/sco-backend/cmd/serve"
)

func TestServe(t *testing.T) {
	g := NewWithT(t)

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

	got := func() (int, error) {
		resp, err := cleanhttp.DefaultClient().Do(req)
		if err != nil {
			return -1, err
		}
		defer resp.Body.Close()
		return resp.StatusCode, nil
	}

	g.Eventually(got).Should(Equal(http.StatusOK))
}
