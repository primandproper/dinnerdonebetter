package converters

import (
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/primandproper/platform-go/v13/uploads/registry"
)

// ConvertUploadedMediaToGRPCUploadedMedia renders a registry row on the wire.
func ConvertUploadedMediaToGRPCUploadedMedia(object *registry.Object) *uploadedmediasvc.UploadedMedia {
	if object == nil {
		return nil
	}

	return &uploadedmediasvc.UploadedMedia{
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(object.CreatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(object.ArchivedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(object.LastUpdatedAt),
		Id:            object.ID,
		ObjectKey:     object.Key,
		ContentType:   object.ContentType,
		OwnerId:       object.OwnerID,
		SizeBytes:     object.Size,
		BelongsToType: object.BelongsTo.Type,
		BelongsToId:   object.BelongsTo.ID,
	}
}

// ConvertGRPCUploadedMediaToUploadedMedia reads a registry row off the wire.
//
// The scope is deliberately absent: it is not on the wire and never should be —
// a client that could name a tenancy could name somebody else's. Whoever writes
// a row stamps it from uploadedmedia.Scope.
func ConvertGRPCUploadedMediaToUploadedMedia(object *uploadedmediasvc.UploadedMedia) *registry.Object {
	if object == nil {
		return nil
	}

	return &registry.Object{
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(object.CreatedAt),
		ArchivedAt:    grpcconverters.ConvertPBTimestampToTimePointer(object.ArchivedAt),
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(object.LastUpdatedAt),
		ID:            object.Id,
		Key:           object.ObjectKey,
		ContentType:   object.ContentType,
		OwnerID:       object.OwnerId,
		Size:          object.SizeBytes,
		BelongsTo:     registry.Subject{Type: object.BelongsToType, ID: object.BelongsToId},
	}
}
