package utils

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	DefaultListPageSize = 10
	MaxListPageSize     = 50
)

// ListPagination stores the page number and size.
type ListPagination struct {
	Page  int
	Limit int
}

// ParseListPagination reads and checks pagination from the URL.
func ParseListPagination(query url.Values) (ListPagination, error) {
	// Use these defaults when no values are supplied.
	settings := ListPagination{
		Page:  1,
		Limit: DefaultListPageSize,
	}

	// Check that a supplied page is a positive whole number.
	if value := query.Get("page"); value != "" {
		page, err := strconv.Atoi(value)
		if err != nil || page < 1 {
			return ListPagination{}, fmt.Errorf("page must be a positive whole number")
		}
		settings.Page = page
	}

	// Check that a supplied page size is a positive whole number.
	if value := query.Get("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 {
			return ListPagination{}, fmt.Errorf("limit must be a positive whole number")
		}
		settings.Limit = limit
	}

	// Prevent requests for more than 50 results per page.
	if settings.Limit > MaxListPageSize {
		return ListPagination{}, fmt.Errorf(
			"limit must be between 1 and %d", MaxListPageSize,
		)
	}

	// Prevent very large page numbers from overflowing the calculation.
	maxInt := int(^uint(0) >> 1)
	if settings.Page-1 > maxInt/settings.Limit {
		return ListPagination{}, fmt.Errorf("page number is too large")
	}

	return settings, nil
}
