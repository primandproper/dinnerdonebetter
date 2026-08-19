package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	uploadedmediagrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mealImageUploadChunkSize = 32 * 1024

func uploadMealImageForTest(t *testing.T, mealID, filename, contentType string, fileData []byte) string {
	t.Helper()
	ctx := t.Context()

	stream, err := adminClient.UploadMealImage(ctx)
	require.NoError(t, err)

	// First message: metadata
	err = stream.Send(&mealplanningsvc.UploadMealMediaRequest{
		MealId: mealID,
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
	for offset := 0; offset < len(fileData); offset += mealImageUploadChunkSize {
		end := min(offset+mealImageUploadChunkSize, len(fileData))
		chunk := fileData[offset:end]
		err = stream.Send(&mealplanningsvc.UploadMealMediaRequest{
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

func TestUploadMealImage(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createdMeal := createMealForTest(t, adminClient, nil)
		fileData := []byte("fake image data for integration test")
		filename := "test-image.jpg"
		contentType := uploadedmedia.MimeTypeImageJPEG

		uploadedMediaID := uploadMealImageForTest(t, createdMeal.ID, filename, contentType, fileData)
		assert.NotEmpty(t, uploadedMediaID)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdMeal := createMealForTest(t, adminClient, nil)
		c := buildUnauthenticatedGRPCClientForTest(t)

		stream, err := c.UploadMealImage(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadMealMediaRequest{
			MealId: createdMeal.ID,
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

	T.Run("nonexistent meal", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		stream, err := adminClient.UploadMealImage(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadMealMediaRequest{
			MealId: nonexistentID,
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
