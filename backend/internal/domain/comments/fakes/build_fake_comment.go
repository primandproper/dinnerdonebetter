package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"

	"github.com/primandproper/platform-go/v10/filtering"
)

// BuildFakeCommentWithParent builds a faked Comment that is a reply.
func BuildFakeCommentWithParent(parentID string) *comments.Comment {
	c := BuildFakeComment()
	c.ParentCommentID = &parentID
	return c
}

// BuildFakeCommentList builds a faked Comment list.
func BuildFakeCommentList(targetType, referencedID string) *filtering.QueryFilteredResult[comments.Comment] {
	var examples []*comments.Comment
	for range 3 {
		examples = append(examples, BuildFakeComment())
	}

	return &filtering.QueryFilteredResult[comments.Comment]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   3,
			TotalCount:      3,
		},
		Data: examples,
	}
}
