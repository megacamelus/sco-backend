package server

import "github.com/gin-gonic/gin"

const version = "/v1"

func (s *Service) routes(engine *gin.Engine) {
	v1 := engine.Group(version)

	// Add routes for pipes
	pipes := v1.Group("/pipes")
	pipes.GET("/", s.getPipes)

	// Add rest of routes
}
