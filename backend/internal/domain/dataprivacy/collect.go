package dataprivacy

import (
	"context"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/filtering"
)

// CollectAcrossAccounts runs fetch once per account and concatenates the results.
//
// Account-scoped data is the awkward half of a subject access request: the
// subject is a person, but webhooks, settings, issue reports, and payments hang
// off the accounts that person appears in. Every collector that has to make that
// hop makes it the same way, so it is made here.
//
// The paging within one account is platform-go's — this is the only part of the
// walk that is this application's, because "a subject's data is spread across
// the accounts they are a member of" is a fact about this schema and not about
// subject access requests in general.
func CollectAcrossAccounts[T any](
	ctx context.Context,
	accountIDs []string,
	fetch func(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[T], error),
) ([]T, error) {
	var out []T

	for _, accountID := range accountIDs {
		values, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[T], error) {
			return fetch(ctx, accountID, filter)
		})
		if err != nil {
			return nil, err
		}

		out = append(out, values...)
	}

	return out, nil
}
