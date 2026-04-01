package pagination

import (
	"math"
	"net/http"
	"strconv"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Params holds pagination parameters extracted from a request.
type Params struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// Offset returns the database offset for the current page.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Response wraps a paginated list result.
type Response[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

// NewResponse creates a paginated response from items and total count.
func NewResponse[T any](items []T, total int, params Params) Response[T] {
	if items == nil {
		items = []T{}
	}
	totalPages := 0
	if params.PerPage > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(params.PerPage)))
	}
	return Response[T]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}
}

// ParseFromQuery extracts and validates page/per_page query parameters.
func ParseFromQuery(r *http.Request) Params {
	page := parseIntOr(r.URL.Query().Get("page"), DefaultPage)
	perPage := parseIntOr(r.URL.Query().Get("per_page"), DefaultPerPage)

	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return Params{Page: page, PerPage: perPage}
}

func parseIntOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
