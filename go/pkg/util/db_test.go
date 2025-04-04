package util

import "testing"

func TestUtilWithPagination(t *testing.T) {
	type args struct {
		query    string
		page     int
		pageSize int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Empty query", args: args{"", 1, 10}, want: " LIMIT 10 OFFSET 0"},
		{name: "Regular query", args: args{"SELECT * FROM products WHERE category = 'test'", 3, 15}, want: "SELECT * FROM products WHERE category = 'test' LIMIT 15 OFFSET 30"},
		{name: "Negative page size", args: args{"SELECT * FROM users", 1, -5}, want: "SELECT * FROM users LIMIT -5 OFFSET 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithPagination(tt.args.query, tt.args.page, tt.args.pageSize); got != tt.want {
				t.Errorf("WithPagination() = %v, want %v", got, tt.want)
			}
		})
	}
}
