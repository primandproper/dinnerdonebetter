package grpc

import (
	"reflect"
	"sync"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	mockmanagers "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/filtering/filteringpb"
	filteringgrpc "github.com/primandproper/primitives-go/v2/filtering/grpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// archiveGrantForListRPC is the archive decision of every list RPC on the service: the archive
// grant that lets a caller see the page's archived rows, or nothing for a surface that never
// honors include_archived. See archive.go for why each surface lands where it does.
//
// GetMealListItems and GetRecipeListItems are absent because the service declares no rpc for
// either: their handlers decide too, but nothing on the wire reaches them.
var archiveGrantForListRPC = map[string]authorization.Permission{
	// Reads across everybody's rows, whose archive grant reaches only the caller's own.
	"GetRecipes":                              "",
	"SearchForRecipes":                        "",
	"SearchForMealEligibleRecipes":            "",
	"SearchForRecipesWithInstrumentOwnership": "",
	"GetRecipeSteps":                          "",
	"GetRecipePrepTasks":                      "",
	"GetRecipeRatingsForRecipe":               "",
	"GetRecipeStepCompletionConditions":       "",
	"GetRecipeStepIngredients":                "",
	"GetRecipeStepInstruments":                "",
	"GetRecipeStepProducts":                   "",
	"GetRecipeStepVessels":                    "",
	"GetRecipeLists":                          "",
	"GetMeals":                                "",
	"SearchForMeals":                          "",
	// No archive to hide.
	"GetMealPlanTasks": "",

	// Reads confined to the caller's own account or rows.
	"GetMealLists":                                       authorization.ArchiveMealListsPermission,
	"GetMealPlansForAccount":                             authorization.ArchiveMealPlansPermission,
	"GetMealPlanEvents":                                  authorization.ArchiveMealPlanEventsPermission,
	"GetMealPlanOptions":                                 authorization.ArchiveMealPlanOptionsPermission,
	"GetMealPlanOptionVotes":                             authorization.ArchiveMealPlanOptionVotesPermission,
	"GetMealPlanGroceryListItemsForMealPlan":             authorization.ArchiveMealPlanGroceryListItemsPermission,
	"GetMealPlanRecipeOptionSelectionsForMealPlanOption": authorization.ArchiveMealPlanRecipeOptionSelectionsPermission,
	"GetUserIngredientPreferences":                       authorization.ArchiveUserIngredientPreferencesPermission,
	"GetAccountInstrumentOwnerships":                     authorization.ArchiveAccountInstrumentOwnershipsPermission,

	// The valid-* catalogs, whose archive grant is an operator's over every row.
	"GetValidIngredients":                                 authorization.ArchiveValidIngredientsPermission,
	"SearchForValidIngredients":                           authorization.ArchiveValidIngredientsPermission,
	"SearchValidIngredientsByPreparation":                 authorization.ArchiveValidIngredientsPermission,
	"GetValidIngredientGroups":                            authorization.ArchiveValidIngredientGroupsPermission,
	"SearchForValidIngredientGroups":                      authorization.ArchiveValidIngredientGroupsPermission,
	"GetValidIngredientStates":                            authorization.ArchiveValidIngredientStatesPermission,
	"SearchForValidIngredientStates":                      authorization.ArchiveValidIngredientStatesPermission,
	"GetValidIngredientMeasurementUnits":                  authorization.ArchiveValidIngredientMeasurementUnitsPermission,
	"GetValidIngredientMeasurementUnitsByIngredient":      authorization.ArchiveValidIngredientMeasurementUnitsPermission,
	"GetValidIngredientMeasurementUnitsByMeasurementUnit": authorization.ArchiveValidIngredientMeasurementUnitsPermission,
	"GetValidIngredientPreparations":                      authorization.ArchiveValidIngredientPreparationsPermission,
	"GetValidIngredientPreparationsByIngredient":          authorization.ArchiveValidIngredientPreparationsPermission,
	"GetValidIngredientPreparationsByPreparation":         authorization.ArchiveValidIngredientPreparationsPermission,
	"GetValidIngredientStateIngredients":                  authorization.ArchiveValidIngredientStateIngredientsPermission,
	"GetValidIngredientStateIngredientsByIngredient":      authorization.ArchiveValidIngredientStateIngredientsPermission,
	"GetValidIngredientStateIngredientsByIngredientState": authorization.ArchiveValidIngredientStateIngredientsPermission,
	"GetValidPrepTaskConfigs":                             authorization.ArchiveValidPrepTaskConfigsPermission,
	"GetValidPrepTaskConfigsByIngredient":                 authorization.ArchiveValidPrepTaskConfigsPermission,
	"GetValidPrepTaskConfigsByPreparation":                authorization.ArchiveValidPrepTaskConfigsPermission,
	"GetValidPrepTaskConfigsByIngredientAndPreparation":   authorization.ArchiveValidPrepTaskConfigsPermission,
	"GetValidInstruments":                                 authorization.ArchiveValidInstrumentsPermission,
	"SearchForValidInstruments":                           authorization.ArchiveValidInstrumentsPermission,
	"SearchForValidInstrumentsNotOwnedByAccount":          authorization.ArchiveValidInstrumentsPermission,
	"GetValidMeasurementUnits":                            authorization.ArchiveValidMeasurementUnitsPermission,
	"SearchForValidMeasurementUnits":                      authorization.ArchiveValidMeasurementUnitsPermission,
	"SearchValidMeasurementUnitsByIngredient":             authorization.ArchiveValidMeasurementUnitsPermission,
	"GetValidMeasurementUnitConversionsForUnit":           authorization.ArchiveValidMeasurementUnitConversionsPermission,
	"GetValidPreparations":                                authorization.ArchiveValidPreparationsPermission,
	"SearchForValidPreparations":                          authorization.ArchiveValidPreparationsPermission,
	"GetValidPreparationInstruments":                      authorization.ArchiveValidPreparationInstrumentsPermission,
	"GetValidPreparationInstrumentsByInstrument":          authorization.ArchiveValidPreparationInstrumentsPermission,
	"GetValidPreparationInstrumentsByPreparation":         authorization.ArchiveValidPreparationInstrumentsPermission,
	"GetValidPreparationVessels":                          authorization.ArchiveValidPreparationVesselsPermission,
	"GetValidPreparationVesselsByPreparation":             authorization.ArchiveValidPreparationVesselsPermission,
	"GetValidPreparationVesselsByVessel":                  authorization.ArchiveValidPreparationVesselsPermission,
	"GetValidVessels":                                     authorization.ArchiveValidVesselsPermission,
	"SearchForValidVessels":                               authorization.ArchiveValidVesselsPermission,
}

// listRPCs returns every method on the service whose request carries a QueryFilter.
func listRPCs(t *testing.T) []protoreflect.MethodDescriptor {
	t.Helper()

	descriptor, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(mealplanninggrpc.MealPlanningService_ServiceDesc.ServiceName))
	require.NoError(t, err)

	filterName := (&filteringpb.QueryFilter{}).ProtoReflect().Descriptor().FullName()

	var out []protoreflect.MethodDescriptor
	methods := descriptor.(protoreflect.ServiceDescriptor).Methods()
	for i := range methods.Len() {
		method := methods.Get(i)
		if field := method.Input().Fields().ByName("filter"); field != nil && field.Message() != nil && field.Message().FullName() == filterName {
			out = append(out, method)
		}
	}

	return out
}

