package util

import "fmt"

func WithPagination(query string, page, pageSize int) string {
	offset := (page - 1) * pageSize
	return fmt.Sprintf("%s LIMIT %d OFFSET %d", query, pageSize, offset)
}
