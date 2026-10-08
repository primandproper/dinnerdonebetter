package fakes

import (
	platformcomments "github.com/primandproper/platform-go/v15/comments"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// BuildFakeComment builds a faked Comment about something of targetType.
//
// The target type is the caller's rather than random because it is not a free string: the
// store refuses one the target catalog does not hold, so a randomized value would build a
// comment that could never be written. Which types the catalog holds is each domain's to say —
// see internal/build/comments — and this package, being generic, names none of them.
func BuildFakeComment(targetType platformcomments.TargetType) *platformcomments.Comment {
	comment := fake.BuildFakeRecord[platformcomments.Comment]()
	comment.Scope = tenancy.Global()
	comment.Target.Type = targetType
	comment.ParentID = platformcomments.RootParentID

	return comment
}

// BuildFakeCommentReply builds a faked Comment that replies to parent.
//
// It shares the parent's target, because a reply belongs to its parent's
// discussion and one naming a different target is refused.
func BuildFakeCommentReply(parent *platformcomments.Comment) *platformcomments.Comment {
	reply := BuildFakeComment(parent.Target.Type)
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
		comment := BuildFakeComment(target.Type)
		comment.Target = target

		return comment
	})
}
