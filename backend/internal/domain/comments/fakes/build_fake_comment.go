package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// commentTargetType is the kind of thing these fakes comment on.
//
// A comment's target is a table name, and the ones a comment may point at are a closed
// set the domain knows and a random string is not in.
const commentTargetType = "recipes"

// BuildFakeComment builds a faked Comment.
func BuildFakeComment() *comments.Comment {
	comment := fake.BuildFakeRecord[comments.Comment]()
	comment.TargetType = commentTargetType

	return comment
}

// BuildFakeCommentWithParent builds a faked Comment that is a reply.
func BuildFakeCommentWithParent(parentID string) *comments.Comment {
	c := BuildFakeComment()
	c.ParentCommentID = &parentID

	return c
}

// BuildFakeCommentList builds a faked Comment list.
//
// The target is what the read path filtered on, so every element of the page carries
// it: a page of comments about one recipe, which is the only page the read path
// returns.
func BuildFakeCommentList(targetType, referencedID string) *filtering.QueryFilteredResult[comments.Comment] {
	return fake.BuildFakePage(func() *comments.Comment {
		comment := BuildFakeComment()
		comment.TargetType = targetType
		comment.ReferencedID = referencedID

		return comment
	})
}

// BuildFakeCommentCreationRequestInput builds a faked CommentCreationRequestInput.
func BuildFakeCommentCreationRequestInput() *comments.CommentCreationRequestInput {
	input := fake.BuildFakeRecord[comments.CommentCreationRequestInput]()
	input.TargetType = commentTargetType

	return input
}

// BuildFakeCommentDatabaseCreationInput builds a faked CommentDatabaseCreationInput.
func BuildFakeCommentDatabaseCreationInput() *comments.CommentDatabaseCreationInput {
	input := fake.BuildFakeRecord[comments.CommentDatabaseCreationInput]()
	input.TargetType = commentTargetType

	return input
}
