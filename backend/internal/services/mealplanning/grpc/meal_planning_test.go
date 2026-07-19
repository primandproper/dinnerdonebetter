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

	"github.com/primandproper/platform-go/v5/fake"
	"github.com/primandproper/platform-go/v5/filtering"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildServiceImplForMealPlanningTest(t *testing.T) *serviceImpl {
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
				ActiveAccountID: mealplanningfakes.BuildFakeID(),
			}, nil
		},
		// Workers are set to nil for most tests since they're only used in specific worker methods
		mealPlanFinalizerWorker:              nil,
		mealPlanGroceryListInitializerWorker: nil,
		mealPlanTaskCreatorWorker:            nil,
	}
}

func TestServiceImpl_ArchiveMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ArchiveMeal), testutils.ContextMatcher, exampleMealID, exampleUserID).Return(nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		res, err := s.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: exampleMealID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_ArchiveMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleAccountID := mealplanningfakes.BuildFakeID()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ArchiveMealPlan), testutils.ContextMatcher, exampleMealPlanID, exampleAccountID).Return(nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		res, err := s.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: exampleMealPlanID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_ArchiveMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleMealPlanEventID := mealplanningfakes.BuildFakeID()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ArchiveMealPlanEvent), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID).Return(nil)
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealPlanEvent(ctx, &mealplanninggrpc.ArchiveMealPlanEventRequest{
			MealPlanId:      exampleMealPlanID,
			MealPlanEventId: exampleMealPlanEventID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_ArchiveMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleMealPlanGroceryListItemID := mealplanningfakes.BuildFakeID()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ArchiveMealPlanGroceryListItem), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanGroceryListItemID).Return(nil)
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealPlanGroceryListItem(ctx, &mealplanninggrpc.ArchiveMealPlanGroceryListItemRequest{
			MealPlanId:                exampleMealPlanID,
			MealPlanGroceryListItemId: exampleMealPlanGroceryListItemID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_ArchiveMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleMealPlanEventID := mealplanningfakes.BuildFakeID()
		exampleMealPlanOptionID := mealplanningfakes.BuildFakeID()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ArchiveMealPlanOption), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID).Return(nil)
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealPlanOption(ctx, &mealplanninggrpc.ArchiveMealPlanOptionRequest{
			MealPlanId:       exampleMealPlanID,
			MealPlanEventId:  exampleMealPlanEventID,
			MealPlanOptionId: exampleMealPlanOptionID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_ArchiveMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleMealPlanEventID := mealplanningfakes.BuildFakeID()
		exampleMealPlanOptionID := mealplanningfakes.BuildFakeID()
		exampleMealPlanOptionVoteID := mealplanningfakes.BuildFakeID()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ArchiveMealPlanOptionVote), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, exampleMealPlanOptionVoteID).Return(nil)
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealPlanOptionVote(ctx, &mealplanninggrpc.ArchiveMealPlanOptionVoteRequest{
			MealPlanId:           exampleMealPlanID,
			MealPlanEventId:      exampleMealPlanEventID,
			MealPlanOptionId:     exampleMealPlanOptionID,
			MealPlanOptionVoteId: exampleMealPlanOptionVoteID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_ArchiveUserIngredientPreference(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleUserIngredientPreferenceID := mealplanningfakes.BuildFakeID()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ArchiveUserIngredientPreference), testutils.ContextMatcher, exampleUserID, exampleUserIngredientPreferenceID).Return(nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		res, err := s.ArchiveUserIngredientPreference(ctx, &mealplanninggrpc.ArchiveUserIngredientPreferenceRequest{
			UserIngredientPreferenceId: exampleUserIngredientPreferenceID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealLists(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		list := &mealplanning.MealList{ID: mealplanningfakes.BuildFakeID()}
		expected := &filtering.QueryFilteredResult[mealplanning.MealList]{Data: []*mealplanning.MealList{list}}

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListMealLists), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
		s.mealPlanningManager = mmpm

		res, err := s.GetMealLists(ctx, &mealplanninggrpc.GetMealListsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Results, 1)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateMealList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		userID := mealplanningfakes.BuildFakeID()
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		input := &mealplanninggrpc.MealListCreationRequestInput{Name: t.Name(), Description: "desc"}
		created := &mealplanning.MealList{ID: mealplanningfakes.BuildFakeID()}

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.CreateMealList), testutils.ContextMatcher, userID, testutils.MatchType[*mealplanning.MealListCreationRequestInput]()).Return(created, nil)
		s.mealPlanningManager = mmpm

		res, err := s.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{Input: input})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, created.ID, res.Created.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateMealList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		userID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		name := t.Name()
		desc := "desc"
		input := &mealplanninggrpc.MealListUpdateRequestInput{
			Name:        &name,
			Description: &desc,
		}

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.UpdateMealList), testutils.ContextMatcher, listID, userID, testutils.MatchType[*mealplanning.MealListUpdateRequestInput]()).Return(nil)
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealList(ctx, &mealplanninggrpc.UpdateMealListRequest{
			MealListId: listID,
			Input:      input,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_ArchiveMealList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		userID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ArchiveMealList), testutils.ContextMatcher, listID, userID).Return(nil)
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealList(ctx, &mealplanninggrpc.ArchiveMealListRequest{MealListId: listID})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealListItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		listID := mealplanningfakes.BuildFakeID()
		item := &mealplanning.MealListItem{ID: mealplanningfakes.BuildFakeID(), Meal: mealplanning.Meal{ID: mealplanningfakes.BuildFakeID()}}
		expected := &filtering.QueryFilteredResult[mealplanning.MealListItem]{Data: []*mealplanning.MealListItem{item}}

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListMealListItems), testutils.ContextMatcher, listID, testutils.QueryFilterMatcher).Return(expected, nil)
		s.mealPlanningManager = mmpm

		res, err := s.GetMealListItems(ctx, &mealplanninggrpc.GetMealListItemsRequest{MealListId: listID})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Results, 1)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateMealListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		listID := mealplanningfakes.BuildFakeID()
		mealID := mealplanningfakes.BuildFakeID()
		input := &mealplanninggrpc.MealListItemCreationRequestInput{
			BelongsToMealList: listID,
			MealId:            mealID,
			Notes:             t.Name(),
		}

		created := &mealplanning.MealListItem{ID: mealplanningfakes.BuildFakeID()}

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.AddMealToMealList), testutils.ContextMatcher, listID, mealID, input.Notes).Return(created, nil)
		s.mealPlanningManager = mmpm

		res, err := s.CreateMealListItem(ctx, &mealplanninggrpc.CreateMealListItemRequest{Input: input})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, created.ID, res.Created.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateMealListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		itemID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()
		mealID := mealplanningfakes.BuildFakeID()
		notes := new(t.Name())
		input := &mealplanninggrpc.MealListItemUpdateRequestInput{
			BelongsToMealList: &listID,
			MealId:            &mealID,
			Notes:             notes,
		}

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.UpdateMealListItem), testutils.ContextMatcher, itemID, listID, mealID, testutils.MatchType[*mealplanning.MealListItemUpdateRequestInput]()).Return(nil)
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealListItem(ctx, &mealplanninggrpc.UpdateMealListItemRequest{
			MealListItemId: itemID,
			Input:          input,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_ArchiveMealListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		itemID := mealplanningfakes.BuildFakeID()
		listID := mealplanningfakes.BuildFakeID()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.RemoveMealFromMealList), testutils.ContextMatcher, listID, itemID).Return(nil)
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealListItem(ctx, &mealplanninggrpc.ArchiveMealListItemRequest{
			MealListItemId: itemID,
			MealListId:     listID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleCreatedMeal := mealplanningfakes.BuildFakeMeal()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.CreateMeal), testutils.ContextMatcher, exampleUserID, testutils.MatchType[*mealplanning.MealCreationRequestInput]()).Return(exampleCreatedMeal, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateMealRequest](t)

		actual, err := s.CreateMeal(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedMeal.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleCreatedMealPlan := mealplanningfakes.BuildFakeMealPlan()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.CreateMealPlan), testutils.ContextMatcher, exampleAccountID, exampleUserID, testutils.MatchType[*mealplanning.MealPlanCreationRequestInput]()).Return(exampleCreatedMealPlan, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific IDs
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateMealPlanRequest](t)

		actual, err := s.CreateMealPlan(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedMealPlan.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleCreatedMealPlanEvent := mealplanningfakes.BuildFakeMealPlanEvent()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.CreateMealPlanEvent), testutils.ContextMatcher, exampleMealPlanID, testutils.MatchType[*mealplanning.MealPlanEventCreationRequestInput]()).Return(exampleCreatedMealPlanEvent, nil)
		s.mealPlanningManager = mmpm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateMealPlanEventRequest](t)
		exampleInput.MealPlanId = exampleMealPlanID

		actual, err := s.CreateMealPlanEvent(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedMealPlanEvent.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleMealPlanEventID := mealplanningfakes.BuildFakeID()
		exampleCreatedMealPlanOption := mealplanningfakes.BuildFakeMealPlanOption()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.CreateMealPlanOptionWithEventID), testutils.ContextMatcher, exampleMealPlanEventID, testutils.MatchType[*mealplanning.MealPlanOptionCreationRequestInput]()).Return(exampleCreatedMealPlanOption, nil)
		s.mealPlanningManager = mmpm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateMealPlanOptionRequest](t)
		exampleInput.MealPlanId = exampleMealPlanID
		exampleInput.MealPlanEventId = exampleMealPlanEventID

		actual, err := s.CreateMealPlanOption(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedMealPlanOption.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleCreatedMealPlanOptionVotes := []*mealplanning.MealPlanOptionVote{
			mealplanningfakes.BuildFakeMealPlanOptionVote(),
		}

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.CreateMealPlanOptionVotes), testutils.ContextMatcher, exampleUserID, testutils.MatchType[*mealplanning.MealPlanOptionVoteCreationRequestInput]()).Return(exampleCreatedMealPlanOptionVotes, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateMealPlanOptionVoteRequest](t)
		exampleInput.MealPlanId = exampleMealPlanID

		actual, err := s.CreateMealPlanOptionVote(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Len(t, actual.Created, len(exampleCreatedMealPlanOptionVotes))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateMealPlanTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleCreatedMealPlanTask := mealplanningfakes.BuildFakeMealPlanTask()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.CreateMealPlanTask), testutils.ContextMatcher, testutils.MatchType[*mealplanning.MealPlanTaskCreationRequestInput]()).Return(exampleCreatedMealPlanTask, nil)
		s.mealPlanningManager = mmpm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateMealPlanTaskRequest](t)
		exampleInput.MealPlanId = exampleMealPlanID

		actual, err := s.CreateMealPlanTask(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedMealPlanTask.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateUserIngredientPreference(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleCreatedUserIngredientPreferences := []*mealplanning.UserIngredientPreference{
			mealplanningfakes.BuildFakeUserIngredientPreference(),
		}

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.CreateUserIngredientPreference), testutils.ContextMatcher, exampleUserID, testutils.MatchType[*mealplanning.UserIngredientPreferenceCreationRequestInput]()).Return(exampleCreatedUserIngredientPreferences, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateUserIngredientPreferenceRequest](t)

		actual, err := s.CreateUserIngredientPreference(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Len(t, actual.Created, len(exampleCreatedUserIngredientPreferences))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_FinalizeMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleFinalized := true

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.FinalizeMealPlan), testutils.ContextMatcher, exampleMealPlanID, exampleAccountID).Return(exampleFinalized, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		res, err := s.FinalizeMealPlan(ctx, &mealplanninggrpc.FinalizeMealPlanRequest{MealPlanId: exampleMealPlanID})
		assert.NotNil(t, res)
		assert.NoError(t, err)
		assert.Equal(t, exampleFinalized, res.Finalized)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMermaidDiagramForMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleMeal := mealplanningfakes.BuildFakeMeal()
		exampleMermaidDiagram := "flowchart TD;\n\tStep1[\"Main\"];\n\tStep100001[\"Side\"];\n"

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadMeal), testutils.ContextMatcher, exampleMeal.ID).Return(exampleMeal, nil)
		mmpm.On(reflection.GetMethodName(mmpm.MealMermaid), testutils.ContextMatcher, exampleMeal).Return(exampleMermaidDiagram, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMermaidDiagramForMeal(ctx, &mealplanninggrpc.GetMermaidDiagramForMealRequest{MealId: exampleMeal.ID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, exampleMermaidDiagram, result.Response)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMeal()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadMeal), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMeal(ctx, &mealplanninggrpc.GetMealRequest{MealId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealPlan()
		exampleAccountID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlan), testutils.ContextMatcher, exampleResult.ID, exampleAccountID).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		result, err := s.GetMealPlan(ctx, &mealplanninggrpc.GetMealPlanRequest{MealPlanId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlansForAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeMealPlansList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListMealPlans), testutils.ContextMatcher, exampleAccountID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		result, err := s.GetMealPlansForAccount(ctx, &mealplanninggrpc.GetMealPlansForAccountRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealPlanEvent()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanEvent), testutils.ContextMatcher, exampleResult.BelongsToMealPlan, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanEvent(ctx, &mealplanninggrpc.GetMealPlanEventRequest{
			MealPlanId:      exampleResult.BelongsToMealPlan,
			MealPlanEventId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanEvents(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeMealPlanEventsList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListMealPlanEvents), testutils.ContextMatcher, exampleMealPlanID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanEvents(ctx, &mealplanninggrpc.GetMealPlanEventsRequest{MealPlanId: exampleMealPlanID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealPlanGroceryListItem()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanGroceryListItem), testutils.ContextMatcher, exampleResult.BelongsToMealPlan, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanGroceryListItem(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemRequest{
			MealPlanId:                exampleResult.BelongsToMealPlan,
			MealPlanGroceryListItemId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanGroceryListItemsForMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeMealPlanGroceryListItemsList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListMealPlanGroceryListItemsByMealPlan), testutils.ContextMatcher, exampleMealPlanID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanGroceryListItemsForMealPlan(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemsForMealPlanRequest{MealPlanId: exampleMealPlanID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealPlanOption()
		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleMealPlanEventID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanOption), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanOption(ctx, &mealplanninggrpc.GetMealPlanOptionRequest{
			MealPlanId:       exampleMealPlanID,
			MealPlanEventId:  exampleMealPlanEventID,
			MealPlanOptionId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealPlanOptionVote()
		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleMealPlanEventID := mealplanningfakes.BuildFakeID()
		exampleMealPlanOptionID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanOptionVote), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanOptionVote(ctx, &mealplanninggrpc.GetMealPlanOptionVoteRequest{
			MealPlanId:           exampleMealPlanID,
			MealPlanEventId:      exampleMealPlanEventID,
			MealPlanOptionId:     exampleMealPlanOptionID,
			MealPlanOptionVoteId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanOptionVotes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleMealPlanEventID := mealplanningfakes.BuildFakeID()
		exampleMealPlanOptionID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeMealPlanOptionVotesList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListMealPlanOptionVotes), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanOptionVotes(ctx, &mealplanninggrpc.GetMealPlanOptionVotesRequest{
			MealPlanId:       exampleMealPlanID,
			MealPlanEventId:  exampleMealPlanEventID,
			MealPlanOptionId: exampleMealPlanOptionID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanOptions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleMealPlanEventID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeMealPlanOptionsList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListMealPlanOptions), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanOptions(ctx, &mealplanninggrpc.GetMealPlanOptionsRequest{
			MealPlanId:      exampleMealPlanID,
			MealPlanEventId: exampleMealPlanEventID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealPlanTask()
		exampleMealPlanID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanTask), testutils.ContextMatcher, exampleMealPlanID, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanTask(ctx, &mealplanninggrpc.GetMealPlanTaskRequest{
			MealPlanId:     exampleMealPlanID,
			MealPlanTaskId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMealPlanTasks(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeMealPlanTasksList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListMealPlanTasksByMealPlan), testutils.ContextMatcher, exampleMealPlanID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanTasks(ctx, &mealplanninggrpc.GetMealPlanTasksRequest{MealPlanId: exampleMealPlanID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetMeals(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealsList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListMeals), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.GetMeals(ctx, &mealplanninggrpc.GetMealsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetUserIngredientPreference(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeUserIngredientPreference()
		exampleUserID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadUserIngredientPreference), testutils.ContextMatcher, exampleUserID, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		result, err := s.GetUserIngredientPreference(ctx, &mealplanninggrpc.GetUserIngredientPreferenceRequest{
			UserIngredientPreferenceId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetUserIngredientPreferences(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeUserIngredientPreferencesList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListUserIngredientPreferences), testutils.ContextMatcher, exampleUserID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		result, err := s.GetUserIngredientPreferences(ctx, &mealplanninggrpc.GetUserIngredientPreferencesRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_SearchForMeals(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForMealsRequest](t)

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.SearchMeals), testutils.ContextMatcher, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		result, err := s.SearchForMeals(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateMealPlanRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeMealPlan()
		exampleAccountID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.UpdateMealPlan), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleAccountID, testutils.MatchType[*mealplanning.MealPlanUpdateRequestInput]()).Return(nil)
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlan), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleAccountID).Return(exampleResponse, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		res, err := s.UpdateMealPlan(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateMealPlanEventRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeMealPlanEvent()

		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.UpdateMealPlanEvent), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleRequest.MealPlanEventId, testutils.MatchType[*mealplanning.MealPlanEventUpdateRequestInput]()).Return(nil)
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanEvent), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleRequest.MealPlanEventId).Return(exampleResponse, nil)
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanEvent(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateMealPlanGroceryListItemRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeMealPlanGroceryListItem()

		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.UpdateMealPlanGroceryListItem), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleRequest.MealPlanGroceryListItemId, testutils.MatchType[*mealplanning.MealPlanGroceryListItemUpdateRequestInput]()).Return(nil)
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanGroceryListItem), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleRequest.MealPlanGroceryListItemId).Return(exampleResponse, nil)
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanGroceryListItem(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateMealPlanOptionRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeMealPlanOption()

		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.UpdateMealPlanOption), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleRequest.MealPlanEventId, exampleRequest.MealPlanOptionId, testutils.MatchType[*mealplanning.MealPlanOptionUpdateRequestInput]()).Return(nil)
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanOption), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleRequest.MealPlanEventId, exampleRequest.MealPlanOptionId).Return(exampleResponse, nil)
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanOption(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateMealPlanOptionVoteRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeMealPlanOptionVote()

		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.UpdateMealPlanOptionVote), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleRequest.MealPlanEventId, exampleRequest.MealPlanOptionId, exampleRequest.MealPlanOptionVoteId, testutils.MatchType[*mealplanning.MealPlanOptionVoteUpdateRequestInput]()).Return(nil)
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanOptionVote), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleRequest.MealPlanEventId, exampleRequest.MealPlanOptionId, exampleRequest.MealPlanOptionVoteId).Return(exampleResponse, nil)
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanOptionVote(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateMealPlanTaskStatus(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateMealPlanTaskStatusRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeMealPlanTask()

		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.MealPlanTaskStatusChange), testutils.ContextMatcher, testutils.MatchType[*mealplanning.MealPlanTaskStatusChangeRequestInput]()).Return(nil)
		mmpm.On(reflection.GetMethodName(mmpm.ReadMealPlanTask), testutils.ContextMatcher, exampleRequest.MealPlanId, exampleRequest.MealPlanTaskId).Return(exampleResponse, nil)
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanTaskStatus(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateUserIngredientPreference(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateUserIngredientPreferenceRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeUserIngredientPreference()
		exampleUserID := mealplanningfakes.BuildFakeID()

		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.UpdateUserIngredientPreference), testutils.ContextMatcher, exampleRequest.UserIngredientPreferenceId, exampleUserID, testutils.MatchType[*mealplanning.UserIngredientPreferenceUpdateRequestInput]()).Return(nil)
		mmpm.On(reflection.GetMethodName(mmpm.ReadUserIngredientPreference), testutils.ContextMatcher, exampleUserID, exampleRequest.UserIngredientPreferenceId).Return(exampleResponse, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific user ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: exampleUserID,
				},
			}, nil
		}

		res, err := s.UpdateUserIngredientPreference(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_CreateAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleCreatedAccountInstrumentOwnership := mealplanningfakes.BuildFakeAccountInstrumentOwnership()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.CreateAccountInstrumentOwnership), testutils.ContextMatcher, exampleAccountID, testutils.MatchType[*mealplanning.AccountInstrumentOwnershipCreationRequestInput]()).Return(exampleCreatedAccountInstrumentOwnership, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateAccountInstrumentOwnershipRequest](t)

		actual, err := s.CreateAccountInstrumentOwnership(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedAccountInstrumentOwnership.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeAccountInstrumentOwnership()
		exampleAccountID := mealplanningfakes.BuildFakeID()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadAccountInstrumentOwnership), testutils.ContextMatcher, exampleAccountID, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		result, err := s.GetAccountInstrumentOwnership(ctx, &mealplanninggrpc.GetAccountInstrumentOwnershipRequest{
			AccountInstrumentOwnershipId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_GetAccountInstrumentOwnerships(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeAccountInstrumentOwnershipsList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ListAccountInstrumentOwnerships), testutils.ContextMatcher, exampleAccountID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		result, err := s.GetAccountInstrumentOwnerships(ctx, &mealplanninggrpc.GetAccountInstrumentOwnershipsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_SearchForValidInstrumentsNotOwnedByAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidInstrumentsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForValidInstrumentsNotOwnedByAccountRequest](t)

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.SearchValidInstrumentsNotOwnedByAccount), testutils.ContextMatcher, exampleAccountID, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		result, err := s.SearchForValidInstrumentsNotOwnedByAccount(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_UpdateAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateAccountInstrumentOwnershipRequest](t)
		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleAccountInstrumentOwnership := mealplanningfakes.BuildFakeAccountInstrumentOwnership()

		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ReadAccountInstrumentOwnership), testutils.ContextMatcher, exampleAccountID, exampleRequest.AccountInstrumentOwnershipId).Return(exampleAccountInstrumentOwnership, nil)
		mmpm.On(reflection.GetMethodName(mmpm.UpdateAccountInstrumentOwnership), testutils.ContextMatcher, exampleAccountInstrumentOwnership.ID, exampleAccountInstrumentOwnership.BelongsToAccount, testutils.MatchType[*mealplanning.AccountInstrumentOwnershipUpdateRequestInput]()).Return(nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		res, err := s.UpdateAccountInstrumentOwnership(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}

func TestServiceImpl_ArchiveAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		exampleAccountID := mealplanningfakes.BuildFakeID()
		exampleAccountInstrumentOwnershipID := mealplanningfakes.BuildFakeID()

		mmpm := &mockmanagers.MockMealPlanningManager{}
		mmpm.On(reflection.GetMethodName(mmpm.ArchiveAccountInstrumentOwnership), testutils.ContextMatcher, exampleAccountID, exampleAccountInstrumentOwnershipID).Return(nil)
		s.mealPlanningManager = mmpm

		// Override session context to return specific account ID
		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: exampleAccountID,
			}, nil
		}

		res, err := s.ArchiveAccountInstrumentOwnership(ctx, &mealplanninggrpc.ArchiveAccountInstrumentOwnershipRequest{
			AccountInstrumentOwnershipId: exampleAccountInstrumentOwnershipID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmpm)
	})
}
