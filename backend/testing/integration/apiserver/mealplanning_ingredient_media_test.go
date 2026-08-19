package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	uploadedmediagrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ingredientMediaUploadChunkSize = 32 * 1024

func uploadIngredientMediaForTest(t *testing.T, validIngredientID, filename, contentType string, fileData []byte) string {
	t.Helper()
	ctx := t.Context()

	stream, err := adminClient.UploadIngredientMedia(ctx)
	require.NoError(t, err)

	// First message: metadata
	err = stream.Send(&mealplanningsvc.UploadIngredientMediaRequest{
		ValidIngredientId: validIngredientID,
		Upload: &uploadedmediagrpc.UploadRequest{
			Payload: &uploadedmediagrpc.UploadRequest_Metadata{
				Metadata: &uploadedmediagrpc.UploadMetadata{
					ObjectName:  filename,
					ContentType: contentType,
				},
			},
		},
	})
	require.NoError(t, err)

	// Stream chunks
	for offset := 0; offset < len(fileData); offset += ingredientMediaUploadChunkSize {
		end := min(offset+ingredientMediaUploadChunkSize, len(fileData))
		chunk := fileData[offset:end]
		err = stream.Send(&mealplanningsvc.UploadIngredientMediaRequest{
			Upload: &uploadedmediagrpc.UploadRequest{
				Payload: &uploadedmediagrpc.UploadRequest_Chunk{Chunk: chunk},
			},
		})
		require.NoError(t, err)
	}

	resp, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.UploadedMediaId)
	return *resp.UploadedMediaId
}

func TestUploadIngredientMedia(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		ingredient := createValidIngredientForTest(t)
		fileData := []byte("fake image data for integration test")
		filename := "test-image.jpg"
		contentType := uploadedmedia.MimeTypeImageJPEG

		uploadedMediaID := uploadIngredientMediaForTest(t, ingredient.ID, filename, contentType, fileData)
		assert.NotEmpty(t, uploadedMediaID)

		// Verify ingredient is enriched with media when read
		retrieved, err := adminClient.GetValidIngredient(ctx, &mealplanningsvc.GetValidIngredientRequest{
			ValidIngredientId: ingredient.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Len(t, retrieved.Result.Media, 1)
		assert.Equal(t, uploadedMediaID, retrieved.Result.Media[0].Id)
		assert.Equal(t, uploadedmediagrpc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_JPEG, retrieved.Result.Media[0].MimeType)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		ingredient := createValidIngredientForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		stream, err := c.UploadIngredientMedia(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadIngredientMediaRequest{
			ValidIngredientId: ingredient.ID,
			Upload: &uploadedmediagrpc.UploadRequest{
				Payload: &uploadedmediagrpc.UploadRequest_Metadata{
					Metadata: &uploadedmediagrpc.UploadMetadata{
						ObjectName:  "test.jpg",
						ContentType: uploadedmedia.MimeTypeImageJPEG,
					},
				},
			},
		})
		requireStreamSend(t, err)

		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})

	T.Run("nonexistent ingredient", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		stream, err := adminClient.UploadIngredientMedia(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadIngredientMediaRequest{
			ValidIngredientId: nonexistentID,
			Upload: &uploadedmediagrpc.UploadRequest{
				Payload: &uploadedmediagrpc.UploadRequest_Metadata{
					Metadata: &uploadedmediagrpc.UploadMetadata{
						ObjectName:  "test.jpg",
						ContentType: uploadedmedia.MimeTypeImageJPEG,
					},
				},
			},
		})
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})
}
