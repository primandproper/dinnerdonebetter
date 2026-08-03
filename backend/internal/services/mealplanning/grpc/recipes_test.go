package grpc

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mockmanagers "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/primandproper/platform-go/v9/fake"
	"github.com/primandproper/platform-go/v9/filtering"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildServiceImplForRecipesTest(t *testing.T) *serviceImpl {
	t.Helper()

	return &serviceImpl{
		tracer:          tracing.NewTracerForTest(t.Name()),
		logger:          loggingnoop.NewLogger(),
		commentsManager: &noopCommentsManager{},
	}
}

func TestServiceImpl_verifyRecipeOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		userID, err := s.verifyRecipeOwnership(ctx, exampleRecipeID, s.logger, span)
		require.NoError(t, err)
		assert.Equal(t, exampleUserID, userID)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		_, err := s.verifyRecipeOwnership(ctx, exampleRecipeID, s.logger, span)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID: exampleUserID,
			},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			ArchiveRecipeFunc: func(_ context.Context, recipeID string, ownerID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleUserID, ownerID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: exampleRecipeID})
		assert.NotNil(t, res)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.ArchiveRecipeCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: exampleRecipeID})
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipePrepTaskID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			ArchiveRecipePrepTaskFunc: func(_ context.Context, recipeID string, recipePrepTaskID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipePrepTaskID, recipePrepTaskID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipePrepTask(ctx, &mealplanninggrpc.ArchiveRecipePrepTaskRequest{
			RecipeId:         exampleRecipeID,
			RecipePrepTaskId: exampleRecipePrepTaskID,
		})
		assert.NotNil(t, res)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.ArchiveRecipePrepTaskCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipePrepTaskID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipePrepTask(ctx, &mealplanninggrpc.ArchiveRecipePrepTaskRequest{
			RecipeId:         exampleRecipeID,
			RecipePrepTaskId: exampleRecipePrepTaskID,
		})
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeRatingID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRating := &mealplanning.RecipeRating{ID: exampleRecipeRatingID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeRatingFunc: func(_ context.Context, recipeID string, recipeRatingID string) (*mealplanning.RecipeRating, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeRatingID, recipeRatingID)

				return exampleRating, nil
			},
			ArchiveRecipeRatingFunc: func(_ context.Context, recipeID string, recipeRatingID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeRatingID, recipeRatingID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       exampleRecipeID,
			RecipeRatingId: exampleRecipeRatingID,
		})
		assert.NotNil(t, res)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeRatingCalls(), 1)
		assert.Len(t, mrm.ArchiveRecipeRatingCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeLists(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		list := &mealplanning.RecipeList{ID: mealplanningfakes.BuildFakeID()}
		expected := &filtering.QueryFilteredResult[mealplanning.RecipeList]{Data: []*mealplanning.RecipeList{list}}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipeListsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeList], error) {
				return expected, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.GetRecipeLists(ctx, &mealplanninggrpc.GetRecipeListsRequest{})
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Results, 1)

		assert.Len(t, mrm.ListRecipeListsCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		userID := mealplanningfakes.BuildFakeID()
		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: userID},
		})

		input := &mealplanninggrpc.RecipeListCreationRequestInput{Name: t.Name(), Description: "desc"}
		created := &mealplanning.RecipeList{ID: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			CreateRecipeListFunc: func(_ context.Context, actualUserID string, _ *mealplanning.RecipeListCreationRequestInput) (*mealplanning.RecipeList, error) {
				assert.Equal(t, userID, actualUserID)

				return created, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.CreateRecipeList(ctx, &mealplanninggrpc.CreateRecipeListRequest{Input: input})
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, created.ID, res.Created.Id)

		assert.Len(t, mrm.CreateRecipeListCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		userID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()
		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: userID},
		})

		name := t.Name()
		desc := "desc"
		input := &mealplanninggrpc.RecipeListUpdateRequestInput{
			Name:        &name,
			Description: &desc,
		}

		mrm := &mockmanagers.MealPlanningManagerMock{
			UpdateRecipeListFunc: func(_ context.Context, recipeListID string, actualUserID string, _ *mealplanning.RecipeListUpdateRequestInput) error {
				assert.Equal(t, listID, recipeListID)
				assert.Equal(t, userID, actualUserID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeList(ctx, &mealplanninggrpc.UpdateRecipeListRequest{
			RecipeListId: listID,
			Input:        input,
		})
		require.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mrm.UpdateRecipeListCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		userID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()
		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: userID},
		})

		mrm := &mockmanagers.MealPlanningManagerMock{
			ArchiveRecipeListFunc: func(_ context.Context, recipeListID string, actualUserID string) error {
				assert.Equal(t, listID, recipeListID)
				assert.Equal(t, userID, actualUserID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeList(ctx, &mealplanninggrpc.ArchiveRecipeListRequest{RecipeListId: listID})
		require.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mrm.ArchiveRecipeListCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeListItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		listID := mealplanningfakes.BuildFakeID()
		item := &mealplanning.RecipeListItem{ID: mealplanningfakes.BuildFakeID(), Recipe: mealplanning.Recipe{ID: mealplanningfakes.BuildFakeID()}}
		expected := &filtering.QueryFilteredResult[mealplanning.RecipeListItem]{Data: []*mealplanning.RecipeListItem{item}}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipeListItemsFunc: func(_ context.Context, recipeListID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeListItem], error) {
				assert.Equal(t, listID, recipeListID)

				return expected, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.GetRecipeListItems(ctx, &mealplanninggrpc.GetRecipeListItemsRequest{RecipeListId: listID})
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Results, 1)

		assert.Len(t, mrm.ListRecipeListItemsCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipeListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		listID := mealplanningfakes.BuildFakeID()
		recipeID := mealplanningfakes.BuildFakeID()
		input := &mealplanninggrpc.RecipeListItemCreationRequestInput{
			BelongsToRecipeList: listID,
			RecipeId:            recipeID,
			Notes:               t.Name(),
		}

		created := &mealplanning.RecipeListItem{ID: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			AddRecipeToRecipeListFunc: func(_ context.Context, recipeListID string, actualRecipeID string, notes string) (*mealplanning.RecipeListItem, error) {
				assert.Equal(t, listID, recipeListID)
				assert.Equal(t, recipeID, actualRecipeID)
				assert.Equal(t, input.Notes, notes)

				return created, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.CreateRecipeListItem(ctx, &mealplanninggrpc.CreateRecipeListItemRequest{Input: input})
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, created.ID, res.Created.Id)

		assert.Len(t, mrm.AddRecipeToRecipeListCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipeListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		itemID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()
		recipeID := mealplanningfakes.BuildFakeID()
		notes := new(t.Name())
		input := &mealplanninggrpc.RecipeListItemUpdateRequestInput{
			BelongsToRecipeList: &listID,
			RecipeId:            &recipeID,
			Notes:               notes,
		}

		mrm := &mockmanagers.MealPlanningManagerMock{
			UpdateRecipeListItemFunc: func(_ context.Context, recipeListItemID string, recipeListID string, actualRecipeID string, _ *mealplanning.RecipeListItemUpdateRequestInput) error {
				assert.Equal(t, itemID, recipeListItemID)
				assert.Equal(t, listID, recipeListID)
				assert.Equal(t, recipeID, actualRecipeID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeListItem(ctx, &mealplanninggrpc.UpdateRecipeListItemRequest{
			RecipeListItemId: itemID,
			Input:            input,
		})
		require.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mrm.UpdateRecipeListItemCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipeListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		itemID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()

		mrm := &mockmanagers.MealPlanningManagerMock{
			RemoveRecipeFromRecipeListFunc: func(_ context.Context, recipeListID string, recipeListItemID string) error {
				assert.Equal(t, listID, recipeListID)
				assert.Equal(t, itemID, recipeListItemID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeListItem(ctx, &mealplanninggrpc.ArchiveRecipeListItemRequest{
			RecipeListItemId: itemID,
			RecipeListId:     listID,
		})
		require.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mrm.RemoveRecipeFromRecipeListCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			ArchiveRecipeStepFunc: func(_ context.Context, recipeID string, recipeStepID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		assert.NotNil(t, res)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.ArchiveRecipeStepCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepCompletionConditionID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			ArchiveRecipeStepCompletionConditionFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepCompletionConditionID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepCompletionConditionID, recipeStepCompletionConditionID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepCompletionCondition(ctx, &mealplanninggrpc.ArchiveRecipeStepCompletionConditionRequest{
			RecipeId:                        exampleRecipeID,
			RecipeStepId:                    exampleRecipeStepID,
			RecipeStepCompletionConditionId: exampleRecipeStepCompletionConditionID,
		})
		assert.NotNil(t, res)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.ArchiveRecipeStepCompletionConditionCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepCompletionConditionID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepCompletionCondition(ctx, &mealplanninggrpc.ArchiveRecipeStepCompletionConditionRequest{
			RecipeId:                        exampleRecipeID,
			RecipeStepId:                    exampleRecipeStepID,
			RecipeStepCompletionConditionId: exampleRecipeStepCompletionConditionID,
		})
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepIngredientID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			ArchiveRecipeStepIngredientFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepIngredientID, recipeStepIngredientID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepIngredient(ctx, &mealplanninggrpc.ArchiveRecipeStepIngredientRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleRecipeStepID,
			RecipeStepIngredientId: exampleRecipeStepIngredientID,
		})
		assert.NotNil(t, res)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.ArchiveRecipeStepIngredientCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepIngredientID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepIngredient(ctx, &mealplanninggrpc.ArchiveRecipeStepIngredientRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleRecipeStepID,
			RecipeStepIngredientId: exampleRecipeStepIngredientID,
		})
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepInstrumentID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			ArchiveRecipeStepInstrumentFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepInstrumentID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepInstrumentID, recipeStepInstrumentID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepInstrument(ctx, &mealplanninggrpc.ArchiveRecipeStepInstrumentRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleRecipeStepID,
			RecipeStepInstrumentId: exampleRecipeStepInstrumentID,
		})
		assert.NotNil(t, res)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.ArchiveRecipeStepInstrumentCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepInstrumentID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepInstrument(ctx, &mealplanninggrpc.ArchiveRecipeStepInstrumentRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleRecipeStepID,
			RecipeStepInstrumentId: exampleRecipeStepInstrumentID,
		})
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepProductID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			ArchiveRecipeStepProductFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepProductID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepProductID, recipeStepProductID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepProduct(ctx, &mealplanninggrpc.ArchiveRecipeStepProductRequest{
			RecipeId:            exampleRecipeID,
			RecipeStepId:        exampleRecipeStepID,
			RecipeStepProductId: exampleRecipeStepProductID,
		})
		assert.NotNil(t, res)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.ArchiveRecipeStepProductCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepProductID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepProduct(ctx, &mealplanninggrpc.ArchiveRecipeStepProductRequest{
			RecipeId:            exampleRecipeID,
			RecipeStepId:        exampleRecipeStepID,
			RecipeStepProductId: exampleRecipeStepProductID,
		})
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_ArchiveRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepVesselID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			ArchiveRecipeStepVesselFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepVesselID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepVesselID, recipeStepVesselID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepVessel(ctx, &mealplanninggrpc.ArchiveRecipeStepVesselRequest{
			RecipeId:           exampleRecipeID,
			RecipeStepId:       exampleRecipeStepID,
			RecipeStepVesselId: exampleRecipeStepVesselID,
		})
		assert.NotNil(t, res)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.ArchiveRecipeStepVesselCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepVesselID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepVessel(ctx, &mealplanninggrpc.ArchiveRecipeStepVesselRequest{
			RecipeId:           exampleRecipeID,
			RecipeStepId:       exampleRecipeStepID,
			RecipeStepVesselId: exampleRecipeStepVesselID,
		})
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_CloneRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleClonedRecipe := mealplanningfakes.BuildFakeRecipe()

		mrm := &mockmanagers.MealPlanningManagerMock{
			CloneRecipeFunc: func(_ context.Context, recipeID string, newOwnerID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleUserID, newOwnerID)

				return exampleClonedRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID: exampleUserID,
			},
		})

		res, err := s.CloneRecipe(ctx, &mealplanninggrpc.CloneRecipeRequest{RecipeId: exampleRecipeID})
		assert.NotNil(t, res)
		require.NoError(t, err)
		assert.Equal(t, exampleClonedRecipe.ID, res.Cloned.Id)

		assert.Len(t, mrm.CloneRecipeCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipe := mealplanningfakes.BuildFakeRecipe()

		mrm := &mockmanagers.MealPlanningManagerMock{
			CreateRecipeFunc: func(_ context.Context, creatorID string, _ *mealplanning.RecipeCreationRequestInput) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleUserID, creatorID)

				return exampleCreatedRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID: exampleUserID,
			},
		})

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeRequest](t)

		actual, err := s.CreateRecipe(ctx, exampleInput)
		assert.NotNil(t, actual)
		require.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipe.ID, actual.Created.Id)

		assert.Len(t, mrm.CreateRecipeCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipePrepTask := mealplanningfakes.BuildFakeRecipePrepTask()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			CreateRecipePrepTaskFunc: func(_ context.Context, recipeID string, _ *mealplanning.RecipePrepTaskCreationRequestInput) (*mealplanning.RecipePrepTask, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleCreatedRecipePrepTask, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipePrepTaskRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		actual, err := s.CreateRecipePrepTask(ctx, exampleInput)
		assert.NotNil(t, actual)
		require.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipePrepTask.ID, actual.Created.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.CreateRecipePrepTaskCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipePrepTaskRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		res, err := s.CreateRecipePrepTask(ctx, exampleInput)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeRating := mealplanningfakes.BuildFakeRecipeRating()

		mrm := &mockmanagers.MealPlanningManagerMock{
			CreateRecipeRatingFunc: func(_ context.Context, recipeID string, _ *mealplanning.RecipeRatingCreationRequestInput) (*mealplanning.RecipeRating, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleCreatedRecipeRating, nil
			},
		}
		s.mealPlanningManager = mrm

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID: exampleUserID,
			},
		})

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeRatingRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		actual, err := s.CreateRecipeRating(ctx, exampleInput)
		assert.NotNil(t, actual)
		require.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeRating.ID, actual.Created.Id)

		assert.Len(t, mrm.CreateRecipeRatingCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStep := mealplanningfakes.BuildFakeRecipeStep()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			CreateRecipeStepFunc: func(_ context.Context, recipeID string, _ *mealplanning.RecipeStepCreationRequestInput) (*mealplanning.RecipeStep, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleCreatedRecipeStep, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		actual, err := s.CreateRecipeStep(ctx, exampleInput)
		assert.NotNil(t, actual)
		require.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStep.ID, actual.Created.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.CreateRecipeStepCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		res, err := s.CreateRecipeStep(ctx, exampleInput)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepCompletionCondition := mealplanningfakes.BuildFakeRecipeStepCompletionCondition()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			CreateRecipeStepCompletionConditionFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput) (*mealplanning.RecipeStepCompletionCondition, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleCreatedRecipeStepCompletionCondition, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepCompletionConditionRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepCompletionCondition(ctx, exampleInput)
		assert.NotNil(t, actual)
		require.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepCompletionCondition.ID, actual.Created.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.CreateRecipeStepCompletionConditionCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepCompletionConditionRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepCompletionCondition(ctx, exampleInput)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepIngredient := mealplanningfakes.BuildFakeRecipeStepIngredient()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			CreateRecipeStepIngredientFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *mealplanning.RecipeStepIngredientCreationRequestInput) (*mealplanning.RecipeStepIngredient, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleCreatedRecipeStepIngredient, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepIngredientRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepIngredient(ctx, exampleInput)
		assert.NotNil(t, actual)
		require.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepIngredient.ID, actual.Created.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.CreateRecipeStepIngredientCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepIngredientRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepIngredient(ctx, exampleInput)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepInstrument := mealplanningfakes.BuildFakeRecipeStepInstrument()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			CreateRecipeStepInstrumentFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *mealplanning.RecipeStepInstrumentCreationRequestInput) (*mealplanning.RecipeStepInstrument, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleCreatedRecipeStepInstrument, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepInstrumentRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepInstrument(ctx, exampleInput)
		assert.NotNil(t, actual)
		require.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepInstrument.ID, actual.Created.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.CreateRecipeStepInstrumentCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepInstrumentRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepInstrument(ctx, exampleInput)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepProduct := mealplanningfakes.BuildFakeRecipeStepProduct()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			CreateRecipeStepProductFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *mealplanning.RecipeStepProductCreationRequestInput) (*mealplanning.RecipeStepProduct, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleCreatedRecipeStepProduct, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepProductRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepProduct(ctx, exampleInput)
		assert.NotNil(t, actual)
		require.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepProduct.ID, actual.Created.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.CreateRecipeStepProductCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepProductRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepProduct(ctx, exampleInput)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_CreateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepVessel := mealplanningfakes.BuildFakeRecipeStepVessel()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
			CreateRecipeStepVesselFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *mealplanning.RecipeStepVesselCreationRequestInput) (*mealplanning.RecipeStepVessel, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleCreatedRecipeStepVessel, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepVesselRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepVessel(ctx, exampleInput)
		assert.NotNil(t, actual)
		require.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepVessel.ID, actual.Created.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.CreateRecipeStepVesselCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepVesselRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepVessel(ctx, exampleInput)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_GetMermaidDiagramForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleMermaidDiagram := "graph TD\nA[Recipe]"

		mrm := &mockmanagers.MealPlanningManagerMock{
			RecipeMermaidFunc: func(_ context.Context, recipeID string) (string, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleMermaidDiagram, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetMermaidDiagramForRecipe(ctx, &mealplanninggrpc.GetMermaidDiagramForRecipeRequest{RecipeId: exampleRecipeID})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, exampleMermaidDiagram, result.Response)

		assert.Len(t, mrm.RecipeMermaidCalls(), 1)
	})
}

