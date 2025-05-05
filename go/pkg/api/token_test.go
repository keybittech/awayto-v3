package api

import (
	"reflect"
	"strings"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/types"
)

var (
	// Sample bearer token string
	bearerToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	// Expected result
	expectedToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
)

// Method 1: Using strings.Contains to check prefix, then strings.Split
func extractTokenWithContainsAndSplit(token string) string {
	if strings.Contains(token, "Bearer") {
		parts := strings.Split(token, " ")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return token
}

// Method 2: Using strings.IndexOf to find space, then substring
func extractTokenWithIndexOf(token string) string {
	spaceIdx := strings.Index(token, " ")
	if spaceIdx != -1 && spaceIdx+1 < len(token) {
		return token[spaceIdx+1:]
	}
	return token
}

// Method 3: Using strings.TrimPrefix (corrected from TrimLeft which doesn't do what we need)
func extractTokenWithTrimPrefix(token string) string {
	return strings.TrimPrefix(token, "Bearer ")
}

// Method 4: Check if first char is 'B', then just use strings.TrimPrefix
func extractTokenWithFirstCharCheck(token string) string {
	if len(token) > 0 && token[0] == 'B' {
		return strings.TrimPrefix(token, "Bearer ")
	}
	return token
}

// Method 5: Direct reassignment with TrimPrefix without any checks
func directReassignment(token string) string {
	return strings.TrimPrefix(token, "Bearer ")
}

// Benchmark tests
func BenchmarkTokenContainsAndSplit(b *testing.B) {
	var result string
	for i := 0; i < b.N; i++ {
		result = extractTokenWithContainsAndSplit(bearerToken)
	}
	if result != expectedToken {
		b.Fatalf("Expected %s, got %s", expectedToken, result)
	}
}

func BenchmarkTokenIndexOf(b *testing.B) {
	var result string
	for i := 0; i < b.N; i++ {
		result = extractTokenWithIndexOf(bearerToken)
	}
	if result != expectedToken {
		b.Fatalf("Expected %s, got %s", expectedToken, result)
	}
}

func BenchmarkTokenTrimPrefix(b *testing.B) {
	var result string
	for i := 0; i < b.N; i++ {
		result = extractTokenWithTrimPrefix(bearerToken)
	}
	if result != expectedToken {
		b.Fatalf("Expected %s, got %s", expectedToken, result)
	}
}

func BenchmarkTokenFirstCharCheck(b *testing.B) {
	var result string
	for i := 0; i < b.N; i++ {
		result = extractTokenWithFirstCharCheck(bearerToken)
	}
	if result != expectedToken {
		b.Fatalf("Expected %s, got %s", expectedToken, result)
	}
}

func BenchmarkTokenDirectReassignment(b *testing.B) {
	for i := 0; i < b.N; i++ {
		directReassignment(bearerToken)
	}
}

// Additional tests with different token lengths
var (
	shortToken = "Bearer short.token"
	longToken  = "Bearer " + strings.Repeat("x", 10000) + ".very.long.token"
)

func BenchmarkTokenContainsAndSplit_ShortToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractTokenWithContainsAndSplit(shortToken)
	}
}

func BenchmarkTokenIndexOf_ShortToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractTokenWithIndexOf(shortToken)
	}
}

func BenchmarkTokenTrimPrefix_ShortToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractTokenWithTrimPrefix(shortToken)
	}
}

func BenchmarkTokenContainsAndSplit_LongToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractTokenWithContainsAndSplit(longToken)
	}
}

func BenchmarkTokenIndexOf_LongToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractTokenWithIndexOf(longToken)
	}
}

func BenchmarkTokenTrimPrefix_LongToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractTokenWithTrimPrefix(longToken)
	}
}

// Edge cases
func BenchmarkTokenMissingBearer(b *testing.B) {
	tokenWithoutBearer := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.token"

	b.Run("ContainsAndSplit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			extractTokenWithContainsAndSplit(tokenWithoutBearer)
		}
	})

	b.Run("IndexOf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			extractTokenWithIndexOf(tokenWithoutBearer)
		}
	})

	b.Run("TrimPrefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			extractTokenWithTrimPrefix(tokenWithoutBearer)
		}
	})
}

func TestNewTokenCache(t *testing.T) {
	tests := []struct {
		name string
		want *TokenCache
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTokenCache(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTokenCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenCache_Set(t *testing.T) {
	type args struct {
		token   string
		session *types.UserSession
	}
	tests := []struct {
		name string
		tc   *TokenCache
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tc.Set(tt.args.token, tt.args.session)
		})
	}
}

func TestTokenCache_Get(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name  string
		tc    *TokenCache
		args  args
		want  *types.UserSession
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.tc.Get(tt.args.token)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TokenCache.Get(%v) got = %v, want %v", tt.args.token, got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TokenCache.Get(%v) got1 = %v, want %v", tt.args.token, got1, tt.want1)
			}
		})
	}
}
