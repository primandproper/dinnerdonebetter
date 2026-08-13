package dataprivacy

import (
	"context"
	"encoding/json"

	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
)

// CollectAllValues is CollectAllPages with the pointers dereferenced.
//
// Every collector wants values rather than pointers — the fragment it encodes is
// a document, and a slice of pointers marshals identically while inviting a nil
// element nobody checks for. Doing it here rather than in each collector removes
// the same four-line copy loop from a dozen call sites.
func CollectAllValues[T any](
	ctx context.Context,
	fetch func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[T], error),
) ([]T, error) {
	pointers, err := CollectAllPages(ctx, fetch)
	if err != nil {
		return nil, err
	}

	if len(pointers) == 0 {
		return nil, nil
	}

	values := make([]T, 0, len(pointers))
	for _, pointer := range pointers {
		if pointer == nil {
			continue
		}

		values = append(values, *pointer)
	}

	return values, nil
}

// CollectAcrossAccounts runs fetch once per account and concatenates the results.
//
// Account-scoped data is the awkward half of a subject access request: the
// subject is a person, but webhooks, settings, issue reports, and payments hang
// off the accounts that person appears in. Every collector that has to make that
// hop makes it the same way, so it is made here.
func CollectAcrossAccounts[T any](
	ctx context.Context,
	accountIDs []string,
	fetch func(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[T], error),
) ([]T, error) {
	var out []T

	for _, accountID := range accountIDs {
		values, err := CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[T], error) {
			return fetch(ctx, accountID, filter)
		})
		if err != nil {
			return nil, err
		}

		out = append(out, values...)
	}

	return out, nil
}

// Fragment encodes a collector's result, or reports that the domain holds
// nothing about the subject.
//
// Returning nil, nil is how platform-go's Collector says "no data here", and the
// section is then omitted from the artifact rather than written as null. That
// distinction is the reason this helper exists: an artifact whose sections are
// the domains that actually held something reads as an answer, while one padded
// with empty objects for every domain in the application reads as a form.
func Fragment(held bool, v any) (json.RawMessage, error) {
	if !held {
		return nil, nil
	}

	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, platformerrors.Wrap(err, "encoding data privacy fragment")
	}

	return encoded, nil
}