func TestServiceImpl_GetRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipe()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleResult.ID, recipeID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{RecipeId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_EstimateRecipePrepTasks(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleEstimatedPrepSteps := []*mealplanning.MealPlanTaskDatabaseCreationEstimate{
			{
				CreationExplanation: "test explanation",
			},
		}

		mrm := &mockmanagers.MealPlanningManagerMock{
			RecipeEstimatedPrepStepsFunc: func(_ context.Context, recipeID string) ([]*mealplanning.MealPlanTaskDatabaseCreationEstimate, error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleEstimatedPrepSteps, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.EstimateRecipePrepTasks(ctx, &mealplanninggrpc.EstimateRecipePrepTasksRequest{RecipeId: exampleRecipeID})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleEstimatedPrepSteps))

		assert.Len(t, mrm.RecipeEstimatedPrepStepsCalls(), 1)
	})
}

func TestServiceImpl_GetRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipePrepTask()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipePrepTaskFunc: func(_ context.Context, recipeID string, recipePrepTaskID string) (*mealplanning.RecipePrepTask, error) {
				assert.Equal(t, exampleResult.BelongsToRecipe, recipeID)
				assert.Equal(t, exampleResult.ID, recipePrepTaskID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipePrepTask(ctx, &mealplanninggrpc.GetRecipePrepTaskRequest{
			RecipeId:         exampleResult.BelongsToRecipe,
			RecipePrepTaskId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipePrepTaskCalls(), 1)
	})
}

