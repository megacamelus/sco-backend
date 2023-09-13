package server

import (
	"context"
	"log/slog"
	"net/http"

	camelv1alpha "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/sco1237896/sco-backend/pkg/client"
)

var (
	logger = slog.Default().With(slog.String("component", "server"))
)

type Options struct {
	Addr string
}

func Start(opts Options, cl *client.Client) error {
	r := setupRouter(cl)
	logger.Info("starting server")
	err := r.Run(opts.Addr)
	if err != nil {
		return err
	}

	return nil
}

func setupRouter(cl *client.Client) *gin.Engine {
	r := gin.Default()
	r.GET("/pipes", func(c *gin.Context) {
		getPipes(cl, c)
	})
	return r
}

func getPipes(cl *client.Client, c *gin.Context) {
	list := &camelv1alpha.KameletBindingList{}
	err := cl.List(context.Background(), list)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, list)
}
