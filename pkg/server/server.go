package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sco1237896/sco-backend/pkg/client"
)

var (
	logger = slog.Default().With(slog.String("component", "server"))
)

type Options struct {
	Addr string
}

func Start(opts Options, cl client.ScoClient) error {
	r := setupRouter(cl)
	logger.Info("starting server")
	err := r.Run(opts.Addr)
	if err != nil {
		return err
	}

	return nil
}

func setupRouter(cl client.ScoClient) *gin.Engine {
	r := gin.Default()
	r.GET("/pipes", func(c *gin.Context) {
		getPipes(cl, c)
	})
	return r
}

func getPipes(cl client.ScoClient, c *gin.Context) {
	list, err := cl.ListPipes(context.Background())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, list)
}