func TestServiceImpl_GetRecipePrepTasks(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipePrepTasksList()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipePrepTaskFunc: func(_ context.Context, recipeID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipePrepTask], error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipePrepTasks(ctx, &mealplanninggrpc.GetRecipePrepTasksRequest{RecipeId: exampleRecipeID})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.ListRecipePrepTaskCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeRating()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeRatingFunc: func(_ context.Context, recipeID string, recipeRatingID string) (*mealplanning.RecipeRating, error) {
				assert.Equal(t, exampleResult.BelongsToRecipe, recipeID)
				assert.Equal(t, exampleResult.ID, recipeRatingID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeRating(ctx, &mealplanninggrpc.GetRecipeRatingRequest{
			RecipeId:       exampleResult.BelongsToRecipe,
			RecipeRatingId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeRatingCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeRatingsForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeRatingsList()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipeRatingsFunc: func(_ context.Context, recipeID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeRating], error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeRatingsForRecipe(ctx, &mealplanninggrpc.GetRecipeRatingsForRecipeRequest{RecipeId: exampleRecipeID})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.ListRecipeRatingsCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStep()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeStepFunc: func(_ context.Context, recipeID string, recipeStepID string) (*mealplanning.RecipeStep, error) {
				assert.Equal(t, exampleResult.BelongsToRecipe, recipeID)
				assert.Equal(t, exampleResult.ID, recipeStepID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStep(ctx, &mealplanninggrpc.GetRecipeStepRequest{
			RecipeId:     exampleResult.BelongsToRecipe,
			RecipeStepId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeStepCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepCompletionCondition()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeStepCompletionConditionFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepCompletionConditionID string) (*mealplanning.RecipeStepCompletionCondition, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleResult.BelongsToRecipeStep, recipeStepID)
				assert.Equal(t, exampleResult.ID, recipeStepCompletionConditionID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepCompletionCondition(ctx, &mealplanninggrpc.GetRecipeStepCompletionConditionRequest{
			RecipeId:                        exampleRecipeID,
			RecipeStepId:                    exampleResult.BelongsToRecipeStep,
			RecipeStepCompletionConditionId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeStepCompletionConditionCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepCompletionConditions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepCompletionConditionsList()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipeStepCompletionConditionsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepCompletionCondition], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepCompletionConditions(ctx, &mealplanninggrpc.GetRecipeStepCompletionConditionsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.ListRecipeStepCompletionConditionsCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepIngredient()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeStepIngredientFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string) (*mealplanning.RecipeStepIngredient, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleResult.BelongsToRecipeStep, recipeStepID)
				assert.Equal(t, exampleResult.ID, recipeStepIngredientID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepIngredient(ctx, &mealplanninggrpc.GetRecipeStepIngredientRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleResult.BelongsToRecipeStep,
			RecipeStepIngredientId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeStepIngredientCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepIngredientsList()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipeStepIngredientsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepIngredient], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepIngredients(ctx, &mealplanninggrpc.GetRecipeStepIngredientsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.ListRecipeStepIngredientsCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepInstrument()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeStepInstrumentFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepInstrumentID string) (*mealplanning.RecipeStepInstrument, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleResult.BelongsToRecipeStep, recipeStepID)
				assert.Equal(t, exampleResult.ID, recipeStepInstrumentID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepInstrument(ctx, &mealplanninggrpc.GetRecipeStepInstrumentRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleResult.BelongsToRecipeStep,
			RecipeStepInstrumentId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeStepInstrumentCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepInstrumentsList()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipeStepInstrumentsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepInstrument], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepInstruments(ctx, &mealplanninggrpc.GetRecipeStepInstrumentsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.ListRecipeStepInstrumentsCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepProduct()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeStepProductFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepProductID string) (*mealplanning.RecipeStepProduct, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleResult.BelongsToRecipeStep, recipeStepID)
				assert.Equal(t, exampleResult.ID, recipeStepProductID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepProduct(ctx, &mealplanninggrpc.GetRecipeStepProductRequest{
			RecipeId:            exampleRecipeID,
			RecipeStepId:        exampleResult.BelongsToRecipeStep,
			RecipeStepProductId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeStepProductCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepProducts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepProductsList()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipeStepProductsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepProduct], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepProducts(ctx, &mealplanninggrpc.GetRecipeStepProductsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.ListRecipeStepProductsCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepVessel()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeStepVesselFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepVesselID string) (*mealplanning.RecipeStepVessel, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleResult.BelongsToRecipeStep, recipeStepID)
				assert.Equal(t, exampleResult.ID, recipeStepVesselID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepVessel(ctx, &mealplanninggrpc.GetRecipeStepVesselRequest{
			RecipeId:           exampleRecipeID,
			RecipeStepId:       exampleResult.BelongsToRecipeStep,
			RecipeStepVesselId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		require.NoError(t, err)

		assert.Len(t, mrm.ReadRecipeStepVesselCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeStepVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepVesselsList()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipeStepVesselsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepVessel], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepVessels(ctx, &mealplanninggrpc.GetRecipeStepVesselsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.ListRecipeStepVesselsCalls(), 1)
	})
}

func TestServiceImpl_GetRecipeSteps(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepsList()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipeStepsFunc: func(_ context.Context, recipeID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStep], error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeSteps(ctx, &mealplanninggrpc.GetRecipeStepsRequest{RecipeId: exampleRecipeID})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.ListRecipeStepsCalls(), 1)
	})
}

func TestServiceImpl_GetRecipes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipesList()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			ListRecipesFunc: func(_ context.Context, status string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
				assert.Empty(t, status)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.GetRecipes(ctx, &mealplanninggrpc.GetRecipesRequest{})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.ListRecipesCalls(), 1)
	})
}

func TestServiceImpl_SearchForRecipes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipesList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForRecipesRequest](t)

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			SearchRecipesFunc: func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.SearchForRecipes(ctx, exampleRequest)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.SearchRecipesCalls(), 1)
	})
}

func TestServiceImpl_SearchForMealEligibleRecipes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipesList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForMealEligibleRecipesRequest](t)

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MealPlanningManagerMock{
			SearchForMealEligibleRecipesFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
				assert.Equal(t, exampleRequest.Query, query)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.SearchForMealEligibleRecipes(ctx, exampleRequest)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.SearchForMealEligibleRecipesCalls(), 1)
	})
}