// buildArchivedListRequest builds the method's request asking for archived rows, with every
// string field filled so that no handler refuses it for a missing identifier.
func buildArchivedListRequest(t *testing.T, method protoreflect.MethodDescriptor) protoreflect.ProtoMessage {
	t.Helper()

	messageType, err := protoregistry.GlobalTypes.FindMessageByName(method.Input().FullName())
	require.NoError(t, err)

	request := messageType.New()
	fields := request.Descriptor().Fields()
	for i := range fields.Len() {
		if field := fields.Get(i); field.Kind() == protoreflect.StringKind && !field.IsList() {
			request.Set(field, protoreflect.ValueOfString(fake.BuildFakeID()))
		}
	}

	request.Set(fields.ByName("filter"), protoreflect.ValueOfMessage((&filteringpb.QueryFilter{IncludeArchived: new(true)}).ProtoReflect()))

	return request.Interface()
}

// filterCapturingManager answers every manager call with an empty, successful result, and keeps
// every QueryFilter it was handed. Ownership checks that answer with a bool answer yes, so the
// handlers that verify the caller's access go on to the read.
func filterCapturingManager(t *testing.T) (manager *mockmanagers.MealPlanningManagerMock, filters func() []*filtering.QueryFilter) {
	t.Helper()

	var (
		mu       sync.Mutex
		captured []*filtering.QueryFilter
	)

	filterType := reflect.TypeFor[*filtering.QueryFilter]()
	manager = &mockmanagers.MealPlanningManagerMock{}
	value := reflect.ValueOf(manager).Elem()
	for _, field := range value.Fields() {
		if field.Kind() != reflect.Func || !field.CanSet() {
			continue
		}

		fnType := field.Type()
		field.Set(reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
			for _, arg := range args {
				if arg.Type() == filterType {
					mu.Lock()
					captured = append(captured, arg.Interface().(*filtering.QueryFilter))
					mu.Unlock()
				}
			}

			results := make([]reflect.Value, fnType.NumOut())
			for j := range results {
				switch out := fnType.Out(j); out.Kind() {
				case reflect.Pointer:
					results[j] = reflect.New(out.Elem())
				case reflect.Bool:
					results[j] = reflect.ValueOf(true)
				default:
					results[j] = reflect.Zero(out)
				}
			}

			return results
		}))
	}

	return manager, func() []*filtering.QueryFilter {
		mu.Lock()
		defer mu.Unlock()

		return captured
	}
}

