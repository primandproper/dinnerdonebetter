package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	platformcomments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// BuildFakeComment builds a faked Comment.
//
// The target type is fixed rather than random because it is not a free string:
// the store refuses one the target catalog does not hold, so a randomized value
// would build a comment that could never be written.
func BuildFakeComment() *platformcomments.Comment {
	comment := fake.BuildFakeRecord[platformcomments.Comment]()
	comment.Scope = comments.Scope()
	comment.Target.Type = mealplanning.CommentTargetTypeRecipes
	comment.ParentID = platformcomments.RootParentID

	return comment
}

// BuildFakeCommentReply builds a faked Comment that replies to parentID.
//
// It shares the parent's target, because a reply belongs to its parent's
// discussion and one naming a different target is refused.
func BuildFakeCommentReply(parent *platformcomments.Comment) *platformcomments.Comment {
	reply := BuildFakeComment()
	reply.ParentID = parent.ID
	reply.Target = parent.Target

	return reply
}

// BuildFakeCommentList builds a faked page of Comments about one target.
//
// Every element carries the target, because that is what the read path filtered
// on: a page of comments about one recipe is the only page the read path returns.
func BuildFakeCommentList(target platformcomments.Target) *filtering.QueryFilteredResult[platformcomments.Comment] {
	return fake.BuildFakePage(func() *platformcomments.Comment {
		comment := BuildFakeComment()
		comment.Target = target

		return comment
	})
}
