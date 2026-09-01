package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
)

// CommentsMethodPermissions is a named type for Wire dependency injection.
type CommentsMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the comments service's method permissions.
func ProvideMethodPermissions() CommentsMethodPermissions {
	return CommentsMethodPermissions{
		commentssvc.CommentsService_CreateComment_FullMethodName:     {authorization.CreateCommentsPermission},
		commentssvc.CommentsService_GetRootComments_FullMethodName:   {authorization.ReadCommentsPermission},
		commentssvc.CommentsService_GetCommentReplies_FullMethodName: {authorization.ReadCommentsPermission},
		commentssvc.CommentsService_UpdateComment_FullMethodName:     {authorization.UpdateCommentsPermission},
		commentssvc.CommentsService_ArchiveComment_FullMethodName:    {authorization.ArchiveCommentsPermission},
	}
}
