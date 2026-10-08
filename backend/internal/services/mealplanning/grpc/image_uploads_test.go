package grpc

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"path"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mockmanagers "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v15/mediaregistry"
	mediaregistrygrpc "github.com/primandproper/platform-go/v15/mediaregistry/grpc"
	"github.com/primandproper/platform-go/v15/mediaregistry/mediaregistrypb"
	registrymock "github.com/primandproper/platform-go/v15/mediaregistry/mock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/fake"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
	"github.com/primandproper/primitives-go/v2/uploads"
	mockuploads "github.com/primandproper/primitives-go/v2/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidObjectName(T *testing.T) {
	T.Parallel()

	T.Run("admits one path segment", func(t *testing.T) {
		t.Parallel()

		assert.True(t, validObjectName(fake.BuildFakeID()+".png"))
	})

	T.Run("refuses an empty name", func(t *testing.T) {
		t.Parallel()

		assert.False(t, validObjectName(""))
	})

	T.Run("refuses the names of a directory", func(t *testing.T) {
		t.Parallel()

		assert.False(t, validObjectName("."))
		assert.False(t, validObjectName(".."))
	})

	T.Run("refuses a name that walks out of its prefix", func(t *testing.T) {
		t.Parallel()

		assert.False(t, validObjectName("../../"+fake.BuildFakeID()+"/"+fake.BuildFakeID()+".png"))
	})

	T.Run("refuses a name carrying either separator", func(t *testing.T) {
		t.Parallel()

		assert.False(t, validObjectName(fake.BuildFakeID()+"/"+fake.BuildFakeID()+".png"))
		assert.False(t, validObjectName(fake.BuildFakeID()+`\`+fake.BuildFakeID()+".png"))
	})
}

func TestBelongsToAgrees(T *testing.T) {
	T.Parallel()

	subject := mediaregistry.Subject{Type: recipeSubjectType, ID: fake.BuildFakeID()}

	T.Run("admits an absent belongs_to", func(t *testing.T) {
		t.Parallel()

		assert.True(t, belongsToAgrees(mediaregistry.Subject{}, subject))
	})

	T.Run("admits the subject the call uploads to", func(t *testing.T) {
		t.Parallel()

		assert.True(t, belongsToAgrees(mediaregistry.Subject{Type: subject.Type, ID: subject.ID}, subject))
	})

	T.Run("refuses another subject of the same type", func(t *testing.T) {
		t.Parallel()

		assert.False(t, belongsToAgrees(mediaregistry.Subject{Type: subject.Type, ID: fake.BuildFakeID()}, subject))
	})

	T.Run("refuses the same id under another type", func(t *testing.T) {
		t.Parallel()

		assert.False(t, belongsToAgrees(mediaregistry.Subject{Type: mealSubjectType, ID: subject.ID}, subject))
	})

	T.Run("refuses a subject naming only half of one", func(t *testing.T) {
		t.Parallel()

		assert.False(t, belongsToAgrees(mediaregistry.Subject{ID: subject.ID}, subject))
		assert.False(t, belongsToAgrees(mediaregistry.Subject{Type: subject.Type}, subject))
	})
}

// fakeUploadStream is a client stream that delivers messages, then end (or err, when set), and
// keeps the response the server closed it with.
type fakeUploadStream[Req, Res any] struct {
	grpc.ServerStream

	ctx      context.Context
	err      error
	response *Res
	messages []*Req
	received int
}

func (f *fakeUploadStream[Req, Res]) Context() context.Context { return f.ctx }

func (f *fakeUploadStream[Req, Res]) Recv() (*Req, error) {
	if f.received < len(f.messages) {
		f.received++

		return f.messages[f.received-1], nil
	}

	if f.err != nil {
		return nil, f.err
	}

	return nil, io.EOF
}

func (f *fakeUploadStream[Req, Res]) SendAndClose(res *Res) error {
	f.response = res

	return nil
}

func uploadHeaderForTest(name, contentType string, belongsTo *mediaregistrypb.Subject) *mediaregistrypb.UploadObjectRequest {
	return &mediaregistrypb.UploadObjectRequest{
		Part: &mediaregistrypb.UploadObjectRequest_Header{
			Header: &mediaregistrypb.UploadObjectHeader{Name: name, ContentType: contentType, BelongsTo: belongsTo},
		},
	}
}

func uploadChunkForTest(chunk []byte) *mediaregistrypb.UploadObjectRequest {
	return &mediaregistrypb.UploadObjectRequest{Part: &mediaregistrypb.UploadObjectRequest_Chunk{Chunk: chunk}}
}

// uploadHarness is a service over a bucket that keeps what it is given, a registry that echoes
// what it records, and a manager that owns nothing until a test says so.
type uploadHarness struct {
	service  *serviceImpl
	manager  *mockmanagers.MealPlanningManagerMock
	registry *registrymock.StoreMock
	bucket   *mockuploads.UploadManagerMock
	saved    map[string][]byte
	ctx      context.Context
	userID   string
}

func buildUploadHarness(t *testing.T) *uploadHarness {
	t.Helper()

	ctx := buildSessionContextForTest(t)
	h := &uploadHarness{
		ctx:     ctx,
		userID:  sessionUserIDForTest(t, ctx),
		saved:   map[string][]byte{},
		manager: &mockmanagers.MealPlanningManagerMock{},
	}

	h.bucket = &mockuploads.UploadManagerMock{
		ExistsFunc: func(_ context.Context, key string) (bool, error) {
			_, ok := h.saved[key]

			return ok, nil
		},
		SaveFunc: func(_ context.Context, key string, r io.Reader, _ ...uploads.SaveOption) error {
			content, err := io.ReadAll(r)
			if err != nil {
				return err
			}

			h.saved[key] = content

			return nil
		},
		DeleteFunc: func(_ context.Context, key string) error {
			delete(h.saved, key)

			return nil
		},
	}

	h.registry = &registrymock.StoreMock{
		RecordObjectFunc: func(_ context.Context, _ database.Tx, scope tenancy.Scope, in mediaregistry.ObjectInput) (*mediaregistry.Object, error) {
			return &mediaregistry.Object{
				ID:          in.ID,
				Key:         in.Key,
				ContentType: in.ContentType,
				OwnerID:     in.OwnerID,
				BelongsTo:   in.BelongsTo,
				Size:        in.Size,
				Scope:       scope,
			}, nil
		},
	}

	h.service = &serviceImpl{
		tracer:              tracing.NewTracerForTest(t.Name()),
		logger:              loggingnoop.NewLogger(),
		db:                  testutils.MockDatabaseClient(),
		mealPlanningManager: h.manager,
		registry:            h.registry,
		uploadManager:       h.bucket,
	}

	return h
}

// ownRecipe makes the harness's caller the author of a fresh recipe, and returns it.
func (h *uploadHarness) ownRecipe() *mealplanning.Recipe {
	recipe := mealplanningfakes.BuildFakeRecipe()
	recipe.CreatedByUser = h.userID

	h.manager.ReadRecipeFunc = func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
		if recipeID != recipe.ID {
			return nil, sql.ErrNoRows
		}

		return recipe, nil
	}
	h.manager.AddRecipeImageFunc = func(context.Context, string, string, string) error { return nil }

	return recipe
}

// uploadRecipeImage runs UploadRecipeImage over messages, each wrapped for the recipe.
func (h *uploadHarness) uploadRecipeImage(recipeID string, streamErr error, messages ...*mediaregistrypb.UploadObjectRequest) (*fakeUploadStream[mealplanningsvc.UploadRecipeMediaRequest, mealplanningsvc.UploadRecipeImageResponse], error) {
	stream := &fakeUploadStream[mealplanningsvc.UploadRecipeMediaRequest, mealplanningsvc.UploadRecipeImageResponse]{ctx: h.ctx, err: streamErr}
	for _, msg := range messages {
		stream.messages = append(stream.messages, &mealplanningsvc.UploadRecipeMediaRequest{RecipeId: recipeID, Upload: msg})
	}

	return stream, h.service.UploadRecipeImage(stream)
}

func imageNameForTest() string {
	return fake.BuildFakeID() + ".png"
}

func TestServiceImpl_UploadRecipeImage(T *testing.T) {
	T.Parallel()

	T.Run("stores, registers and attaches the upload", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()
		name := imageNameForTest()
		first, second := []byte(fake.BuildFakeID()), []byte(fake.BuildFakeID())

		stream, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(name, uploadedmedia.MimeTypeImagePNG, nil),
			uploadChunkForTest(first),
			uploadChunkForTest(second),
		)
		require.NoError(t, err)

		require.Len(t, h.registry.RecordObjectCalls(), 1)
		recorded := h.registry.RecordObjectCalls()[0].In
		assert.Equal(t, mediaregistry.Subject{Type: recipeSubjectType, ID: recipe.ID}, recorded.BelongsTo)
		assert.Equal(t, h.userID, recorded.OwnerID)
		assert.Equal(t, uploadedmedia.MimeTypeImagePNG, recorded.ContentType)
		assert.Equal(t, path.Join("recipes", recipe.ID, recorded.ID, name), recorded.Key)
		assert.Equal(t, int64(len(first)+len(second)), recorded.Size)
		assert.Equal(t, append(append([]byte{}, first...), second...), h.saved[recorded.Key])

		require.Len(t, h.manager.AddRecipeImageCalls(), 1)
		attached := h.manager.AddRecipeImageCalls()[0]
		assert.Equal(t, recipe.ID, attached.RecipeID)
		assert.Equal(t, recorded.ID, attached.UploadedMediaID)
		assert.Equal(t, h.userID, attached.UploadedByUser)

		require.NotNil(t, stream.response)
		assert.Equal(t, recorded.ID, stream.response.GetUploadedMediaId())
	})

	T.Run("admits a belongs_to naming the recipe it uploads to", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, &mediaregistrypb.Subject{Type: recipeSubjectType, Id: recipe.ID}),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
		)
		require.NoError(t, err)
		assert.Len(t, h.manager.AddRecipeImageCalls(), 1)
	})

	T.Run("refuses a first message that is not a header", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, nil, uploadChunkForTest([]byte(fake.BuildFakeID())))
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, h.saved)
	})

	T.Run("refuses a stream that sent nothing", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		h.ownRecipe()

		_, err := h.uploadRecipeImage("", nil)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, h.saved)
	})

	T.Run("refuses a name that walks out of its prefix before reading the recipe", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest("../../"+imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
		)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, h.manager.ReadRecipeCalls())
		assert.Empty(t, h.saved)
	})

	T.Run("refuses a content type this application does not store", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), "text/html", nil),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
		)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, h.saved)
	})

	T.Run("refuses an upload that states no content type", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), "", nil),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
		)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, h.saved)
	})

	T.Run("refuses a belongs_to naming another recipe", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, &mediaregistrypb.Subject{Type: recipeSubjectType, Id: fake.BuildFakeID()}),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
		)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, h.saved)
	})

	T.Run("refuses a recipe the caller did not write", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()
		recipe.CreatedByUser = fake.BuildFakeID()

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
		)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		assert.Empty(t, h.saved)
		assert.Empty(t, h.manager.AddRecipeImageCalls())
	})

	T.Run("refuses a header with no bytes after it", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil),
			uploadChunkForTest(nil),
		)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, h.bucket.SaveCalls())
		assert.Empty(t, h.registry.RecordObjectCalls())
	})

	T.Run("refuses an upload past the cap without registering it", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil),
			uploadChunkForTest(make([]byte, maxImageUploadSize)),
			uploadChunkForTest([]byte{1}),
		)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		require.ErrorContains(t, err, mediaregistrygrpc.ErrObjectTooLarge.Error())
		assert.Empty(t, h.registry.RecordObjectCalls())
		assert.Empty(t, h.manager.AddRecipeImageCalls())
	})

	T.Run("refuses a second header", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil),
		)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, h.registry.RecordObjectCalls())
	})

	T.Run("answers a stream that broke with the transport's code", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()

		_, err := h.uploadRecipeImage(recipe.ID, status.Error(codes.DeadlineExceeded, fake.BuildFakeID()),
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
		)
		assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
		assert.Empty(t, h.registry.RecordObjectCalls())
	})

	T.Run("removes the bytes when the registration fails", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()
		h.registry.RecordObjectFunc = func(context.Context, database.Tx, tenancy.Scope, mediaregistry.ObjectInput) (*mediaregistry.Object, error) {
			return nil, errors.New(fake.BuildFakeID())
		}

		_, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
		)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Len(t, h.bucket.SaveCalls(), 1)
		assert.Len(t, h.bucket.DeleteCalls(), 1)
		assert.Empty(t, h.saved)
		assert.Empty(t, h.manager.AddRecipeImageCalls())
	})

	T.Run("fails when the bridge row cannot be written", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()
		h.manager.AddRecipeImageFunc = func(context.Context, string, string, string) error { return errors.New(fake.BuildFakeID()) }

		stream, err := h.uploadRecipeImage(recipe.ID, nil,
			uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil),
			uploadChunkForTest([]byte(fake.BuildFakeID())),
		)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Nil(t, stream.response)
	})
}

func TestServiceImpl_UploadMealImage(T *testing.T) {
	T.Parallel()

	T.Run("files the upload under the meal and attaches it", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		meal := mealplanningfakes.BuildFakeMeal()
		meal.CreatedByUser = h.userID
		h.manager.ReadMealFunc = func(context.Context, string) (*mealplanning.Meal, error) { return meal, nil }
		h.manager.AddMealImageFunc = func(context.Context, string, string, string) error { return nil }
		name := imageNameForTest()

		stream := &fakeUploadStream[mealplanningsvc.UploadMealMediaRequest, mealplanningsvc.UploadMealImageResponse]{
			ctx: h.ctx,
			messages: []*mealplanningsvc.UploadMealMediaRequest{
				{MealId: meal.ID, Upload: uploadHeaderForTest(name, uploadedmedia.MimeTypeImageJPEG, nil)},
				{Upload: uploadChunkForTest([]byte(fake.BuildFakeID()))},
			},
		}
		require.NoError(t, h.service.UploadMealImage(stream))

		recorded := h.registry.RecordObjectCalls()[0].In
		assert.Equal(t, mediaregistry.Subject{Type: mealSubjectType, ID: meal.ID}, recorded.BelongsTo)
		assert.Equal(t, path.Join("meals", meal.ID, recorded.ID, name), recorded.Key)

		require.Len(t, h.manager.AddMealImageCalls(), 1)
		assert.Equal(t, meal.ID, h.manager.AddMealImageCalls()[0].MealID)
		assert.Equal(t, recorded.ID, h.manager.AddMealImageCalls()[0].UploadedMediaID)
		assert.Equal(t, h.userID, h.manager.AddMealImageCalls()[0].UploadedByUser)
		assert.Equal(t, recorded.ID, stream.response.GetUploadedMediaId())
	})

	T.Run("refuses a meal the caller did not write", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		meal := mealplanningfakes.BuildFakeMeal()
		h.manager.ReadMealFunc = func(context.Context, string) (*mealplanning.Meal, error) { return meal, nil }

		stream := &fakeUploadStream[mealplanningsvc.UploadMealMediaRequest, mealplanningsvc.UploadMealImageResponse]{
			ctx: h.ctx,
			messages: []*mealplanningsvc.UploadMealMediaRequest{
				{MealId: meal.ID, Upload: uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImageJPEG, nil)},
				{Upload: uploadChunkForTest([]byte(fake.BuildFakeID()))},
			},
		}
		assert.Equal(t, codes.PermissionDenied, status.Code(h.service.UploadMealImage(stream)))
		assert.Empty(t, h.saved)
	})
}

func TestServiceImpl_UploadRecipeStepImage(T *testing.T) {
	T.Parallel()

	T.Run("files the upload under the step and attaches it", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()
		step := mealplanningfakes.BuildFakeRecipeStep()
		h.manager.ReadRecipeStepFunc = func(context.Context, string, string) (*mealplanning.RecipeStep, error) { return step, nil }
		h.manager.AddRecipeStepImageFunc = func(context.Context, string, string, string) error { return nil }
		name := imageNameForTest()

		stream := &fakeUploadStream[mealplanningsvc.UploadRecipeStepImageRequest, mealplanningsvc.UploadRecipeStepImageResponse]{
			ctx: h.ctx,
			messages: []*mealplanningsvc.UploadRecipeStepImageRequest{
				{RecipeId: recipe.ID, RecipeStepId: step.ID, Upload: uploadHeaderForTest(name, uploadedmedia.MimeTypeImagePNG, nil)},
				{Upload: uploadChunkForTest([]byte(fake.BuildFakeID()))},
			},
		}
		require.NoError(t, h.service.UploadRecipeStepImage(stream))

		recorded := h.registry.RecordObjectCalls()[0].In
		assert.Equal(t, mediaregistry.Subject{Type: recipeStepSubjectType, ID: step.ID}, recorded.BelongsTo)
		assert.Equal(t, path.Join("recipes", recipe.ID, "steps", step.ID, recorded.ID, name), recorded.Key)

		require.Len(t, h.manager.ReadRecipeStepCalls(), 1)
		assert.Equal(t, recipe.ID, h.manager.ReadRecipeStepCalls()[0].RecipeID)

		require.Len(t, h.manager.AddRecipeStepImageCalls(), 1)
		assert.Equal(t, step.ID, h.manager.AddRecipeStepImageCalls()[0].RecipeStepID)
		assert.Equal(t, recorded.ID, h.manager.AddRecipeStepImageCalls()[0].UploadedMediaID)
		assert.Equal(t, recorded.ID, stream.response.GetUploadedMediaId())
	})

	T.Run("refuses a step that is not the recipe's", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		recipe := h.ownRecipe()
		h.manager.ReadRecipeStepFunc = func(context.Context, string, string) (*mealplanning.RecipeStep, error) { return nil, sql.ErrNoRows }

		stream := &fakeUploadStream[mealplanningsvc.UploadRecipeStepImageRequest, mealplanningsvc.UploadRecipeStepImageResponse]{
			ctx: h.ctx,
			messages: []*mealplanningsvc.UploadRecipeStepImageRequest{
				{RecipeId: recipe.ID, RecipeStepId: fake.BuildFakeID(), Upload: uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil)},
				{Upload: uploadChunkForTest([]byte(fake.BuildFakeID()))},
			},
		}
		assert.Equal(t, codes.NotFound, status.Code(h.service.UploadRecipeStepImage(stream)))
		assert.Empty(t, h.saved)
	})
}

func TestServiceImpl_UploadPreparationMedia(T *testing.T) {
	T.Parallel()

	T.Run("files the upload under the preparation and attaches it for the ingredient", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		preparation := mealplanningfakes.BuildFakeValidPreparation()
		ingredientID := fake.BuildFakeID()
		h.manager.ReadValidPreparationFunc = func(context.Context, string) (*mealplanning.ValidPreparation, error) { return preparation, nil }
		h.manager.AddPreparationMediaFunc = func(context.Context, string, *string, string, int32) error { return nil }
		name := imageNameForTest()

		stream := &fakeUploadStream[mealplanningsvc.UploadPreparationMediaRequest, mealplanningsvc.UploadPreparationMediaResponse]{
			ctx: h.ctx,
			messages: []*mealplanningsvc.UploadPreparationMediaRequest{
				{ValidPreparationId: preparation.ID, ForIngredientId: &ingredientID, Upload: uploadHeaderForTest(name, uploadedmedia.MimeTypeVideoMP4, nil)},
				{Upload: uploadChunkForTest([]byte(fake.BuildFakeID()))},
			},
		}
		require.NoError(t, h.service.UploadPreparationMedia(stream))

		recorded := h.registry.RecordObjectCalls()[0].In
		assert.Equal(t, mediaregistry.Subject{Type: validPreparationSubjectType, ID: preparation.ID}, recorded.BelongsTo)
		assert.Equal(t, path.Join("preparations", preparation.ID, recorded.ID, name), recorded.Key)

		require.Len(t, h.manager.AddPreparationMediaCalls(), 1)
		attached := h.manager.AddPreparationMediaCalls()[0]
		assert.Equal(t, preparation.ID, attached.ValidPreparationID)
		require.NotNil(t, attached.ForIngredientID)
		assert.Equal(t, ingredientID, *attached.ForIngredientID)
		assert.Equal(t, recorded.ID, attached.UploadedMediaID)
		assert.Equal(t, recorded.ID, stream.response.GetUploadedMediaId())
	})

	T.Run("refuses a preparation that does not exist", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		h.manager.ReadValidPreparationFunc = func(context.Context, string) (*mealplanning.ValidPreparation, error) { return nil, sql.ErrNoRows }

		stream := &fakeUploadStream[mealplanningsvc.UploadPreparationMediaRequest, mealplanningsvc.UploadPreparationMediaResponse]{
			ctx: h.ctx,
			messages: []*mealplanningsvc.UploadPreparationMediaRequest{
				{ValidPreparationId: fake.BuildFakeID(), Upload: uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil)},
				{Upload: uploadChunkForTest([]byte(fake.BuildFakeID()))},
			},
		}
		assert.Equal(t, codes.NotFound, status.Code(h.service.UploadPreparationMedia(stream)))
		assert.Empty(t, h.saved)
	})
}

func TestServiceImpl_UploadIngredientMedia(T *testing.T) {
	T.Parallel()

	T.Run("files the upload under the ingredient and attaches it", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		ingredient := mealplanningfakes.BuildFakeValidIngredient()
		h.manager.ReadValidIngredientFunc = func(context.Context, string) (*mealplanning.ValidIngredient, error) { return ingredient, nil }
		h.manager.AddIngredientMediaFunc = func(context.Context, string, string, int32) error { return nil }
		name := imageNameForTest()

		stream := &fakeUploadStream[mealplanningsvc.UploadIngredientMediaRequest, mealplanningsvc.UploadIngredientMediaResponse]{
			ctx: h.ctx,
			messages: []*mealplanningsvc.UploadIngredientMediaRequest{
				{ValidIngredientId: ingredient.ID, Upload: uploadHeaderForTest(name, uploadedmedia.MimeTypeImageGIF, nil)},
				{Upload: uploadChunkForTest([]byte(fake.BuildFakeID()))},
			},
		}
		require.NoError(t, h.service.UploadIngredientMedia(stream))

		recorded := h.registry.RecordObjectCalls()[0].In
		assert.Equal(t, mediaregistry.Subject{Type: validIngredientSubjectType, ID: ingredient.ID}, recorded.BelongsTo)
		assert.Equal(t, path.Join("ingredients", ingredient.ID, recorded.ID, name), recorded.Key)

		require.Len(t, h.manager.AddIngredientMediaCalls(), 1)
		assert.Equal(t, ingredient.ID, h.manager.AddIngredientMediaCalls()[0].ValidIngredientID)
		assert.Equal(t, recorded.ID, h.manager.AddIngredientMediaCalls()[0].UploadedMediaID)
		assert.Equal(t, recorded.ID, stream.response.GetUploadedMediaId())
	})

	T.Run("refuses an ingredient that does not exist", func(t *testing.T) {
		t.Parallel()

		h := buildUploadHarness(t)
		h.manager.ReadValidIngredientFunc = func(context.Context, string) (*mealplanning.ValidIngredient, error) { return nil, sql.ErrNoRows }

		stream := &fakeUploadStream[mealplanningsvc.UploadIngredientMediaRequest, mealplanningsvc.UploadIngredientMediaResponse]{
			ctx: h.ctx,
			messages: []*mealplanningsvc.UploadIngredientMediaRequest{
				{ValidIngredientId: fake.BuildFakeID(), Upload: uploadHeaderForTest(imageNameForTest(), uploadedmedia.MimeTypeImagePNG, nil)},
				{Upload: uploadChunkForTest([]byte(fake.BuildFakeID()))},
			},
		}
		assert.Equal(t, codes.NotFound, status.Code(h.service.UploadIngredientMedia(stream)))
		assert.Empty(t, h.saved)
	})
}
