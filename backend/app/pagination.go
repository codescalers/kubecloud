package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	defaultLimit  = 20
	maxLimit      = 100
	defaultOffset = 0
)

// PaginationQuery is a reusable binding model for limit/offset validation via Gin
type PaginationQuery struct {
	Limit  int `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int `form:"offset" binding:"omitempty,min=0"`
}

// bindPagination binds and validates pagination query params using Gin bindings
func bindPagination(c *gin.Context) (int, int, bool) {
	var q PaginationQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		Error(c, http.StatusBadRequest, "Invalid pagination parameters", err.Error())
		return 0, 0, false
	}

	// Apply defaults and caps
	limit := q.Limit
	if limit == 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := q.Offset
	if offset < 0 {
		offset = defaultOffset
	}

	return limit, offset, true
}
