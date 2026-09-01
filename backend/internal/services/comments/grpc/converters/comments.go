package grpcconverters

import (
	platformconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"

	comments "github.com/primandproper/platform-go/v13/comments"
)

// ConvertProtoCommentTargetToDomain converts a proto CommentTarget to the
// platform's. A nil target is the zero target, which is what a reply that adopts
// its parent's target looks like on the way in.
func ConvertProtoCommentTargetToDomain(in *commentssvc.CommentTarget) comments.Target {
	if in == nil {
		return comments.Target{}
	}

	return comments.Target{
		Type: comments.TargetType(in.GetType()),
		ID:   in.GetId(),
	}
}

// ConvertCommentTargetToGRPC converts the platform's target to proto.
func ConvertCommentTargetToGRPC(in comments.Target) *commentssvc.CommentTarget {
	return &commentssvc.CommentTarget{
		Type: in.Type.String(),
		Id:   in.ID,
	}
}

// ConvertProtoCommentCreationRequestInputToDomain converts a proto
// CommentCreationRequestInput to the comment the store writes.
//
// The target is a parameter as well as a field on the input because two paths
// reach here: CreateComment, where the client names the target, and the
// AddCommentTo* methods, where the URL does. A non-zero target argument is the
// second case and wins, which is what keeps AddCommentToRecipe from filing a
// comment against whatever the body happened to say.
func ConvertProtoCommentCreationRequestInputToDomain(in *commentssvc.CommentCreationRequestInput, target comments.Target, author string) *comments.Comment {
	if in == nil {
		return nil
	}

	if target.Zero() {
		target = ConvertProtoCommentTargetToDomain(in.GetTarget())
	}

	return &comments.Comment{
		Body:     in.GetBody(),
		ParentID: in.GetParentId(),
		Target:   target,
		Author:   author,
	}
}

// ConvertCommentToGRPCComment converts a stored comment to proto.
func ConvertCommentToGRPCComment(input *comments.Comment) *commentssvc.Comment {
	if input == nil {
		return nil
	}

	return &commentssvc.Comment{
		Id:            input.ID,
		Body:          input.Body,
		Target:        ConvertCommentTargetToGRPC(input.Target),
		ParentId:      input.ParentID,
		Author:        input.Author,
		CreatedAt:     platformconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt: platformconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
	}
}
