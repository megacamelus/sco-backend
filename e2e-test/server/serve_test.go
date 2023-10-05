package server_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sco1237896/sco-backend/cmd/serve"

	"github.com/hashicorp/go-cleanhttp"

	. "github.com/onsi/gomega"
)

func TestServe(t *testing.T) {
	g := NewWithT(t)

	context, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cmd := serve.NewServeCmd()
	cmd.SetArgs([]string{"--bind-address", "localhost:9090"})

	go func() {
		err := cmd.ExecuteContext(context)
		if err != nil {
			t.Error(err)
		}
	}()

	req, err := http.NewRequestWithContext(context, http.MethodGet, "http://localhost:9090/pipes", http.NoBody)
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

func TestMainHealthCheck(t *testing.T) {
	g := NewWithT(t)

	sharedAddr := "localhost:9091"

	c, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// start first server
	cmd := serve.NewServeCmd()
	cmd.SetArgs([]string{
		"--health-check-enabled",
		"--bind-address", sharedAddr,
		"--debug-bind-address", "localhost:9092",
		"--health-check-address", "localhost:9093",
	})
	go func() {
		err := cmd.ExecuteContext(c)
		if err != nil {
			t.Error(err)
		}
	}()

	// tries to start second server in the same addr
	cmd2 := serve.NewServeCmd()
	cmd2.SetArgs([]string{
		"--health-check-enabled",
		"--bind-address", sharedAddr,
		"--debug-bind-address", "localhost:9082",
		"--health-check-address", "localhost:9083",
	})
	go func() {
		err := cmd2.ExecuteContext(c)
		if err != nil {
			t.Error(err)
		}
	}()

	// check that first server is ready
	req, err := http.NewRequestWithContext(c, http.MethodGet, "http://localhost:9093/health/ready", http.NoBody)
	if err != nil {
		t.Error(err)
	}
	g.Eventually(cleanhttp.DefaultClient().Do).WithArguments(req).Should(HaveHTTPStatus(http.StatusOK))

	// check debug endpoint
	reqDebug, err := http.NewRequestWithContext(c, http.MethodGet, "http://localhost:9092/debug/vars", http.NoBody)
	if err != nil {
		t.Error(err)
	}
	g.Eventually(cleanhttp.DefaultClient().Do).WithArguments(reqDebug).Should(HaveHTTPStatus(http.StatusOK))

	// check that first server is ready
	req, err = http.NewRequestWithContext(c, http.MethodGet, "http://localhost:9083/health/ready?full=true", http.NoBody)
	if err != nil {
		t.Error(err)
	}
	g.Consistently(cleanhttp.DefaultClient().Do).WithArguments(req).Within(1 * time.Second).Should(HaveHTTPStatus(http.StatusServiceUnavailable))
}
