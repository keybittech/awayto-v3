package util

import (
	"strconv"
)

func WithPagination(query string, page, pageSize int) string {
	return query + " LIMIT " + strconv.Itoa(pageSize) + " OFFSET " + strconv.Itoa((page-1)*pageSize)
}
