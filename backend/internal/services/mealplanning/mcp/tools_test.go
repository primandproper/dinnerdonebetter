package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registeredToolCount is how many tools RegisterOn adds. It is counted by hand because the SDK
// offers no way to ask a server what it holds short of a session, and a tool dropped from
// RegisterOn would otherwise go unnoticed.
const registeredToolCount = 47

// everyPermission is every grant the surface's tools can require, for a caller the tests want
// admitted everywhere.
func everyPermission() []authorization.Permission {
	var perms []authorization.Permission
	for _, required := range methodPermissions {
		perms = append(perms, required...)
	}

	return perms
}

// gateAdmitting is a gate whose authenticator puts a caller holding perms on every call.
func gateAdmitting(t *testing.T, perms ...authorization.Permission) *mcptools.Gate {
	t.Helper()

	authenticate := func(ctx context.Context, _ *sdkmcp.CallToolRequest) (context.Context, error) {
		return sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID:             fake.BuildFakeID(),
				ServicePermissions: authorization.NewServiceRolePermissionChecker(nil, perms),
			},
			ActiveAccountID: fake.BuildFakeID(),
		}), nil
	}

	gate, err := mcptools.NewGate(authenticate, sessions.PrincipalFromContext, sessions.GrantsFromContext)
	require.NoError(t, err)

	return gate
}

// gateRefusing is a gate whose authenticator finds no credential on any call.
func gateRefusing(t *testing.T) *mcptools.Gate {
	t.Helper()

	gate, err := mcptools.NewGate(
		func(ctx context.Context, _ *sdkmcp.CallToolRequest) (context.Context, error) { return ctx, nil },
		sessions.PrincipalFromContext,
		sessions.GrantsFromContext,
	)
	require.NoError(t, err)

	return gate
}

// call is a tool call; what it carries is the gate's to decide.
func call() *sdkmcp.CallToolRequest {
	return &sdkmcp.CallToolRequest{}
}

func buildTestTools(t *testing.T, perms ...authorization.Permission) (*Tools, *mealplanningmock.MealPlanningManagerMock) {
	t.Helper()

	return buildTestToolsBehind(t, gateAdmitting(t, perms...))
}

func buildTestToolsBehind(t *testing.T, gate *mcptools.Gate) (*Tools, *mealplanningmock.MealPlanningManagerMock) {
	t.Helper()

	manager := &mealplanningmock.MealPlanningManagerMock{}

	tools, err := NewTools(manager, gate)
	require.NoError(t, err)

	return tools, manager
}

// connect opens an in-memory session against a server carrying only this surface.
func connect(t *testing.T, tools *Tools) *sdkmcp.ClientSession {
	t.Helper()

	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: t.Name(), Version: "test"}, nil)
	tools.RegisterOn(server)

	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()

	serverSession, err := server.Connect(t.Context(), serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: t.Name() + "-client", Version: "test"}, nil)

	clientSession, err := client.Connect(t.Context(), clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession
}

// listTools asks a server carrying only this surface what it serves.
func listTools(t *testing.T, tools *Tools) []*sdkmcp.Tool {
	t.Helper()

	listed, err := connect(t, tools).ListTools(t.Context(), &sdkmcp.ListToolsParams{})
	require.NoError(t, err)

	return listed.Tools
}

func TestNewTools(T *testing.T) {
	T.Parallel()

	T.Run("refuses a nil manager", func(t *testing.T) {
		t.Parallel()

		_, err := NewTools(nil, gateAdmitting(t))
		require.ErrorIs(t, err, ErrNilManager)
	})

	T.Run("refuses a nil gate", func(t *testing.T) {
		t.Parallel()

		_, err := NewTools(&mealplanningmock.MealPlanningManagerMock{}, nil)
		require.ErrorIs(t, err, ErrNilGate)
	})
}

func TestTools_RegisterOn(T *testing.T) {
	T.Parallel()

	T.Run("registers every tool under a distinct name, read-only, with both schemas", func(t *testing.T) {
		t.Parallel()

		tools, _ := buildTestTools(t)

		listed := listTools(t, tools)
		require.NotEmpty(t, listed)

		names := make(map[string]struct{}, len(listed))
		for _, tool := range listed {
			// The SDK keeps the last tool registered under a name and says nothing about
			// the first, so a collision here would surface as a shorter list rather than
			// an error. The count is asserted against the registrations for that reason.
			_, seen := names[tool.Name]
			assert.False(t, seen, "tool %q registered twice", tool.Name)
			names[tool.Name] = struct{}{}

			assert.NotEmpty(t, tool.Description, tool.Name)
			assert.NotNil(t, tool.InputSchema, tool.Name)
			assert.NotNil(t, tool.OutputSchema, tool.Name)

			require.NotNil(t, tool.Annotations, tool.Name)
			assert.True(t, tool.Annotations.ReadOnlyHint, tool.Name)
		}

		assert.Len(t, listed, registeredToolCount)
	})

	T.Run("every tool is a method of the gRPC surface with a declared permission", func(t *testing.T) {
		t.Parallel()

		tools, _ := buildTestTools(t)

		// A tool is named for its gRPC counterpart, and that is what its permission is read
		// from. One whose name matches no method would be refused on every call, so the
		// mismatch is caught here rather than by the first model to try it.
		for _, tool := range listTools(t, tools) {
			required, err := permissionsFor(tool)
			require.NoError(t, err, tool.Name)
			assert.NotEmpty(t, required, "tool %q requires no permission", tool.Name)
		}
	})

	T.Run("the schemas name the properties the wire carries", func(t *testing.T) {
		t.Parallel()

		tools, _ := buildTestTools(t)

		var getRecipe *sdkmcp.Tool
		for _, tool := range listTools(t, tools) {
			if tool.Name == getRecipeTool.Name {
				getRecipe = tool
			}
		}
		require.NotNil(t, getRecipe)

		// The hand-written schemas this replaced described the output as "ID" and
		// "CreatedAt" where the type marshals "id" and "createdAt": a model reading the
		// schema was told last year's names. Reflected off the type, the two cannot differ.
		input := properties(t, getRecipe.InputSchema)
		assert.Contains(t, input, "recipeID")

		output := properties(t, getRecipe.OutputSchema)
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "createdAt")
		assert.NotContains(t, output, "ID")
	})
}

