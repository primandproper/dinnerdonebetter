package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
)

// UploadedMediaMethodPermissions is a named type for Wire dependency injection.
type UploadedMediaMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the uploaded media service's method permissions.
func ProvideMethodPermissions() UploadedMediaMethodPermissions {
	return UploadedMediaMethodPermissions{
		uploadedmediasvc.UploadedMediaService_Upload_FullMethodName: {
			authorization.CreateUploadedMediaPermission,
		},
		uploadedmediasvc.UploadedMediaService_CreateUploadedMedia_FullMethodName: {
			authorization.CreateUploadedMediaPermission,
		},
		uploadedmediasvc.UploadedMediaService_GetUploadedMedia_FullMethodName: {
			authorization.ReadUploadedMediaPermission,
		},
		uploadedmediasvc.UploadedMediaService_GetUploadedMediaWithIDs_FullMethodName: {
			authorization.ReadUploadedMediaPermission,
		},
		uploadedmediasvc.UploadedMediaService_GetUploadedMediaForUser_FullMethodName: {
			authorization.ReadUploadedMediaPermission,
		},
		uploadedmediasvc.UploadedMediaService_ArchiveUploadedMedia_FullMethodName: {
			authorization.ArchiveUploadedMediaPermission,
		},
	}
}