func TestServiceImpl_SearchForRecipesWithInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipesList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForRecipesWithInstrumentOwnershipRequest](t)

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			ActiveAccountID: exampleAccountID,
		})

		mrm := &mockmanagers.MealPlanningManagerMock{
			SearchRecipesWithInstrumentOwnershipFunc: func(_ context.Context, accountID string, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
				assert.Equal(t, exampleAccountID, accountID)
				assert.Equal(t, exampleRequest.Query, query)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mrm

		result, err := s.SearchForRecipesWithInstrumentOwnership(ctx, exampleRequest)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mrm.SearchRecipesWithInstrumentOwnershipCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipe()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleResponse.CreatedByUser = exampleUserID

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleResponse, nil
			},
			UpdateRecipeFunc: func(_ context.Context, recipeID string, _ *mealplanning.RecipeUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipe(ctx, exampleRequest)
		require.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 2) // the service re-reads the record after updating it
		assert.Len(t, mrm.UpdateRecipeCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipe(ctx, exampleRequest)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipePrepTaskRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipePrepTask()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
			UpdateRecipePrepTaskFunc: func(_ context.Context, recipeID string, recipePrepTaskID string, _ *mealplanning.RecipePrepTaskUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipePrepTaskId, recipePrepTaskID)

				return nil
			},
			ReadRecipePrepTaskFunc: func(_ context.Context, recipeID string, recipePrepTaskID string) (*mealplanning.RecipePrepTask, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipePrepTaskId, recipePrepTaskID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipePrepTask(ctx, exampleRequest)
		require.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.UpdateRecipePrepTaskCalls(), 1)
		assert.Len(t, mrm.ReadRecipePrepTaskCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipePrepTaskRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipePrepTask(ctx, exampleRequest)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeRatingRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeRating()

		s := buildServiceImplForRecipesTest(t)
		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleResponse.CreatedByUser},
		})

		mrm := &mockmanagers.MealPlanningManagerMock{
			UpdateRecipeRatingFunc: func(_ context.Context, recipeID string, recipeRatingID string, _ *mealplanning.RecipeRatingUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeRatingId, recipeRatingID)

				return nil
			},
			ReadRecipeRatingFunc: func(_ context.Context, recipeID string, recipeRatingID string) (*mealplanning.RecipeRating, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeRatingId, recipeRatingID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeRating(ctx, exampleRequest)
		require.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mrm.UpdateRecipeRatingCalls(), 1)
		assert.Len(t, mrm.ReadRecipeRatingCalls(), 2) // the service re-reads the record after updating it
	})
}

