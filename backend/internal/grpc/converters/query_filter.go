package grpcconverters

import (
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"

	"github.com/primandproper/platform-go/v7/filtering"
)

func ConvertGRPCQueryFilterToQueryFilter(qf *grpcfiltering.QueryFilter) *filtering.QueryFilter {
	if qf == nil {
		return filtering.DefaultQueryFilter()
	}

	filter := &filtering.QueryFilter{
		MaxResponseSize: new(uint8(50)),
		IncludeArchived: qf.IncludeArchived,
		Cursor:          qf.Cursor,
		SortBy:          qf.SortBy,
		CreatedAfter:    ConvertPBTimestampToTimePointer(qf.CreatedAfter),
		CreatedBefore:   ConvertPBTimestampToTimePointer(qf.CreatedBefore),
		UpdatedAfter:    ConvertPBTimestampToTimePointer(qf.UpdatedAfter),
		UpdatedBefore:   ConvertPBTimestampToTimePointer(qf.UpdatedBefore),
	}
	if qf.MaxResponseSize != nil {
		size := min(*qf.MaxResponseSize, filtering.MaxQueryFilterLimit)
		filter.MaxResponseSize = new(uint8(size))
	}

	return filter
}

func ConvertQueryFilterToGRPCQueryFilter(qf *filtering.QueryFilter, resultPagination filtering.Pagination) *grpcfiltering.QueryFilter {
	if qf == nil {
		qf = filtering.DefaultQueryFilter()
	}

	// For AppliedQueryFilter, we should use the filter's cursor (the cursor that was used to get to this page),
	// NOT the result pagination's cursor (which is the next page cursor).
	// The resultPagination parameter is only used for other purposes, not for setting AppliedQueryFilter.Cursor.
	cursor := qf.Cursor

	// Safeguard: If the filter's cursor matches the result pagination's cursor (both pointing to next page),
	// it means we're on page 1 (no cursor was used to get here), so set cursor to nil.
	if cursor != nil && resultPagination.Cursor != "" && *cursor == resultPagination.Cursor {
		cursor = nil
	}

	f := &grpcfiltering.QueryFilter{
		IncludeArchived: qf.IncludeArchived,
		Cursor:          cursor,
		SortBy:          qf.SortBy,
		CreatedAfter:    ConvertTimePointerToPBTimestamp(qf.CreatedAfter),
		CreatedBefore:   ConvertTimePointerToPBTimestamp(qf.CreatedBefore),
		UpdatedAfter:    ConvertTimePointerToPBTimestamp(qf.UpdatedAfter),
		UpdatedBefore:   ConvertTimePointerToPBTimestamp(qf.UpdatedBefore),
	}
	if qf.MaxResponseSize != nil {
		f.MaxResponseSize = new(uint32(*qf.MaxResponseSize))
	}

	return f
}

func ConvertPaginationToGRPCPagination(pagination filtering.Pagination, qf *filtering.QueryFilter) *grpcfiltering.Pagination {
	// Use AppliedQueryFilter from pagination if it exists (it's what was actually applied),
	// otherwise fall back to the qf parameter
	appliedFilter := pagination.AppliedQueryFilter
	if appliedFilter == nil {
		if qf == nil {
			appliedFilter = filtering.DefaultQueryFilter()
		} else {
			appliedFilter = qf
		}
	}

	f := &grpcfiltering.Pagination{
		AppliedQueryFilter: ConvertQueryFilterToGRPCQueryFilter(appliedFilter, pagination),
		Cursor:             pagination.Cursor,
		FilteredCount:      pagination.FilteredCount,
		TotalCount:         pagination.TotalCount,
		MaxResponseSize:    uint32(pagination.MaxResponseSize),
		PreviousCursor:     pagination.PreviousCursor,
	}

	return f
}
