package grpcconverters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	platformconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
)

// ConvertProtoCommentCreationRequestInputToDomain converts proto CommentCreationRequestInput to domain.
// When targetType and referencedID are non-empty (from AddCommentTo* path), they are used.
// When empty (for CreateComment), target_type and referenced_id from in are used.
func ConvertProtoCommentCreationRequestInputToDomain(in *commentssvc.CommentCreationRequestInput, targetType, referencedID, belongsToUser string) *comments.CommentCreationRequestInput {
	if in == nil {
		return nil
	}
	if targetType == "" {
		targetType = in.TargetType
	}
	if referencedID == "" {
		referencedID = in.ReferencedId
	}
	return &comments.CommentCreationRequestInput{
		Content:         in.Content,
		TargetType:      targetType,
		ReferencedID:    referencedID,
		ParentCommentID: in.ParentCommentId,
		BelongsToUser:   belongsToUser,
	}
}

// ConvertProtoCommentUpdateRequestInputToDomain converts proto CommentUpdateRequestInput to domain.
func ConvertProtoCommentUpdateRequestInputToDomain(in *commentssvc.CommentUpdateRequestInput) *comments.CommentUpdateRequestInput {
	if in == nil {
		return nil
	}
	return &comments.CommentUpdateRequestInput{
		Content: in.Content,
	}
}

// ConvertCommentToGRPCComment converts a domain Comment to proto Comment.
func ConvertCommentToGRPCComment(input *comments.Comment) *commentssvc.Comment {
	if input == nil {
		return nil
	}
	return &commentssvc.Comment{
		Id:              input.ID,
		Content:         input.Content,
		TargetType:      input.TargetType,
		ReferencedId:    input.ReferencedID,
		ParentCommentId: input.ParentCommentID,
		BelongsToUser:   input.BelongsToUser,
		CreatedAt:       platformconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt:   platformconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
	}
}
