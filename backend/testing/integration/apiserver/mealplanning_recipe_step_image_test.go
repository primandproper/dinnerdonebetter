package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	uploadedmediagrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const recipeStepImageUploadChunkSize = 32 * 1024

func uploadRecipeStepImageForTest(t *testing.T, recipeID, recipeStepID, filename, contentType string, fileData []byte) string {
	t.Helper()
	ctx := t.Context()

	stream, err := adminClient.UploadRecipeStepImage(ctx)
	require.NoError(t, err)

	// First message: metadata
	err = stream.Send(&mealplanningsvc.UploadRecipeStepImageRequest{
		RecipeId:     recipeID,
		RecipeStepId: recipeStepID,
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
	for offset := 0; offset < len(fileData); offset += recipeStepImageUploadChunkSize {
		end := min(offset+recipeStepImageUploadChunkSize, len(fileData))
		chunk := fileData[offset:end]
		err = stream.Send(&mealplanningsvc.UploadRecipeStepImageRequest{
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

func TestUploadRecipeStepImage(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		require.NotEmpty(t, createdRecipe.Steps)
		firstStep := createdRecipe.Steps[0]

		fileData := []byte("fake step image for integration test")
		filename := "step-image.jpg"
		contentType := uploadedmedia.MimeTypeImageJPEG

		uploadedMediaID := uploadRecipeStepImageForTest(t, createdRecipe.ID, firstStep.ID, filename, contentType, fileData)
		assert.NotEmpty(t, uploadedMediaID)

		// Verify recipe step is enriched with step images when step is read via GetRecipeStep
		retrieved, err := adminClient.GetRecipeStep(ctx, &mealplanningsvc.GetRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: firstStep.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Len(t, retrieved.Result.StepImages, 1)
		assert.Equal(t, uploadedMediaID, retrieved.Result.StepImages[0].Id)
		assert.Equal(t, uploadedmediagrpc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_JPEG, retrieved.Result.StepImages[0].MimeType)

		// Verify GetRecipe also returns step images for each step
		fullRecipe, err := adminClient.GetRecipe(ctx, &mealplanningsvc.GetRecipeRequest{RecipeId: createdRecipe.ID})
		require.NoError(t, err)
		require.NotNil(t, fullRecipe)
		require.NotEmpty(t, fullRecipe.Result.Steps)
		require.Len(t, fullRecipe.Result.Steps[0].StepImages, 1)
		assert.Equal(t, uploadedMediaID, fullRecipe.Result.Steps[0].StepImages[0].Id)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		require.NotEmpty(t, createdRecipe.Steps)
		firstStep := createdRecipe.Steps[0]

		c := buildUnauthenticatedGRPCClientForTest(t)

		stream, err := c.UploadRecipeStepImage(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadRecipeStepImageRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: firstStep.ID,
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

	T.Run("nonexistent recipe step", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		stream, err := adminClient.UploadRecipeStepImage(ctx)
		require.NoError(t, err)

		err = stream.Send(&mealplanningsvc.UploadRecipeStepImageRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: nonexistentID,
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
