package grpc

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mockmanagers "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	mealplanninggrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/database/filtering"
	"github.com/primandproper/platform-go/v3/fake"
	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	"github.com/primandproper/platform-go/v3/observability/tracing"
	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildServiceImplForRecipesTest(t *testing.T) *serviceImpl {
	t.Helper()

	return &serviceImpl{
		tracer:          tracing.NewTracerForTest(t.Name()),
		logger:          loggingnoop.NewLogger(),
		commentsManager: &noopCommentsManager{},
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: mealplanningfakes.BuildFakeID(),
				},
			}, nil
		},
	}
}

func TestServiceImpl_verifyRecipeOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		userID, err := s.verifyRecipeOwnership(ctx, exampleRecipeID, s.logger, span)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserID, userID)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		_, err := s.verifyRecipeOwnership(ctx, exampleRecipeID, s.logger, span)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipe), testutils.ContextMatcher, exampleRecipeID, exampleUserID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: exampleRecipeID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: exampleRecipeID})
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipePrepTaskID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipePrepTask), testutils.ContextMatcher, exampleRecipeID, exampleRecipePrepTaskID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipePrepTask(ctx, &mealplanninggrpc.ArchiveRecipePrepTaskRequest{
			RecipeId:         exampleRecipeID,
			RecipePrepTaskId: exampleRecipePrepTaskID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipePrepTaskID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipePrepTask(ctx, &mealplanninggrpc.ArchiveRecipePrepTaskRequest{
			RecipeId:         exampleRecipeID,
			RecipePrepTaskId: exampleRecipePrepTaskID,
		})
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeRatingID := mealplanningfakes.BuildFakeID()

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipeRating), testutils.ContextMatcher, exampleRecipeID, exampleRecipeRatingID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       exampleRecipeID,
			RecipeRatingId: exampleRecipeRatingID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeLists(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		list := &mealplanning.RecipeList{ID: mealplanningfakes.BuildFakeID()}
		expected := &filtering.QueryFilteredResult[mealplanning.RecipeList]{Data: []*mealplanning.RecipeList{list}}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipeLists), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
		s.mealPlanningManager = mrm

		res, err := s.GetRecipeLists(ctx, &mealplanninggrpc.GetRecipeListsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Results, 1)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		userID := mealplanningfakes.BuildFakeID()
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		input := &mealplanninggrpc.RecipeListCreationRequestInput{Name: t.Name(), Description: "desc"}
		created := &mealplanning.RecipeList{ID: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.CreateRecipeList), testutils.ContextMatcher, userID, testutils.MatchType[*mealplanning.RecipeListCreationRequestInput]()).Return(created, nil)
		s.mealPlanningManager = mrm

		res, err := s.CreateRecipeList(ctx, &mealplanninggrpc.CreateRecipeListRequest{Input: input})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, created.ID, res.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		userID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		name := t.Name()
		desc := "desc"
		input := &mealplanninggrpc.RecipeListUpdateRequestInput{
			Name:        &name,
			Description: &desc,
		}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipeList), testutils.ContextMatcher, listID, userID, testutils.MatchType[*mealplanning.RecipeListUpdateRequestInput]()).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeList(ctx, &mealplanninggrpc.UpdateRecipeListRequest{
			RecipeListId: listID,
			Input:        input,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		userID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipeList), testutils.ContextMatcher, listID, userID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeList(ctx, &mealplanninggrpc.ArchiveRecipeListRequest{RecipeListId: listID})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeListItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		listID := mealplanningfakes.BuildFakeID()
		item := &mealplanning.RecipeListItem{ID: mealplanningfakes.BuildFakeID(), Recipe: mealplanning.Recipe{ID: mealplanningfakes.BuildFakeID()}}
		expected := &filtering.QueryFilteredResult[mealplanning.RecipeListItem]{Data: []*mealplanning.RecipeListItem{item}}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipeListItems), testutils.ContextMatcher, listID, testutils.QueryFilterMatcher).Return(expected, nil)
		s.mealPlanningManager = mrm

		res, err := s.GetRecipeListItems(ctx, &mealplanninggrpc.GetRecipeListItemsRequest{RecipeListId: listID})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Results, 1)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipeListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		listID := mealplanningfakes.BuildFakeID()
		recipeID := mealplanningfakes.BuildFakeID()
		input := &mealplanninggrpc.RecipeListItemCreationRequestInput{
			BelongsToRecipeList: listID,
			RecipeId:            recipeID,
			Notes:               t.Name(),
		}

		created := &mealplanning.RecipeListItem{ID: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.AddRecipeToRecipeList), testutils.ContextMatcher, listID, recipeID, input.Notes).Return(created, nil)
		s.mealPlanningManager = mrm

		res, err := s.CreateRecipeListItem(ctx, &mealplanninggrpc.CreateRecipeListItemRequest{Input: input})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, created.ID, res.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipeListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
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

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipeListItem), testutils.ContextMatcher, itemID, listID, recipeID, testutils.MatchType[*mealplanning.RecipeListItemUpdateRequestInput]()).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeListItem(ctx, &mealplanninggrpc.UpdateRecipeListItemRequest{
			RecipeListItemId: itemID,
			Input:            input,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipeListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		itemID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.RemoveRecipeFromRecipeList), testutils.ContextMatcher, listID, itemID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeListItem(ctx, &mealplanninggrpc.ArchiveRecipeListItemRequest{
			RecipeListItemId: itemID,
			RecipeListId:     listID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipeStep), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepCompletionConditionID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipeStepCompletionCondition), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepCompletionConditionID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepCompletionCondition(ctx, &mealplanninggrpc.ArchiveRecipeStepCompletionConditionRequest{
			RecipeId:                        exampleRecipeID,
			RecipeStepId:                    exampleRecipeStepID,
			RecipeStepCompletionConditionId: exampleRecipeStepCompletionConditionID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepCompletionConditionID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepCompletionCondition(ctx, &mealplanninggrpc.ArchiveRecipeStepCompletionConditionRequest{
			RecipeId:                        exampleRecipeID,
			RecipeStepId:                    exampleRecipeStepID,
			RecipeStepCompletionConditionId: exampleRecipeStepCompletionConditionID,
		})
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepIngredientID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipeStepIngredient), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepIngredientID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepIngredient(ctx, &mealplanninggrpc.ArchiveRecipeStepIngredientRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleRecipeStepID,
			RecipeStepIngredientId: exampleRecipeStepIngredientID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepIngredientID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepIngredient(ctx, &mealplanninggrpc.ArchiveRecipeStepIngredientRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleRecipeStepID,
			RecipeStepIngredientId: exampleRecipeStepIngredientID,
		})
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepInstrumentID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipeStepInstrument), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepInstrumentID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepInstrument(ctx, &mealplanninggrpc.ArchiveRecipeStepInstrumentRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleRecipeStepID,
			RecipeStepInstrumentId: exampleRecipeStepInstrumentID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepInstrumentID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepInstrument(ctx, &mealplanninggrpc.ArchiveRecipeStepInstrumentRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleRecipeStepID,
			RecipeStepInstrumentId: exampleRecipeStepInstrumentID,
		})
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepProductID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipeStepProduct), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepProductID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepProduct(ctx, &mealplanninggrpc.ArchiveRecipeStepProductRequest{
			RecipeId:            exampleRecipeID,
			RecipeStepId:        exampleRecipeStepID,
			RecipeStepProductId: exampleRecipeStepProductID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepProductID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepProduct(ctx, &mealplanninggrpc.ArchiveRecipeStepProductRequest{
			RecipeId:            exampleRecipeID,
			RecipeStepId:        exampleRecipeStepID,
			RecipeStepProductId: exampleRecipeStepProductID,
		})
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_ArchiveRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepVesselID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.ArchiveRecipeStepVessel), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepVesselID).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepVessel(ctx, &mealplanninggrpc.ArchiveRecipeStepVesselRequest{
			RecipeId:           exampleRecipeID,
			RecipeStepId:       exampleRecipeStepID,
			RecipeStepVesselId: exampleRecipeStepVesselID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepVesselID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.ArchiveRecipeStepVessel(ctx, &mealplanninggrpc.ArchiveRecipeStepVesselRequest{
			RecipeId:           exampleRecipeID,
			RecipeStepId:       exampleRecipeStepID,
			RecipeStepVesselId: exampleRecipeStepVesselID,
		})
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CloneRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleClonedRecipe := mealplanningfakes.BuildFakeRecipe()

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.CloneRecipe), testutils.ContextMatcher, exampleRecipeID, exampleUserID).Return(exampleClonedRecipe, nil)
		s.mealPlanningManager = mrm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		res, err := s.CloneRecipe(ctx, &mealplanninggrpc.CloneRecipeRequest{RecipeId: exampleRecipeID})
		assert.NotNil(t, res)
		assert.NoError(t, err)
		assert.Equal(t, exampleClonedRecipe.ID, res.Cloned.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipe := mealplanningfakes.BuildFakeRecipe()

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.CreateRecipe), testutils.ContextMatcher, exampleUserID, testutils.MatchType[*mealplanning.RecipeCreationRequestInput]()).Return(exampleCreatedRecipe, nil)
		s.mealPlanningManager = mrm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeRequest](t)

		actual, err := s.CreateRecipe(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipe.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipePrepTask := mealplanningfakes.BuildFakeRecipePrepTask()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.CreateRecipePrepTask), testutils.ContextMatcher, exampleRecipeID, testutils.MatchType[*mealplanning.RecipePrepTaskCreationRequestInput]()).Return(exampleCreatedRecipePrepTask, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipePrepTaskRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		actual, err := s.CreateRecipePrepTask(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipePrepTask.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipePrepTaskRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		res, err := s.CreateRecipePrepTask(ctx, exampleInput)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeRating := mealplanningfakes.BuildFakeRecipeRating()

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.CreateRecipeRating), testutils.ContextMatcher, exampleRecipeID, testutils.MatchType[*mealplanning.RecipeRatingCreationRequestInput]()).Return(exampleCreatedRecipeRating, nil)
		s.mealPlanningManager = mrm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeRatingRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		actual, err := s.CreateRecipeRating(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeRating.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStep := mealplanningfakes.BuildFakeRecipeStep()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.CreateRecipeStep), testutils.ContextMatcher, exampleRecipeID, testutils.MatchType[*mealplanning.RecipeStepCreationRequestInput]()).Return(exampleCreatedRecipeStep, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		actual, err := s.CreateRecipeStep(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStep.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepRequest](t)
		exampleInput.RecipeId = exampleRecipeID

		res, err := s.CreateRecipeStep(ctx, exampleInput)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepCompletionCondition := mealplanningfakes.BuildFakeRecipeStepCompletionCondition()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.CreateRecipeStepCompletionCondition), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.MatchType[*mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput]()).Return(exampleCreatedRecipeStepCompletionCondition, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepCompletionConditionRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepCompletionCondition(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepCompletionCondition.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepCompletionConditionRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepCompletionCondition(ctx, exampleInput)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepIngredient := mealplanningfakes.BuildFakeRecipeStepIngredient()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.CreateRecipeStepIngredient), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.MatchType[*mealplanning.RecipeStepIngredientCreationRequestInput]()).Return(exampleCreatedRecipeStepIngredient, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepIngredientRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepIngredient(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepIngredient.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepIngredientRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepIngredient(ctx, exampleInput)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepInstrument := mealplanningfakes.BuildFakeRecipeStepInstrument()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.CreateRecipeStepInstrument), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.MatchType[*mealplanning.RecipeStepInstrumentCreationRequestInput]()).Return(exampleCreatedRecipeStepInstrument, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepInstrumentRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepInstrument(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepInstrument.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepInstrumentRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepInstrument(ctx, exampleInput)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepProduct := mealplanningfakes.BuildFakeRecipeStepProduct()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.CreateRecipeStepProduct), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.MatchType[*mealplanning.RecipeStepProductCreationRequestInput]()).Return(exampleCreatedRecipeStepProduct, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepProductRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepProduct(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepProduct.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepProductRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepProduct(ctx, exampleInput)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_CreateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleCreatedRecipeStepVessel := mealplanningfakes.BuildFakeRecipeStepVessel()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.CreateRecipeStepVessel), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.MatchType[*mealplanning.RecipeStepVesselCreationRequestInput]()).Return(exampleCreatedRecipeStepVessel, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepVesselRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		actual, err := s.CreateRecipeStepVessel(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedRecipeStepVessel.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRecipeID, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRecipeID).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateRecipeStepVesselRequest](t)
		exampleInput.RecipeId = exampleRecipeID
		exampleInput.RecipeStepId = exampleRecipeStepID

		res, err := s.CreateRecipeStepVessel(ctx, exampleInput)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetMermaidDiagramForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleMermaidDiagram := "graph TD\nA[Recipe]"

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.RecipeMermaid), testutils.ContextMatcher, exampleRecipeID).Return(exampleMermaidDiagram, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetMermaidDiagramForRecipe(ctx, &mealplanninggrpc.GetMermaidDiagramForRecipeRequest{RecipeId: exampleRecipeID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, exampleMermaidDiagram, result.Response)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipe()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{RecipeId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_EstimateRecipePrepTasks(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleEstimatedPrepSteps := []*mealplanning.MealPlanTaskDatabaseCreationEstimate{
			{
				CreationExplanation: "test explanation",
			},
		}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.RecipeEstimatedPrepSteps), testutils.ContextMatcher, exampleRecipeID).Return(exampleEstimatedPrepSteps, nil)
		s.mealPlanningManager = mrm

		result, err := s.EstimateRecipePrepTasks(ctx, &mealplanninggrpc.EstimateRecipePrepTasksRequest{RecipeId: exampleRecipeID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleEstimatedPrepSteps))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipePrepTask()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipePrepTask), testutils.ContextMatcher, exampleResult.BelongsToRecipe, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipePrepTask(ctx, &mealplanninggrpc.GetRecipePrepTaskRequest{
			RecipeId:         exampleResult.BelongsToRecipe,
			RecipePrepTaskId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipePrepTasks(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipePrepTasksList()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipePrepTask), testutils.ContextMatcher, exampleRecipeID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipePrepTasks(ctx, &mealplanninggrpc.GetRecipePrepTasksRequest{RecipeId: exampleRecipeID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeRating()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeRating), testutils.ContextMatcher, exampleResult.BelongsToRecipe, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeRating(ctx, &mealplanninggrpc.GetRecipeRatingRequest{
			RecipeId:       exampleResult.BelongsToRecipe,
			RecipeRatingId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeRatingsForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeRatingsList()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipeRatings), testutils.ContextMatcher, exampleRecipeID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeRatingsForRecipe(ctx, &mealplanninggrpc.GetRecipeRatingsForRecipeRequest{RecipeId: exampleRecipeID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStep()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStep), testutils.ContextMatcher, exampleResult.BelongsToRecipe, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStep(ctx, &mealplanninggrpc.GetRecipeStepRequest{
			RecipeId:     exampleResult.BelongsToRecipe,
			RecipeStepId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepCompletionCondition()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepCompletionCondition), testutils.ContextMatcher, exampleRecipeID, exampleResult.BelongsToRecipeStep, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepCompletionCondition(ctx, &mealplanninggrpc.GetRecipeStepCompletionConditionRequest{
			RecipeId:                        exampleRecipeID,
			RecipeStepId:                    exampleResult.BelongsToRecipeStep,
			RecipeStepCompletionConditionId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepCompletionConditions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepCompletionConditionsList()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipeStepCompletionConditions), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepCompletionConditions(ctx, &mealplanninggrpc.GetRecipeStepCompletionConditionsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepIngredient()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepIngredient), testutils.ContextMatcher, exampleRecipeID, exampleResult.BelongsToRecipeStep, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepIngredient(ctx, &mealplanninggrpc.GetRecipeStepIngredientRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleResult.BelongsToRecipeStep,
			RecipeStepIngredientId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepIngredientsList()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipeStepIngredients), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepIngredients(ctx, &mealplanninggrpc.GetRecipeStepIngredientsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepInstrument()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepInstrument), testutils.ContextMatcher, exampleRecipeID, exampleResult.BelongsToRecipeStep, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepInstrument(ctx, &mealplanninggrpc.GetRecipeStepInstrumentRequest{
			RecipeId:               exampleRecipeID,
			RecipeStepId:           exampleResult.BelongsToRecipeStep,
			RecipeStepInstrumentId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepInstrumentsList()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipeStepInstruments), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepInstruments(ctx, &mealplanninggrpc.GetRecipeStepInstrumentsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepProduct()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepProduct), testutils.ContextMatcher, exampleRecipeID, exampleResult.BelongsToRecipeStep, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepProduct(ctx, &mealplanninggrpc.GetRecipeStepProductRequest{
			RecipeId:            exampleRecipeID,
			RecipeStepId:        exampleResult.BelongsToRecipeStep,
			RecipeStepProductId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepProducts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepProductsList()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipeStepProducts), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepProducts(ctx, &mealplanninggrpc.GetRecipeStepProductsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipeStepVessel()
		exampleRecipeID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepVessel), testutils.ContextMatcher, exampleRecipeID, exampleResult.BelongsToRecipeStep, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepVessel(ctx, &mealplanninggrpc.GetRecipeStepVesselRequest{
			RecipeId:           exampleRecipeID,
			RecipeStepId:       exampleResult.BelongsToRecipeStep,
			RecipeStepVesselId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeStepVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleRecipeStepID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepVesselsList()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipeStepVessels), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeStepVessels(ctx, &mealplanninggrpc.GetRecipeStepVesselsRequest{
			RecipeId:     exampleRecipeID,
			RecipeStepId: exampleRecipeStepID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipeSteps(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipeStepsList()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipeSteps), testutils.ContextMatcher, exampleRecipeID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipeSteps(ctx, &mealplanninggrpc.GetRecipeStepsRequest{RecipeId: exampleRecipeID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_GetRecipes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipesList()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ListRecipes), testutils.ContextMatcher, "", testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.GetRecipes(ctx, &mealplanninggrpc.GetRecipesRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_SearchForRecipes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipesList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForRecipesRequest](t)

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.SearchRecipes), testutils.ContextMatcher, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.SearchForRecipes(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_SearchForMealEligibleRecipes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeRecipesList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForMealEligibleRecipesRequest](t)

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.SearchForMealEligibleRecipes), testutils.ContextMatcher, exampleRequest.Query, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.SearchForMealEligibleRecipes(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_SearchForRecipesWithInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeRecipesList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForRecipesWithInstrumentOwnershipRequest](t)

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.SearchRecipesWithInstrumentOwnership), testutils.ContextMatcher, exampleAccountID, exampleRequest.Query, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mrm

		result, err := s.SearchForRecipesWithInstrumentOwnership(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipe()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleResponse.CreatedByUser = exampleUserID

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleResponse, nil)
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipe), testutils.ContextMatcher, exampleRequest.RecipeId, testutils.MatchType[*mealplanning.RecipeUpdateRequestInput]()).Return(nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipe(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipe(ctx, exampleRequest)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipePrepTaskRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipePrepTask()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipePrepTask), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipePrepTaskId, testutils.MatchType[*mealplanning.RecipePrepTaskUpdateRequestInput]()).Return(nil)
		mrm.On(reflection.GetMethodName(mrm.ReadRecipePrepTask), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipePrepTaskId).Return(exampleResponse, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipePrepTask(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipePrepTaskRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipePrepTask(ctx, exampleRequest)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeRatingRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeRating()

		s := buildServiceImplForRecipesTest(t)

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipeRating), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeRatingId, testutils.MatchType[*mealplanning.RecipeRatingUpdateRequestInput]()).Return(nil)
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeRating), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeRatingId).Return(exampleResponse, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeRating(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStep()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipeStep), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, testutils.MatchType[*mealplanning.RecipeStepUpdateRequestInput]()).Return(nil)
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStep), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId).Return(exampleResponse, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStep(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStep(ctx, exampleRequest)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepCompletionConditionRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepCompletionCondition()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipeStepCompletionCondition), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepCompletionConditionId, testutils.MatchType[*mealplanning.RecipeStepCompletionConditionUpdateRequestInput]()).Return(nil)
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepCompletionCondition), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepCompletionConditionId).Return(exampleResponse, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepCompletionCondition(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepCompletionConditionRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepCompletionCondition(ctx, exampleRequest)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepIngredientRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepIngredient()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipeStepIngredient), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepIngredientId, testutils.MatchType[*mealplanning.RecipeStepIngredientUpdateRequestInput]()).Return(nil)
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepIngredient), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepIngredientId).Return(exampleResponse, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepIngredient(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepIngredientRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepIngredient(ctx, exampleRequest)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepInstrumentRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepInstrument()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipeStepInstrument), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepInstrumentId, testutils.MatchType[*mealplanning.RecipeStepInstrumentUpdateRequestInput]()).Return(nil)
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepInstrument), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepInstrumentId).Return(exampleResponse, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepInstrument(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepInstrumentRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepInstrument(ctx, exampleRequest)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepProductRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepProduct()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipeStepProduct), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepProductId, testutils.MatchType[*mealplanning.RecipeStepProductUpdateRequestInput]()).Return(nil)
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepProduct), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepProductId).Return(exampleResponse, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepProduct(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepProductRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepProduct(ctx, exampleRequest)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}

func TestServiceImpl_UpdateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepVesselRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeRecipeStepVessel()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForRecipesTest(t)

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: exampleUserID}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		mrm.On(reflection.GetMethodName(mrm.UpdateRecipeStepVessel), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepVesselId, testutils.MatchType[*mealplanning.RecipeStepVesselUpdateRequestInput]()).Return(nil)
		mrm.On(reflection.GetMethodName(mrm.ReadRecipeStepVessel), testutils.ContextMatcher, exampleRequest.RecipeId, exampleRequest.RecipeStepId, exampleRequest.RecipeStepVesselId).Return(exampleResponse, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepVessel(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mrm)
	})

	T.Run("returns permission denied for non-owner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForRecipesTest(t)

		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateRecipeStepVesselRequest](t)
		exampleUserID := mealplanningfakes.BuildFakeID()

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: exampleUserID},
			}, nil
		}

		exampleRecipe := &mealplanning.Recipe{ID: exampleRequest.RecipeId, CreatedByUser: mealplanningfakes.BuildFakeID()}

		mrm := &mockmanagers.MockMealPlanningManager{}
		mrm.On(reflection.GetMethodName(mrm.ReadRecipe), testutils.ContextMatcher, exampleRequest.RecipeId).Return(exampleRecipe, nil)
		s.mealPlanningManager = mrm

		res, err := s.UpdateRecipeStepVessel(ctx, exampleRequest)
		assert.Nil(t, res)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mrm)
	})
}
