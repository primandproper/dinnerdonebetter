package converters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/primandproper/platform-go/v3/identifiers"
)

// ConvertMimeTypeToGRPCMimeType converts domain MIME type string to protobuf enum.
func ConvertMimeTypeToGRPCMimeType(mimeType string) uploadedmediasvc.UploadedMediaMimeType {
	switch mimeType {
	case uploadedmedia.MimeTypeImagePNG:
		return uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_PNG
	case uploadedmedia.MimeTypeImageJPEG:
		return uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_JPEG
	case uploadedmedia.MimeTypeImageGIF:
		return uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_GIF
	case uploadedmedia.MimeTypeVideoMP4:
		return uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_VIDEO_MP4
	default:
		return uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_UNSPECIFIED
	}
}

// ConvertGRPCMimeTypeToMimeType converts protobuf enum to domain MIME type string.
func ConvertGRPCMimeTypeToMimeType(mimeType uploadedmediasvc.UploadedMediaMimeType) string {
	switch mimeType {
	case uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_PNG:
		return uploadedmedia.MimeTypeImagePNG
	case uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_JPEG:
		return uploadedmedia.MimeTypeImageJPEG
	case uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_GIF:
		return uploadedmedia.MimeTypeImageGIF
	case uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_VIDEO_MP4:
		return uploadedmedia.MimeTypeVideoMP4
	default:
		return ""
	}
}

func ConvertUploadedMediaToGRPCUploadedMedia(uploadedMedia *uploadedmedia.UploadedMedia) *uploadedmediasvc.UploadedMedia {
	return &uploadedmediasvc.UploadedMedia{
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(uploadedMedia.CreatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(uploadedMedia.ArchivedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(uploadedMedia.LastUpdatedAt),
		Id:            uploadedMedia.ID,
		StoragePath:   uploadedMedia.StoragePath,
		MimeType:      ConvertMimeTypeToGRPCMimeType(uploadedMedia.MimeType),
		CreatedByUser: uploadedMedia.CreatedByUser,
	}
}

func ConvertGRPCUploadedMediaToUploadedMedia(uploadedMedia *uploadedmediasvc.UploadedMedia) *uploadedmedia.UploadedMedia {
	return &uploadedmedia.UploadedMedia{
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(uploadedMedia.CreatedAt),
		ArchivedAt:    grpcconverters.ConvertPBTimestampToTimePointer(uploadedMedia.ArchivedAt),
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(uploadedMedia.LastUpdatedAt),
		ID:            uploadedMedia.Id,
		StoragePath:   uploadedMedia.StoragePath,
		MimeType:      ConvertGRPCMimeTypeToMimeType(uploadedMedia.MimeType),
		CreatedByUser: uploadedMedia.CreatedByUser,
	}
}

func ConvertGRPCUploadedMediaCreationRequestInputToUploadedMediaDatabaseCreationInput(input *uploadedmediasvc.UploadedMediaCreationRequestInput, userID string) *uploadedmedia.UploadedMediaDatabaseCreationInput {
	return &uploadedmedia.UploadedMediaDatabaseCreationInput{
		ID:            identifiers.New(),
		StoragePath:   input.StoragePath,
		MimeType:      ConvertGRPCMimeTypeToMimeType(input.MimeType),
		CreatedByUser: userID,
	}
}

func ConvertUploadedMediaCreationRequestInputToGRPCUploadedMediaCreationRequestInput(input *uploadedmedia.UploadedMediaCreationRequestInput) *uploadedmediasvc.UploadedMediaCreationRequestInput {
	return &uploadedmediasvc.UploadedMediaCreationRequestInput{
		StoragePath: input.StoragePath,
		MimeType:    ConvertMimeTypeToGRPCMimeType(input.MimeType),
	}
}

func ConvertGRPCUploadedMediaUpdateRequestInputToUploadedMediaUpdateRequestInput(input *uploadedmediasvc.UploadedMediaUpdateRequestInput) *uploadedmedia.UploadedMediaUpdateRequestInput {
	output := &uploadedmedia.UploadedMediaUpdateRequestInput{}

	if input.StoragePath != nil {
		output.StoragePath = input.StoragePath
	}

	if input.MimeType != nil {
		mimeType := ConvertGRPCMimeTypeToMimeType(*input.MimeType)
		output.MimeType = &mimeType
	}

	return output
}
