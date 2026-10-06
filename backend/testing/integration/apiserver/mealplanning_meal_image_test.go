package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/primandproper/platform-go/v15/mediaregistry/mediaregistrypb"
	"github.com/primandproper/primitives-go/v2/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const mealImageUploadChunkSize = 32 * 1024

func uploadMealImageForTest(t *testing.T, mealID, filename, contentType string, fileData []byte) string {
	t.Helper()
	ctx := t.Context()

	stream, err := adminClient.UploadMealImage(ctx)
	require.NoError(t, err)

	// First message: the upload header
	err = stream.Send(&mealplanningsvc.UploadMealMediaRequest{
		MealId: mealID,
		Upload: &mediaregistrypb.UploadObjectRequest{
			Part: &mediaregistrypb.UploadObjectRequest_Header{
				Header: &mediaregistrypb.UploadObjectHeader{
					Name:        filename,
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
			Upload: &mediaregistrypb.UploadObjectRequest{
				Part: &mediaregistrypb.UploadObjectRequest_Chunk{Chunk: chunk},
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

	T.Run("a name that walks out of its prefix is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdMeal := createMealForTest(t, adminClient, nil)

		stream, err := adminClient.UploadMealImage(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadMealMediaRequest{
			MealId: createdMeal.ID,
			Upload: &mediaregistrypb.UploadObjectRequest{
				Part: &mediaregistrypb.UploadObjectRequest_Header{
					Header: &mediaregistrypb.UploadObjectHeader{
						Name:        "../../../" + identifiers.New() + "/" + identifiers.New() + ".jpg",
						ContentType: uploadedmedia.MimeTypeImageJPEG,
					},
				},
			},
		})
		requireStreamSend(t, err)

		_, err = stream.CloseAndRecv()
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
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
			Upload: &mediaregistrypb.UploadObjectRequest{
				Part: &mediaregistrypb.UploadObjectRequest_Header{
					Header: &mediaregistrypb.UploadObjectHeader{
						Name:        "test.jpg",
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
			Upload: &mediaregistrypb.UploadObjectRequest{
				Part: &mediaregistrypb.UploadObjectRequest_Header{
					Header: &mediaregistrypb.UploadObjectHeader{
						Name:        "test.jpg",
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