// properties is the property names of a schema as the client decoded it.
func properties(t *testing.T, schema any) []string {
	t.Helper()

	encoded, err := json.Marshal(schema)
	require.NoError(t, err)

	var decoded struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(encoded, &decoded))

	names := make([]string, 0, len(decoded.Properties))
	for name := range decoded.Properties {
		names = append(names, name)
	}

	return names
}

func TestTools_GetRecipe(T *testing.T) {
	T.Parallel()

	T.Run("reads the recipe through the manager", func(t *testing.T) {
		t.Parallel()

		tools, manager := buildTestTools(t, authorization.ReadRecipesPermission)
		expected := mealplanningfakes.BuildFakeRecipe()

		manager.ReadRecipeFunc = func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
			assert.Equal(t, expected.ID, recipeID)

			return expected, nil
		}

		_, actual, err := tools.GetRecipe(t.Context(), call(), GetRecipeInput{RecipeID: expected.ID})
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("refuses a caller with no token before reading anything", func(t *testing.T) {
		t.Parallel()

		tools, manager := buildTestToolsBehind(t, gateRefusing(t))

		_, _, err := tools.GetRecipe(t.Context(), call(), GetRecipeInput{RecipeID: fake.BuildFakeID()})
		require.ErrorIs(t, err, mcptools.ErrNoPrincipal)
		assert.Empty(t, manager.ReadRecipeCalls())
	})

	T.Run("refuses a caller without the grant the gRPC method requires", func(t *testing.T) {
		t.Parallel()

		// Every grant but the one GetRecipe's gRPC counterpart declares.
		var perms []authorization.Permission
		for _, perm := range everyPermission() {
			if perm != authorization.ReadRecipesPermission {
				perms = append(perms, perm)
			}
		}

		tools, manager := buildTestTools(t, perms...)

		_, _, err := tools.GetRecipe(t.Context(), call(), GetRecipeInput{RecipeID: fake.BuildFakeID()})
		require.ErrorIs(t, err, mcptools.ErrPermissionDenied)
		assert.Empty(t, manager.ReadRecipeCalls())
	})

	T.Run("answers the row over the wire under the type's names", func(t *testing.T) {
		t.Parallel()

		tools, manager := buildTestTools(t, authorization.ReadValidIngredientsPermission)
		expected := mealplanningfakes.BuildFakeValidIngredient()

		manager.ReadValidIngredientFunc = func(context.Context, string) (*mealplanning.ValidIngredient, error) {
			return expected, nil
		}

		// Through a session rather than the handler: the SDK validates the arguments
		// against the input schema and the answer against the output schema, so this is
		// what proves the reflected schemas describe what is sent and what comes back.
		res, err := connect(t, tools).CallTool(t.Context(), &sdkmcp.CallToolParams{
			Name:      getValidIngredientTool.Name,
			Arguments: map[string]any{"validIngredientID": expected.ID},
		})
		require.NoError(t, err)
		require.False(t, res.IsError, errorText(res))

		encoded, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		actual := &mealplanning.ValidIngredient{}
		require.NoError(t, json.Unmarshal(encoded, actual))
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
	})
}

func TestTools_SearchForRecipes(T *testing.T) {
	T.Parallel()

	T.Run("searches the database through the manager with the caller's query", func(t *testing.T) {
		t.Parallel()

		tools, manager := buildTestTools(t, authorization.ReadRecipesPermission)
		expected := mealplanningfakes.BuildFakeRecipe()
		query := fake.BuildFakeID()

		manager.SearchRecipesFunc = func(_ context.Context, q string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
			assert.Equal(t, query, q)
			assert.False(t, useSearchService)

			return &filtering.QueryFilteredResult[mealplanning.Recipe]{Data: []*mealplanning.Recipe{expected}}, nil
		}

		_, actual, err := tools.SearchForRecipes(t.Context(), call(), SearchInput{Query: query})
		require.NoError(t, err)
		require.Len(t, actual.Data, 1)
		assert.Equal(t, expected, actual.Data[0])
	})
}

func TestTools_begin(T *testing.T) {
	T.Parallel()

	T.Run("refuses a tool the gRPC surface does not declare", func(t *testing.T) {
		t.Parallel()

		tools, _ := buildTestTools(t, everyPermission()...)

		_, err := tools.begin(t.Context(), call(), &sdkmcp.Tool{Name: "DeleteEverything"})
		require.ErrorIs(t, err, ErrUndeclaredTool)
	})
}

// errorText is what a failed tool call said.
func errorText(res *sdkmcp.CallToolResult) string {
	for _, content := range res.Content {
		if text, ok := content.(*sdkmcp.TextContent); ok {
			return text.Text
		}
	}

	return ""
}
