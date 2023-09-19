package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sco1237896/sco-backend/pkg/client"
	"github.com/sco1237896/sco-backend/pkg/logger"
)

type Options struct {
	Addr string
}

func Start(opts Options, cl client.Interface) error {
	l := logger.With(slog.String("component", "server"))
	r := setupRouter(cl)

	l.Info("starting server")

	err := r.Run(opts.Addr)
	if err != nil {
		return err
	}

	return nil
}

func setupRouter(cl client.Interface) *gin.Engine {
	r := gin.Default()

	r.GET("/pipes", func(c *gin.Context) {
		getPipes(cl, c)
	})

	return r
}

func getPipes(cl client.Interface, c *gin.Context) {
	list, err := cl.ListPipes(context.Background())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, list)
}
