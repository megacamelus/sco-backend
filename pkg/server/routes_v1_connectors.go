package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pagination "github.com/webstradev/gin-pagination"
)

func (s *Service) connectorTypes(parent *gin.RouterGroup) {
	group := parent.Group("/connector_types")
	group.Use(NewPagination())

	group.GET("", s.getConnectorTypes)
	group.GET("/:name", s.getConnectorByName)
}

func (s *Service) getConnectorTypes(c *gin.Context) {

	page := c.GetInt(pagination.DEFAULT_PAGE_TEXT)
	size := c.GetInt(pagination.DEFAULT_SIZE_TEXT)

	count := len(s.catalog.Connectors)
	begin := (page - 1) * size
	end := min(begin+size, len(s.catalog.Connectors))

	if begin >= count {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	connectors := s.catalog.Connectors[begin:end]

	c.IndentedJSON(http.StatusOK, gin.H{
		"page":  page,
		"size":  end - begin,
		"total": len(s.catalog.Connectors),
		"items": connectors,
	})
}

func (s *Service) getConnectorByName(c *gin.Context) {

	name := c.Param("name")

	if connector, found := s.catalog.ByName[name]; found {
		c.IndentedJSON(http.StatusOK, connector)
		return
	}

	c.AbortWithStatus(http.StatusNotFound)
}
