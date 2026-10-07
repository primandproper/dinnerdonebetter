package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/primandproper/platform-go/v15/mediaregistry/mediaregistrypb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const recipeImageUploadChunkSize = 32 * 1024

func uploadRecipeImageForTest(t *testing.T, recipeID, filename, contentType string, fileData []byte) string {
	t.Helper()
	ctx := t.Context()

	stream, err := adminClient.UploadRecipeImage(ctx)
	require.NoError(t, err)

	// First message: the upload header
	err = stream.Send(&mealplanningsvc.UploadRecipeMediaRequest{
		RecipeId: recipeID,
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
	for offset := 0; offset < len(fileData); offset += recipeImageUploadChunkSize {
		end := min(offset+recipeImageUploadChunkSize, len(fileData))
		chunk := fileData[offset:end]
		err = stream.Send(&mealplanningsvc.UploadRecipeMediaRequest{
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

func TestUploadRecipeImage(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		fileData := []byte("fake image data for integration test")
		filename := "test-image.jpg"
		contentType := uploadedmedia.MimeTypeImageJPEG

		uploadedMediaID := uploadRecipeImageForTest(t, createdRecipe.ID, filename, contentType, fileData)
		assert.NotEmpty(t, uploadedMediaID)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		c := buildUnauthenticatedGRPCClientForTest(t)

		stream, err := c.UploadRecipeImage(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadRecipeMediaRequest{
			RecipeId: createdRecipe.ID,
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

	T.Run("nonexistent recipe", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		stream, err := adminClient.UploadRecipeImage(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadRecipeMediaRequest{
			RecipeId: nonexistentID,
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