// callListRPC calls the method's handler as a caller holding grants, and returns the filters the
// handler handed the manager.
func callListRPC(t *testing.T, method protoreflect.MethodDescriptor, grants []authorization.Permission) []*filtering.QueryFilter {
	t.Helper()

	s := buildServiceImplForMealPlanningTest(t)
	manager, captured := filterCapturingManager(t)
	s.mealPlanningManager = manager

	accountID := fake.BuildFakeID()
	ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID:    accountID,
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{accountID: authorization.NewAccountRolePermissionChecker(grants)},
		Requester:          sessions.RequesterInfo{UserID: fake.BuildFakeID()},
	})

	handler := reflect.ValueOf(s).MethodByName(string(method.Name()))
	require.True(t, handler.IsValid(), "no handler for %s", method.Name())

	results := handler.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(buildArchivedListRequest(t, method))})
	if err, _ := results[1].Interface().(error); err != nil {
		require.NoError(t, err)
	}

	filters := captured()
	require.NotEmpty(t, filters, "%s never handed the manager a filter", method.Name())

	return filters
}

// readGrantsFor is what the authorization interceptor requires of a caller of the method: a
// caller holding exactly these is the least privileged one who reaches the handler.
func readGrantsFor(method protoreflect.MethodDescriptor) []authorization.Permission {
	fullMethod := "/" + mealplanninggrpc.MealPlanningService_ServiceDesc.ServiceName + "/" + string(method.Name())

	return ProvideMethodPermissions()[fullMethod]
}

