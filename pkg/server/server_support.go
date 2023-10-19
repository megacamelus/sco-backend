package server

import (
	"github.com/gin-gonic/gin"
	pagination "github.com/webstradev/gin-pagination"
)

func NewPagination() gin.HandlerFunc {
	return pagination.New(
		pagination.DEFAULT_PAGE_TEXT,
		pagination.DEFAULT_SIZE_TEXT,
		pagination.DEFAULT_PAGE,
		pagination.DEFAULT_PAGE_SIZE,
		1,
		pagination.DEFAULT_MAX_PAGESIZE)
}
