package pagination

import (
	"net/http"
	"strconv"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
	MaxPage        = 1_000_000
)

// Params holds pagination parameters.
type Params struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// Offset returns the SQL offset value.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the SQL limit value.
func (p Params) Limit() int {
	return p.PerPage
}

// Result holds a paginated result set.
type Result[T any] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// NewResult creates a paginated result.
func NewResult[T any](data []T, params Params, totalCount int) Result[T] {
	totalPages := totalCount / params.PerPage
	if totalCount%params.PerPage != 0 {
		totalPages++
	}
	return Result[T]{
		Data:       data,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}

// FromRequest extracts pagination params from an HTTP request.
func FromRequest(r *http.Request) Params {
	page := queryInt(r, "page", DefaultPage)
	perPage := queryInt(r, "per_page", DefaultPerPage)

	if page < 1 {
		page = DefaultPage
	}
	if page > MaxPage {
		page = MaxPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return Params{Page: page, PerPage: perPage}
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}
