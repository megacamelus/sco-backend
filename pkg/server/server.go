package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	camelv1alpha "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/sco1237896/sco-backend/pkg/client"
)

var (
	logger = slog.Default()
	cl     client.Client
)

type Options struct {
	Addr string
	Port int
}

func Start(opts Options, client *client.Client) error {
	r := setupRouter(client)
	logger.Info("starting server")
	err := r.Run(opts.Addr + ":" + strconv.Itoa(opts.Port))
	if err != nil {
		return err
	}

	return nil
}

func setupRouter(client *client.Client) *gin.Engine {
	cl = *client
	r := gin.Default()
	r.GET("/pipes", getPipes)
	return r
}

func getPipes(c *gin.Context) {
	list := &camelv1alpha.KameletBindingList{}
	err := cl.List(context.Background(), list)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, list)
}
