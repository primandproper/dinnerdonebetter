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
		// The saga starter is nil for most tests: only the three admin RPCs reach it.
		mealPlanFinalizationStarter: nil,
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ArchiveMealFunc: func(_ context.Context, mealID string, ownerID string) error {
				assert.Equal(t, exampleMealID, mealID)
				assert.Equal(t, exampleUserID, ownerID)

				return nil
			},
		}
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

		assert.Len(t, mmpm.ArchiveMealCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ArchiveMealPlanFunc: func(_ context.Context, mealPlanID string, ownerID string) error {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleAccountID, ownerID)

				return nil
			},
		}
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

		assert.Len(t, mmpm.ArchiveMealPlanCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ArchiveMealPlanEventFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) error {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealPlanEvent(ctx, &mealplanninggrpc.ArchiveMealPlanEventRequest{
			MealPlanId:      exampleMealPlanID,
			MealPlanEventId: exampleMealPlanEventID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ArchiveMealPlanEventCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ArchiveMealPlanGroceryListItemFunc: func(_ context.Context, mealPlanID string, mealPlanGroceryListItemID string) error {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanGroceryListItemID, mealPlanGroceryListItemID)

				return nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealPlanGroceryListItem(ctx, &mealplanninggrpc.ArchiveMealPlanGroceryListItemRequest{
			MealPlanId:                exampleMealPlanID,
			MealPlanGroceryListItemId: exampleMealPlanGroceryListItemID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ArchiveMealPlanGroceryListItemCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ArchiveMealPlanOptionFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string) error {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)

				return nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealPlanOption(ctx, &mealplanninggrpc.ArchiveMealPlanOptionRequest{
			MealPlanId:       exampleMealPlanID,
			MealPlanEventId:  exampleMealPlanEventID,
			MealPlanOptionId: exampleMealPlanOptionID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ArchiveMealPlanOptionCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ArchiveMealPlanOptionVoteFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, mealPlanOptionVoteID string) error {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)
				assert.Equal(t, exampleMealPlanOptionVoteID, mealPlanOptionVoteID)

				return nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealPlanOptionVote(ctx, &mealplanninggrpc.ArchiveMealPlanOptionVoteRequest{
			MealPlanId:           exampleMealPlanID,
			MealPlanEventId:      exampleMealPlanEventID,
			MealPlanOptionId:     exampleMealPlanOptionID,
			MealPlanOptionVoteId: exampleMealPlanOptionVoteID,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ArchiveMealPlanOptionVoteCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ArchiveUserIngredientPreferenceFunc: func(_ context.Context, ownerID string, ingredientPreferenceID string) error {
				assert.Equal(t, exampleUserID, ownerID)
				assert.Equal(t, exampleUserIngredientPreferenceID, ingredientPreferenceID)

				return nil
			},
		}
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

		assert.Len(t, mmpm.ArchiveUserIngredientPreferenceCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ListMealListsFunc: func(_ context.Context, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealList], error) {
				return expected, nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.GetMealLists(ctx, &mealplanninggrpc.GetMealListsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Results, 1)

		assert.Len(t, mmpm.ListMealListsCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			CreateMealListFunc: func(_ context.Context, actualUserID string, _ *mealplanning.MealListCreationRequestInput) (*mealplanning.MealList, error) {
				assert.Equal(t, userID, actualUserID)

				return created, nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{Input: input})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, created.ID, res.Created.Id)

		assert.Len(t, mmpm.CreateMealListCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			UpdateMealListFunc: func(_ context.Context, mealListID string, actualUserID string, _ *mealplanning.MealListUpdateRequestInput) error {
				assert.Equal(t, listID, mealListID)
				assert.Equal(t, userID, actualUserID)

				return nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealList(ctx, &mealplanninggrpc.UpdateMealListRequest{
			MealListId: listID,
			Input:      input,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mmpm.UpdateMealListCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ArchiveMealListFunc: func(_ context.Context, mealListID string, actualUserID string) error {
				assert.Equal(t, listID, mealListID)
				assert.Equal(t, userID, actualUserID)

				return nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealList(ctx, &mealplanninggrpc.ArchiveMealListRequest{MealListId: listID})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mmpm.ArchiveMealListCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ListMealListItemsFunc: func(_ context.Context, mealListID string, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealListItem], error) {
				assert.Equal(t, listID, mealListID)

				return expected, nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.GetMealListItems(ctx, &mealplanninggrpc.GetMealListItemsRequest{MealListId: listID})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Results, 1)

		assert.Len(t, mmpm.ListMealListItemsCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			AddMealToMealListFunc: func(_ context.Context, mealListID string, actualMealID string, notes string) (*mealplanning.MealListItem, error) {
				assert.Equal(t, listID, mealListID)
				assert.Equal(t, mealID, actualMealID)
				assert.Equal(t, input.Notes, notes)

				return created, nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.CreateMealListItem(ctx, &mealplanninggrpc.CreateMealListItemRequest{Input: input})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, created.ID, res.Created.Id)

		assert.Len(t, mmpm.AddMealToMealListCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			UpdateMealListItemFunc: func(_ context.Context, mealListItemID string, mealListID string, actualMealID string, _ *mealplanning.MealListItemUpdateRequestInput) error {
				assert.Equal(t, itemID, mealListItemID)
				assert.Equal(t, listID, mealListID)
				assert.Equal(t, mealID, actualMealID)

				return nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealListItem(ctx, &mealplanninggrpc.UpdateMealListItemRequest{
			MealListItemId: itemID,
			Input:          input,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mmpm.UpdateMealListItemCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			RemoveMealFromMealListFunc: func(_ context.Context, mealListID string, mealListItemID string) error {
				assert.Equal(t, listID, mealListID)
				assert.Equal(t, itemID, mealListItemID)

				return nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.ArchiveMealListItem(ctx, &mealplanninggrpc.ArchiveMealListItemRequest{
			MealListItemId: itemID,
			MealListId:     listID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mmpm.RemoveMealFromMealListCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			CreateMealFunc: func(_ context.Context, creatorID string, _ *mealplanning.MealCreationRequestInput) (*mealplanning.Meal, error) {
				assert.Equal(t, exampleUserID, creatorID)

				return exampleCreatedMeal, nil
			},
		}
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

		assert.Len(t, mmpm.CreateMealCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			CreateMealPlanFunc: func(_ context.Context, ownerID string, creatorID string, _ *mealplanning.MealPlanCreationRequestInput) (*mealplanning.MealPlan, error) {
				assert.Equal(t, exampleAccountID, ownerID)
				assert.Equal(t, exampleUserID, creatorID)

				return exampleCreatedMealPlan, nil
			},
		}
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

		assert.Len(t, mmpm.CreateMealPlanCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			CreateMealPlanEventFunc: func(_ context.Context, mealPlanID string, _ *mealplanning.MealPlanEventCreationRequestInput) (*mealplanning.MealPlanEvent, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return exampleCreatedMealPlanEvent, nil
			},
		}
		s.mealPlanningManager = mmpm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateMealPlanEventRequest](t)
		exampleInput.MealPlanId = exampleMealPlanID

		actual, err := s.CreateMealPlanEvent(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedMealPlanEvent.ID, actual.Created.Id)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.CreateMealPlanEventCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			CreateMealPlanOptionWithEventIDFunc: func(_ context.Context, mealPlanEventID string, _ *mealplanning.MealPlanOptionCreationRequestInput) (*mealplanning.MealPlanOption, error) {
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return exampleCreatedMealPlanOption, nil
			},
		}
		s.mealPlanningManager = mmpm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateMealPlanOptionRequest](t)
		exampleInput.MealPlanId = exampleMealPlanID
		exampleInput.MealPlanEventId = exampleMealPlanEventID

		actual, err := s.CreateMealPlanOption(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedMealPlanOption.ID, actual.Created.Id)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.CreateMealPlanOptionWithEventIDCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			CreateMealPlanOptionVotesFunc: func(_ context.Context, mealPlanID string, _ string, creatorID string, _ *mealplanning.MealPlanOptionVoteCreationRequestInput) ([]*mealplanning.MealPlanOptionVote, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleUserID, creatorID)

				return exampleCreatedMealPlanOptionVotes, nil
			},
		}
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

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.CreateMealPlanOptionVotesCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			CreateMealPlanTaskFunc: func(_ context.Context, _ *mealplanning.MealPlanTaskCreationRequestInput) (*mealplanning.MealPlanTask, error) {
				return exampleCreatedMealPlanTask, nil
			},
		}
		s.mealPlanningManager = mmpm

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateMealPlanTaskRequest](t)
		exampleInput.MealPlanId = exampleMealPlanID

		actual, err := s.CreateMealPlanTask(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleCreatedMealPlanTask.ID, actual.Created.Id)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.CreateMealPlanTaskCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			CreateUserIngredientPreferenceFunc: func(_ context.Context, ownerID string, _ *mealplanning.UserIngredientPreferenceCreationRequestInput) ([]*mealplanning.UserIngredientPreference, error) {
				assert.Equal(t, exampleUserID, ownerID)

				return exampleCreatedUserIngredientPreferences, nil
			},
		}
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

		assert.Len(t, mmpm.CreateUserIngredientPreferenceCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			FinalizeMealPlanFunc: func(_ context.Context, mealPlanID string, ownerID string) (bool, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleAccountID, ownerID)

				return exampleFinalized, nil
			},
		}
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

		assert.Len(t, mmpm.FinalizeMealPlanCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealFunc: func(_ context.Context, mealID string) (*mealplanning.Meal, error) {
				assert.Equal(t, exampleMeal.ID, mealID)

				return exampleMeal, nil
			},
			MealMermaidFunc: func(_ context.Context, meal *mealplanning.Meal) (string, error) {
				assert.Equal(t, exampleMeal, meal)

				return exampleMermaidDiagram, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMermaidDiagramForMeal(ctx, &mealplanninggrpc.GetMermaidDiagramForMealRequest{MealId: exampleMeal.ID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, exampleMermaidDiagram, result.Response)

		assert.Len(t, mmpm.ReadMealCalls(), 1)
		assert.Len(t, mmpm.MealMermaidCalls(), 1)
	})
}

func TestServiceImpl_GetMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMeal()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealFunc: func(_ context.Context, mealID string) (*mealplanning.Meal, error) {
				assert.Equal(t, exampleResult.ID, mealID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMeal(ctx, &mealplanninggrpc.GetMealRequest{MealId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, mealPlanID string, ownerID string) (*mealplanning.MealPlan, error) {
				assert.Equal(t, exampleResult.ID, mealPlanID)
				assert.Equal(t, exampleAccountID, ownerID)

				return exampleResult, nil
			},
		}
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

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ListMealPlansFunc: func(_ context.Context, ownerID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlan], error) {
				assert.Equal(t, exampleAccountID, ownerID)

				return exampleResult, nil
			},
		}
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

		assert.Len(t, mmpm.ListMealPlansCalls(), 1)
	})
}

func TestServiceImpl_GetMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealPlanEvent()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ReadMealPlanEventFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) (*mealplanning.MealPlanEvent, error) {
				assert.Equal(t, exampleResult.BelongsToMealPlan, mealPlanID)
				assert.Equal(t, exampleResult.ID, mealPlanEventID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanEvent(ctx, &mealplanninggrpc.GetMealPlanEventRequest{
			MealPlanId:      exampleResult.BelongsToMealPlan,
			MealPlanEventId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanEventCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ListMealPlanEventsFunc: func(_ context.Context, mealPlanID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanEvent], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanEvents(ctx, &mealplanninggrpc.GetMealPlanEventsRequest{MealPlanId: exampleMealPlanID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ListMealPlanEventsCalls(), 1)
	})
}

func TestServiceImpl_GetMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealPlanGroceryListItem()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ReadMealPlanGroceryListItemFunc: func(_ context.Context, mealPlanID string, mealPlanGroceryListItemID string) (*mealplanning.MealPlanGroceryListItem, error) {
				assert.Equal(t, exampleResult.BelongsToMealPlan, mealPlanID)
				assert.Equal(t, exampleResult.ID, mealPlanGroceryListItemID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanGroceryListItem(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemRequest{
			MealPlanId:                exampleResult.BelongsToMealPlan,
			MealPlanGroceryListItemId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanGroceryListItemCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ListMealPlanGroceryListItemsByMealPlanFunc: func(_ context.Context, mealPlanID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanGroceryListItem], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanGroceryListItemsForMealPlan(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemsForMealPlanRequest{MealPlanId: exampleMealPlanID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ListMealPlanGroceryListItemsByMealPlanCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ReadMealPlanOptionFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string) (*mealplanning.MealPlanOption, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleResult.ID, mealPlanOptionID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanOption(ctx, &mealplanninggrpc.GetMealPlanOptionRequest{
			MealPlanId:       exampleMealPlanID,
			MealPlanEventId:  exampleMealPlanEventID,
			MealPlanOptionId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanOptionCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ReadMealPlanOptionVoteFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, mealPlanOptionVoteID string) (*mealplanning.MealPlanOptionVote, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)
				assert.Equal(t, exampleResult.ID, mealPlanOptionVoteID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanOptionVote(ctx, &mealplanninggrpc.GetMealPlanOptionVoteRequest{
			MealPlanId:           exampleMealPlanID,
			MealPlanEventId:      exampleMealPlanEventID,
			MealPlanOptionId:     exampleMealPlanOptionID,
			MealPlanOptionVoteId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanOptionVoteCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ListMealPlanOptionVotesFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanOptionVote], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanOptionVotes(ctx, &mealplanninggrpc.GetMealPlanOptionVotesRequest{
			MealPlanId:       exampleMealPlanID,
			MealPlanEventId:  exampleMealPlanEventID,
			MealPlanOptionId: exampleMealPlanOptionID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ListMealPlanOptionVotesCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ListMealPlanOptionsFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanOption], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanOptions(ctx, &mealplanninggrpc.GetMealPlanOptionsRequest{
			MealPlanId:      exampleMealPlanID,
			MealPlanEventId: exampleMealPlanEventID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ListMealPlanOptionsCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ReadMealPlanTaskFunc: func(_ context.Context, mealPlanID string, mealPlanTaskID string) (*mealplanning.MealPlanTask, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleResult.ID, mealPlanTaskID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanTask(ctx, &mealplanninggrpc.GetMealPlanTaskRequest{
			MealPlanId:     exampleMealPlanID,
			MealPlanTaskId: exampleResult.ID,
		})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanTaskCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			ListMealPlanTasksByMealPlanFunc: func(_ context.Context, mealPlanID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanTask], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMealPlanTasks(ctx, &mealplanninggrpc.GetMealPlanTasksRequest{MealPlanId: exampleMealPlanID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.ListMealPlanTasksByMealPlanCalls(), 1)
	})
}

func TestServiceImpl_GetMeals(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeMealsList()

		ctx := t.Context()
		s := buildServiceImplForMealPlanningTest(t)

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ListMealsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.GetMeals(ctx, &mealplanninggrpc.GetMealsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mmpm.ListMealsCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadUserIngredientPreferenceFunc: func(_ context.Context, ownerID string, ingredientPreferenceID string) (*mealplanning.UserIngredientPreference, error) {
				assert.Equal(t, exampleUserID, ownerID)
				assert.Equal(t, exampleResult.ID, ingredientPreferenceID)

				return exampleResult, nil
			},
		}
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

		assert.Len(t, mmpm.ReadUserIngredientPreferenceCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ListUserIngredientPreferencesFunc: func(_ context.Context, ownerID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference], error) {
				assert.Equal(t, exampleUserID, ownerID)

				return exampleResult, nil
			},
		}
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

		assert.Len(t, mmpm.ListUserIngredientPreferencesCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			SearchMealsFunc: func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mmpm

		result, err := s.SearchForMeals(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mmpm.SearchMealsCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			UpdateMealPlanFunc: func(_ context.Context, mealPlanID string, ownerID string, _ *mealplanning.MealPlanUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleAccountID, ownerID)

				return nil
			},
			ReadMealPlanFunc: func(_ context.Context, mealPlanID string, ownerID string) (*mealplanning.MealPlan, error) {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleAccountID, ownerID)

				return exampleResponse, nil
			},
		}
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

		assert.Len(t, mmpm.UpdateMealPlanCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			UpdateMealPlanEventFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, _ *mealplanning.MealPlanEventUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleRequest.MealPlanEventId, mealPlanEventID)

				return nil
			},
			ReadMealPlanEventFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) (*mealplanning.MealPlanEvent, error) {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleRequest.MealPlanEventId, mealPlanEventID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanEvent(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.UpdateMealPlanEventCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanEventCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			UpdateMealPlanGroceryListItemFunc: func(_ context.Context, mealPlanID string, mealPlanGroceryListItemID string, _ *mealplanning.MealPlanGroceryListItemUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleRequest.MealPlanGroceryListItemId, mealPlanGroceryListItemID)

				return nil
			},
			ReadMealPlanGroceryListItemFunc: func(_ context.Context, mealPlanID string, mealPlanGroceryListItemID string) (*mealplanning.MealPlanGroceryListItem, error) {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleRequest.MealPlanGroceryListItemId, mealPlanGroceryListItemID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanGroceryListItem(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.UpdateMealPlanGroceryListItemCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanGroceryListItemCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			UpdateMealPlanOptionFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, _ *mealplanning.MealPlanOptionUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleRequest.MealPlanEventId, mealPlanEventID)
				assert.Equal(t, exampleRequest.MealPlanOptionId, mealPlanOptionID)

				return nil
			},
			ReadMealPlanOptionFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string) (*mealplanning.MealPlanOption, error) {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleRequest.MealPlanEventId, mealPlanEventID)
				assert.Equal(t, exampleRequest.MealPlanOptionId, mealPlanOptionID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanOption(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.UpdateMealPlanOptionCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanOptionCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			UpdateMealPlanOptionVoteFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, mealPlanOptionVoteID string, _ *mealplanning.MealPlanOptionVoteUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleRequest.MealPlanEventId, mealPlanEventID)
				assert.Equal(t, exampleRequest.MealPlanOptionId, mealPlanOptionID)
				assert.Equal(t, exampleRequest.MealPlanOptionVoteId, mealPlanOptionVoteID)

				return nil
			},
			ReadMealPlanOptionVoteFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, mealPlanOptionVoteID string) (*mealplanning.MealPlanOptionVote, error) {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleRequest.MealPlanEventId, mealPlanEventID)
				assert.Equal(t, exampleRequest.MealPlanOptionId, mealPlanOptionID)
				assert.Equal(t, exampleRequest.MealPlanOptionVoteId, mealPlanOptionVoteID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanOptionVote(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.UpdateMealPlanOptionVoteCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanOptionVoteCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadMealPlanFunc: func(_ context.Context, _ string, _ string) (*mealplanning.MealPlan, error) {
				return &mealplanning.MealPlan{}, nil
			},
			MealPlanTaskStatusChangeFunc: func(_ context.Context, _ *mealplanning.MealPlanTaskStatusChangeRequestInput) error {
				return nil
			},
			ReadMealPlanTaskFunc: func(_ context.Context, mealPlanID string, mealPlanTaskID string) (*mealplanning.MealPlanTask, error) {
				assert.Equal(t, exampleRequest.MealPlanId, mealPlanID)
				assert.Equal(t, exampleRequest.MealPlanTaskId, mealPlanTaskID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mmpm

		res, err := s.UpdateMealPlanTaskStatus(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Updated.Id)

		assert.Len(t, mmpm.ReadMealPlanCalls(), 1)
		assert.Len(t, mmpm.MealPlanTaskStatusChangeCalls(), 1)
		assert.Len(t, mmpm.ReadMealPlanTaskCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			UpdateUserIngredientPreferenceFunc: func(_ context.Context, ingredientPreferenceID string, ownerID string, _ *mealplanning.UserIngredientPreferenceUpdateRequestInput) error {
				assert.Equal(t, exampleRequest.UserIngredientPreferenceId, ingredientPreferenceID)
				assert.Equal(t, exampleUserID, ownerID)

				return nil
			},
			ReadUserIngredientPreferenceFunc: func(_ context.Context, ownerID string, ingredientPreferenceID string) (*mealplanning.UserIngredientPreference, error) {
				assert.Equal(t, exampleUserID, ownerID)
				assert.Equal(t, exampleRequest.UserIngredientPreferenceId, ingredientPreferenceID)

				return exampleResponse, nil
			},
		}
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

		assert.Len(t, mmpm.UpdateUserIngredientPreferenceCalls(), 1)
		assert.Len(t, mmpm.ReadUserIngredientPreferenceCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			CreateAccountInstrumentOwnershipFunc: func(_ context.Context, ownerID string, _ *mealplanning.AccountInstrumentOwnershipCreationRequestInput) (*mealplanning.AccountInstrumentOwnership, error) {
				assert.Equal(t, exampleAccountID, ownerID)

				return exampleCreatedAccountInstrumentOwnership, nil
			},
		}
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

		assert.Len(t, mmpm.CreateAccountInstrumentOwnershipCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadAccountInstrumentOwnershipFunc: func(_ context.Context, ownerID string, instrumentOwnershipID string) (*mealplanning.AccountInstrumentOwnership, error) {
				assert.Equal(t, exampleAccountID, ownerID)
				assert.Equal(t, exampleResult.ID, instrumentOwnershipID)

				return exampleResult, nil
			},
		}
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

		assert.Len(t, mmpm.ReadAccountInstrumentOwnershipCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ListAccountInstrumentOwnershipsFunc: func(_ context.Context, ownerID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.AccountInstrumentOwnership], error) {
				assert.Equal(t, exampleAccountID, ownerID)

				return exampleResult, nil
			},
		}
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

		assert.Len(t, mmpm.ListAccountInstrumentOwnershipsCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			SearchValidInstrumentsNotOwnedByAccountFunc: func(_ context.Context, accountID string, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
				assert.Equal(t, exampleAccountID, accountID)
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
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

		assert.Len(t, mmpm.SearchValidInstrumentsNotOwnedByAccountCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ReadAccountInstrumentOwnershipFunc: func(_ context.Context, ownerID string, instrumentOwnershipID string) (*mealplanning.AccountInstrumentOwnership, error) {
				assert.Equal(t, exampleAccountID, ownerID)
				assert.Equal(t, exampleRequest.AccountInstrumentOwnershipId, instrumentOwnershipID)

				return exampleAccountInstrumentOwnership, nil
			},
			UpdateAccountInstrumentOwnershipFunc: func(_ context.Context, instrumentOwnershipID string, ownerID string, _ *mealplanning.AccountInstrumentOwnershipUpdateRequestInput) error {
				assert.Equal(t, exampleAccountInstrumentOwnership.ID, instrumentOwnershipID)
				assert.Equal(t, exampleAccountInstrumentOwnership.BelongsToAccount, ownerID)

				return nil
			},
		}
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

		assert.Len(t, mmpm.ReadAccountInstrumentOwnershipCalls(), 1)
		assert.Len(t, mmpm.UpdateAccountInstrumentOwnershipCalls(), 1)
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

		mmpm := &mockmanagers.MealPlanningManagerMock{
			ArchiveAccountInstrumentOwnershipFunc: func(_ context.Context, ownerID string, instrumentOwnershipID string) error {
				assert.Equal(t, exampleAccountID, ownerID)
				assert.Equal(t, exampleAccountInstrumentOwnershipID, instrumentOwnershipID)

				return nil
			},
		}
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

		assert.Len(t, mmpm.ArchiveAccountInstrumentOwnershipCalls(), 1)
	})
}
