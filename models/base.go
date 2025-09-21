package models

import (
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Base struct {
	// gorm:"<-:create" only allows create and read but not update
	// this is avoid updating created_at with a zero value by mistake
	CreatedAt time.Time      `json:"createdAt" gorm:"<-:create"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

type BaseWithId struct {
	Base
	ID string `gorm:"primaryKey" json:"id"`
}

type ErrorResponse struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	TraceId string      `json:"traceId"`
	Details interface{} `json:"details,omitempty"`
}

type PaginatedResult[T any] struct {
	Meta struct {
		TotalCount  int `json:"totalCount"`
		TotalPages  int `json:"totalPages"`
		CurrentPage int `json:"currentPage"`
		NextPage    int `json:"nextPage"`
	} `json:"meta"`
	Data []*T `json:"data"`
}

func BuildPaginatedResult[T any](data []*T, totalCount int64, page int, limit int) PaginatedResult[T] {
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	nextPage := 0
	if page < totalPages-1 {
		nextPage = page + 1
	}

	result := PaginatedResult[T]{
		Data: data,
	}
	result.Meta.TotalCount = int(totalCount)
	result.Meta.TotalPages = totalPages
	result.Meta.CurrentPage = page
	result.Meta.NextPage = nextPage

	return result
}

func ParsePaginationParams(c *gin.Context) (page int, perPage int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil {
		page = 0
	}
	perPage, err = strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if err != nil {
		perPage = 10
	}

	// Validate pagination parameters
	if page < 0 {
		page = 0
	}
	if perPage < 1 || perPage > 100 {
		perPage = 100
	}

	return page, perPage
}

func ParseSortParams(c *gin.Context, validSortFields map[string]bool, defaultSortBy string) (sortBy string, sort string) {
	sortBy = c.DefaultQuery("sort_by", defaultSortBy)
	sort = c.DefaultQuery("sort", "desc")

	// Validate sort parameters
	validSortOrder := map[string]bool{
		"asc":  true,
		"desc": true,
	}

	if !validSortFields[sortBy] || !validSortOrder[sort] {
		sortBy = defaultSortBy
		sort = "desc"
	}

	return sortBy, sort
}
