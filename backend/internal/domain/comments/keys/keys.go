package keys

const (
	idSuffix = ".id"

	// CommentIDKey is the standard key for referring to a comment ID.
	CommentIDKey = "comment" + idSuffix

	// CommentTargetTypeKey is the standard key for the kind of thing a comment is
	// about.
	CommentTargetTypeKey = "comment.target_type"

	// CommentTargetIDKey is the standard key for which one of them.
	CommentTargetIDKey = "comment.target" + idSuffix
)
