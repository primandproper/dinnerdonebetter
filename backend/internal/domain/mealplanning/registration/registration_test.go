package registration

import (
	"context"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningmgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	mockmanagers "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc"

	platformcomments "github.com/primandproper/platform-go/v15/comments"
	jobscfg "github.com/primandproper/primitives-go/v2/jobs/config"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCPermissions(T *testing.T) {
	T.Parallel()

	T.Run("is the surface's table", func(t *testing.T) {
		t.Parallel()

		table := GRPCPermissions()
		require.NotEmpty(t, table)
		assert.Len(t, table, len(mealplanningsvc.ProvideMethodPermissions()))
	})
}

func TestScheduledJobs(T *testing.T) {
	T.Parallel()

	T.Run("renders both jobs from the domain's block of the config", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, &config.ScheduledJobsConfig{
			MealPlanning: config.MealPlanningScheduledJobsConfig{
				MealPlanFinalizationStarter: jobscfg.JobConfig{Interval: time.Minute},
				MealPlanTaskNotifications:   jobscfg.JobConfig{Interval: time.Minute},
			},
		})

		jobs, err := ScheduledJobs(i)
		require.NoError(t, err)

		names := make([]string, 0, len(jobs))
		for _, job := range jobs {
			names = append(names, job.Name)
		}

		assert.ElementsMatch(t, []string{jobMealPlanFinalizationStarter, jobMealPlanTaskNotifications}, names)
	})

	T.Run("leaves a disabled job out", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, &config.ScheduledJobsConfig{
			MealPlanning: config.MealPlanningScheduledJobsConfig{
				MealPlanFinalizationStarter: jobscfg.JobConfig{Interval: time.Minute, Disabled: true},
				MealPlanTaskNotifications:   jobscfg.JobConfig{Interval: time.Minute},
			},
		})

		jobs, err := ScheduledJobs(i)
		require.NoError(t, err)
		require.Len(t, jobs, 1)
		assert.Equal(t, jobMealPlanTaskNotifications, jobs[0].Name)
	})

	T.Run("refuses a job it cannot render", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, &config.ScheduledJobsConfig{
			MealPlanning: config.MealPlanningScheduledJobsConfig{
				MealPlanFinalizationStarter: jobscfg.JobConfig{Schedule: "not a cron spec"},
				MealPlanTaskNotifications:   jobscfg.JobConfig{Interval: time.Minute},
			},
		})

		_, err := ScheduledJobs(i)
		assert.Error(t, err)
	})
}

func TestCommentTargets(T *testing.T) {
	T.Parallel()

	T.Run("names the three types and checks nothing", func(t *testing.T) {
		t.Parallel()

		targets := CommentTargets()

		for _, targetType := range []platformcomments.TargetType{
			mealplanning.CommentTargetTypeRecipes,
			mealplanning.CommentTargetTypeMeals,
			mealplanning.CommentTargetTypeMealPlans,
		} {
			require.True(t, targets.Known(targetType))
			assert.Nil(t, targets[targetType].Exists, "%s carries a check in the hookless catalog", targetType)
		}

		assert.Len(t, targets, 3)
	})
}

func TestCheckedCommentTargets(T *testing.T) {
	T.Parallel()

	T.Run("checks recipes and meals through the manager and leaves meal plans unchecked", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager := &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(context.Context, string) (*mealplanning.Recipe, error) {
				return &mealplanning.Recipe{}, nil
			},
			ReadMealFunc: func(context.Context, string) (*mealplanning.Meal, error) {
				return &mealplanning.Meal{}, nil
			},
		}

		i := do.New()
		do.ProvideValue[mealplanningmgr.MealPlanningManager](i, manager)

		targets := CheckedCommentTargets(i)
		assert.Len(t, targets, len(CommentTargets()))

		for _, targetType := range []platformcomments.TargetType{
			mealplanning.CommentTargetTypeRecipes,
			mealplanning.CommentTargetTypeMeals,
		} {
			require.NotNil(t, targets[targetType].Exists, "%s is unchecked", targetType)

			exists, err := targets[targetType].Exists(ctx, tenancy.Global(), t.Name())
			require.NoError(t, err)
			assert.True(t, exists)
		}

		assert.Nil(t, targets[mealplanning.CommentTargetTypeMealPlans].Exists)
		assert.Len(t, manager.ReadRecipeCalls(), 1)
		assert.Len(t, manager.ReadMealCalls(), 1)
	})
}
