package dataprivacy

import (
	"context"

	"github.com/primandproper/platform-go/v13/filtering"
)

// CollectAllPages exhaustively pages through a cursor-paginated list accessor and returns
// every row. Data disclosures must be complete, so a single default-sized page is never
// enough — this loops until a short page indicates the final one.
func CollectAllPages[T any](ctx context.Context, fetch func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[T], error)) ([]*T, error) {
	var out []*T

	filter := filtering.DefaultQueryFilter()
	filter.MaxResponseSize = new(filtering.MaxQueryFilterLimit)

	for {
		page, err := fetch(ctx, filter)
		if err != nil {
			return nil, err
		}
		out = append(out, page.Data...)

		if uint64(len(page.Data)) < uint64(*filter.MaxResponseSize) || page.Cursor == "" {
			break
		}
		if filter.Cursor != nil && page.Cursor == *filter.Cursor {
			// the cursor did not advance; bail rather than loop forever
			break
		}

		cursor := page.Cursor
		filter.Cursor = &cursor
	}

	return out, nil
}
