package support

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/go-cleanhttp"

	"github.com/onsi/gomega"
	"github.com/sco1237896/sco-backend/cmd/serve"
)

type Test interface {
	T() *testing.T
	Ctx() context.Context

	Serve([]string)
	HTTP() *HTTP

	gomega.Gomega
}

func With(t *testing.T) Test {
	t.Helper()
	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		withDeadline, cancel := context.WithDeadline(ctx, deadline)
		t.Cleanup(cancel)
		ctx = withDeadline
	}

	return &T{
		WithT: gomega.NewWithT(t),
		t:     t,
		ctx:   ctx,
	}
}

type T struct {
	*gomega.WithT

	t *testing.T

	//nolint: containedctx
	ctx context.Context
}

func (t *T) T() *testing.T {
	return t.t
}

func (t *T) Ctx() context.Context {
	return t.ctx
}

func (t *T) Serve(args []string) {
	cmd := serve.NewServeCmd()
	cmd.SetArgs(args)

	go func() {
		err := cmd.ExecuteContext(t.Ctx())
		t.Expect(err).ShouldNot(gomega.HaveOccurred())
	}()
}

func (t *T) HTTP() *HTTP {
	return &HTTP{
		t: t,
	}
}

type HTTP struct {
	t Test
}

func (h *HTTP) GET(url string) func() (*http.Response, error) {
	client := cleanhttp.DefaultClient()

	req, err := http.NewRequestWithContext(h.t.Ctx(), http.MethodGet, url, http.NoBody)
	h.t.Expect(err).ShouldNot(gomega.HaveOccurred())

	return func() (*http.Response, error) {
		return client.Do(req)
	}
}
