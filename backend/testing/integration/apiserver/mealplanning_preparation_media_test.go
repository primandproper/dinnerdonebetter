package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	uploadedmediagrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const preparationMediaUploadChunkSize = 32 * 1024

func uploadPreparationMediaForTest(t *testing.T, validPreparationID, filename, contentType string, fileData []byte) string {
	t.Helper()
	ctx := t.Context()

	stream, err := adminClient.UploadPreparationMedia(ctx)
	require.NoError(t, err)

	// First message: metadata
	err = stream.Send(&mealplanningsvc.UploadPreparationMediaRequest{
		ValidPreparationId: validPreparationID,
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
	for offset := 0; offset < len(fileData); offset += preparationMediaUploadChunkSize {
		end := min(offset+preparationMediaUploadChunkSize, len(fileData))
		chunk := fileData[offset:end]
		err = stream.Send(&mealplanningsvc.UploadPreparationMediaRequest{
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

func TestUploadPreparationMedia(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		fileData := []byte("fake image data for integration test")
		filename := "test-image.jpg"
		contentType := uploadedmedia.MimeTypeImageJPEG

		uploadedMediaID := uploadPreparationMediaForTest(t, preparation.ID, filename, contentType, fileData)
		assert.NotEmpty(t, uploadedMediaID)

		// Verify preparation is enriched with media when read
		retrieved, err := adminClient.GetValidPreparation(ctx, &mealplanningsvc.GetValidPreparationRequest{
			ValidPreparationId: preparation.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Len(t, retrieved.Result.Media, 1)
		assert.Equal(t, uploadedMediaID, retrieved.Result.Media[0].Id)
		assert.Equal(t, uploadedmedia.MimeTypeImageJPEG, retrieved.Result.Media[0].ContentType)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		stream, err := c.UploadPreparationMedia(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadPreparationMediaRequest{
			ValidPreparationId: preparation.ID,
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

	T.Run("nonexistent preparation", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		stream, err := adminClient.UploadPreparationMedia(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadPreparationMediaRequest{
			ValidPreparationId: nonexistentID,
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
