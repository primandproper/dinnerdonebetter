package testutils

import (
	"context"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

const (
	Example32ByteKey = "HEREISA32CHARSECRETWHICHISMADEUP"
	Example64ByteKey = "HEREISA64CHARSECRETWHICHISMADEUPHEREISA64CHARSECRETWHICHISMADEUP"
)

// ContextMatcher is a testify/mock argument matcher that matches any context.Context value.
var ContextMatcher = mock.MatchedBy(func(_ context.Context) bool { return true })

// QueryFilterMatcher is a testify/mock argument matcher that matches any *filtering.QueryFilter value.
var QueryFilterMatcher = mock.MatchedBy(func(_ *filtering.QueryFilter) bool { return true })

// MatchType returns a testify/mock argument matcher that matches any value of type T.
func MatchType[T any]() any {
	return mock.MatchedBy(func(_ T) bool { return true })
}
