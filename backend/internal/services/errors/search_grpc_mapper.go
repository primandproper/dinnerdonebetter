package errors

import (
	"errors"

	"github.com/primandproper/platform-go/v10/errors/grpc"
	textsearch "github.com/primandproper/platform-go/v10/search/text"
	"github.com/primandproper/platform-go/v10/search/text/elasticsearch"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(searchGRPCMapper{})
}

// searchGRPCMapper maps the text search index's refusals to page further. They are
// separated from the meal planning and identity mappers because both of those
// domains search, and neither refusal is about what was searched for.
type searchGRPCMapper struct{}

func (searchGRPCMapper) Map(err error) (code codes.Code, ok bool) {
	if err == nil {
		return codes.Unknown, false
	}

	switch {
	// The caller asked to page deeper than Elasticsearch will serve. OutOfRange is
	// gRPC's code for a well-formed request that ran past the end of a range, and it
	// says the thing a client can act on: paging further will not work, so narrow the
	// query. Reporting it as an empty last page would have read as "no more results",
	// which is a different fact.
	case errors.Is(err, elasticsearch.ErrResultWindowExceeded):
		return codes.OutOfRange, true
	// A cursor the index did not issue. Cursors are tagged with the backend that made
	// them, so this is usually one carried over from a database-backed page — the two
	// kinds of cursor travel in the same field — or one left over from a backend
	// swap. Either way the client sent something it should not have.
	case errors.Is(err, textsearch.ErrInvalidCursor):
		return codes.InvalidArgument, true
	default:
		return codes.Unknown, false
	}
}