func TestServiceImpl_UpdateRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStep()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
			UpdateRecipeStepFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *mealplanning.RecipeStepUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)

				return nil
			},
			ReadRecipeStepFunc: func(_ context.Context, recipeID string, recipeStepID string) (*mealplanning.RecipeStep, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStep(ctx, exampleRequest)
		require.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.UpdateRecipeStepCalls(), 1)
		assert.Len(t, mrm.ReadRecipeStepCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStep(ctx, exampleRequest)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepCompletionConditionRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepCompletionCondition()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
			UpdateRecipeStepCompletionConditionFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepCompletionConditionID string, _ *mealplanning.RecipeStepCompletionConditionUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepCompletionConditionId, recipeStepCompletionConditionID)

				return nil
			},
			ReadRecipeStepCompletionConditionFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepCompletionConditionID string) (*mealplanning.RecipeStepCompletionCondition, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepCompletionConditionId, recipeStepCompletionConditionID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepCompletionCondition(ctx, exampleRequest)
		require.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.UpdateRecipeStepCompletionConditionCalls(), 1)
		assert.Len(t, mrm.ReadRecipeStepCompletionConditionCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepCompletionConditionRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepCompletionCondition(ctx, exampleRequest)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepIngredientRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepIngredient()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
			UpdateRecipeStepIngredientFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string, _ *mealplanning.RecipeStepIngredientUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepIngredientId, recipeStepIngredientID)

				return nil
			},
			ReadRecipeStepIngredientFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string) (*mealplanning.RecipeStepIngredient, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepIngredientId, recipeStepIngredientID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepIngredient(ctx, exampleRequest)
		require.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.UpdateRecipeStepIngredientCalls(), 1)
		assert.Len(t, mrm.ReadRecipeStepIngredientCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepIngredientRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepIngredient(ctx, exampleRequest)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepInstrumentRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepInstrument()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
			UpdateRecipeStepInstrumentFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepInstrumentID string, _ *mealplanning.RecipeStepInstrumentUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepInstrumentId, recipeStepInstrumentID)

				return nil
			},
			ReadRecipeStepInstrumentFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepInstrumentID string) (*mealplanning.RecipeStepInstrument, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepInstrumentId, recipeStepInstrumentID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepInstrument(ctx, exampleRequest)
		require.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.UpdateRecipeStepInstrumentCalls(), 1)
		assert.Len(t, mrm.ReadRecipeStepInstrumentCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepInstrumentRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepInstrument(ctx, exampleRequest)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepProductRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepProduct()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
			UpdateRecipeStepProductFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepProductID string, _ *mealplanning.RecipeStepProductUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepProductId, recipeStepProductID)

				return nil
			},
			ReadRecipeStepProductFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepProductID string) (*mealplanning.RecipeStepProduct, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepProductId, recipeStepProductID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepProduct(ctx, exampleRequest)
		require.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.UpdateRecipeStepProductCalls(), 1)
		assert.Len(t, mrm.ReadRecipeStepProductCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepProductRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepProduct(ctx, exampleRequest)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}

func TestServiceImpl_UpdateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepVesselRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepVessel()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
			UpdateRecipeStepVesselFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepVesselID string, _ *mealplanning.RecipeStepVesselUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepVesselId, recipeStepVesselID)

				return nil
			},
			ReadRecipeStepVesselFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepVesselID string) (*mealplanning.RecipeStepVessel, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)
				assert.Equal(t, exampleRequest.RecipeStepId, recipeStepID)
				assert.Equal(t, exampleRequest.RecipeStepVesselId, recipeStepVesselID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepVessel(ctx, exampleRequest)
		require.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
		assert.Len(t, mrm.UpdateRecipeStepVesselCalls(), 1)
		assert.Len(t, mrm.ReadRecipeStepVesselCalls(), 1)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepVesselRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: exampleUserID},
		})

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRequest.RecipeId, recipeID)

				return exampleRecipe, nil
			},
		}
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepVessel(ctx, exampleRequest)
		assert.Nil(t, res)
		require.Error(t, err)

		assert.Len(t, mrm.ReadRecipeCalls(), 1)
	})
}
