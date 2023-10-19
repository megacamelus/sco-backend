package server_test

import (
	"net/http"
	"path"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sco1237896/sco-backend/test/support"
	. "github.com/sco1237896/sco-backend/test/support/matechers"
)

func TestCatalog(t *testing.T) {
	g := support.With(t)

	root := support.GetProjectRootDir()

	g.Serve([]string{
		"--health-check-enabled",
		"--bind-address", "localhost:9090",
		"--health-check-address", "localhost:9091",
		"--connector-catalog-dirs", path.Join(root, "e2e-test/data/connectors"),
	})

	//nolint:bodyclose
	g.Eventually(g.HTTP().GET("http://localhost:9090/v1/connector_types/")).Should(And(
		HaveHTTPStatus(http.StatusOK),
		HaveHTTPBody(
			And(
				WithTransform(ExtractJQ(".items | length"), Equal(10)),
				WithTransform(ExtractJQ(".total"), Equal(float64(12))),
			)),
	))

	//nolint:bodyclose
	g.Eventually(g.HTTP().GET("http://localhost:9090/v1/connector_types/aws_cloudwatch_sink_v1")).Should(And(
		HaveHTTPStatus(http.StatusOK),
		HaveHTTPBody(MatchJQ(`.metadata.name == "aws_cloudwatch_sink_v1"`)),
	))
}
