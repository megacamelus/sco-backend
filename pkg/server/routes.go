package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const version = "/v1"

func (s *Service) routes(engine *gin.Engine) {
	v1 := engine.Group(version)

	// Add routes for pipes
	// TODO: remove
	pipes := v1.Group("/pipes")
	pipes.GET("/", s.getPipes)

	s.connectorTypes(v1)

	// Add rest of routes
}

func (s *Service) getPipes(c *gin.Context) {
	list, err := s.cl.ListPipes(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, list)
}