func TestListRPCArchiveDecisions(T *testing.T) {
	T.Parallel()

	methods := listRPCs(T)

	T.Run("every list RPC has a decision", func(t *testing.T) {
		t.Parallel()

		listed := map[string]bool{}
		for _, method := range methods {
			name := string(method.Name())
			listed[name] = true

			_, ok := archiveGrantForListRPC[name]
			assert.True(t, ok, "%s takes a QueryFilter but has no archive decision here", name)
		}

		for name := range archiveGrantForListRPC {
			assert.True(t, listed[name], "%s has an archive decision here but takes no QueryFilter", name)
		}
	})

	T.Run("a caller without the archive grant gets no archived rows", func(t *testing.T) {
		t.Parallel()

		for _, method := range methods {
			t.Run(string(method.Name()), func(t *testing.T) {
				t.Parallel()

				for _, filter := range callListRPC(t, method, readGrantsFor(method)) {
					assert.Nil(t, filter.IncludeArchived, "include_archived reached the store")
				}
			})
		}
	})

	T.Run("a caller holding the archive grant gets what they asked for", func(t *testing.T) {
		t.Parallel()

		for _, method := range methods {
			grant := archiveGrantForListRPC[string(method.Name())]
			if grant == "" {
				continue
			}

			t.Run(string(method.Name()), func(t *testing.T) {
				t.Parallel()

				for _, filter := range callListRPC(t, method, append(readGrantsFor(method), grant)) {
					require.NotNil(t, filter.IncludeArchived)
					assert.True(t, *filter.IncludeArchived)
				}
			})
		}
	})

	T.Run("a denied surface stays denied for a caller holding every grant", func(t *testing.T) {
		t.Parallel()

		for _, method := range methods {
			if archiveGrantForListRPC[string(method.Name())] != "" {
				continue
			}

			t.Run(string(method.Name()), func(t *testing.T) {
				t.Parallel()

				for _, filter := range callListRPC(t, method, authorization.MealPlanningPermissions) {
					assert.Nil(t, filter.IncludeArchived, "include_archived reached the store")
				}
			})
		}
	})
}

func TestArchivedIfHeld(T *testing.T) {
	T.Parallel()

	grant := authorization.ArchiveValidIngredientsPermission

	T.Run("allowed for a service role holding the grant", func(t *testing.T) {
		t.Parallel()

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				ServicePermissions: authorization.NewServiceRolePermissionChecker(nil, []authorization.Permission{grant}),
			},
		})

		assert.True(t, archivedIfHeld(ctx, grant).Allowed())
	})

	T.Run("allowed for an account membership holding the grant", func(t *testing.T) {
		t.Parallel()

		accountID := fake.BuildFakeID()
		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			ActiveAccountID:    accountID,
			AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{accountID: authorization.NewAccountRolePermissionChecker([]authorization.Permission{grant})},
		})

		assert.True(t, archivedIfHeld(ctx, grant).Allowed())
	})

	T.Run("denied for a grant held in an account that is not the active one", func(t *testing.T) {
		t.Parallel()

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			ActiveAccountID:    fake.BuildFakeID(),
			AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{fake.BuildFakeID(): authorization.NewAccountRolePermissionChecker([]authorization.Permission{grant})},
		})

		assert.False(t, archivedIfHeld(ctx, grant).Allowed())
	})

	T.Run("denied without a session", func(t *testing.T) {
		t.Parallel()

		assert.False(t, archivedIfHeld(t.Context(), grant).Allowed())
	})
}

func TestDecodeQueryFilter(T *testing.T) {
	T.Parallel()

	T.Run("denied clears include_archived", func(t *testing.T) {
		t.Parallel()

		_, span := buildServiceImplForMealPlanningTest(t).tracer.StartSpan(t.Context())
		defer span.End()

		filter, err := decodeQueryFilter(span, &filteringpb.QueryFilter{IncludeArchived: new(true)}, filteringgrpc.ArchivedDenied)
		require.NoError(t, err)
		assert.Nil(t, filter.IncludeArchived)
	})

	T.Run("allowed keeps include_archived", func(t *testing.T) {
		t.Parallel()

		_, span := buildServiceImplForMealPlanningTest(t).tracer.StartSpan(t.Context())
		defer span.End()

		filter, err := decodeQueryFilter(span, &filteringpb.QueryFilter{IncludeArchived: new(true)}, filteringgrpc.ArchivedAllowed)
		require.NoError(t, err)
		require.NotNil(t, filter.IncludeArchived)
		assert.True(t, *filter.IncludeArchived)
	})
}
