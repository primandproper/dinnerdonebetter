package mcp

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	"github.com/modelcontextprotocol/go-sdk/auth"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// toolRequest is a tool call carrying a token issued to an account.
func toolRequest() *sdkmcp.CallToolRequest {
	return &sdkmcp.CallToolRequest{Extra: &sdkmcp.RequestExtra{TokenInfo: &auth.TokenInfo{
		Extra: map[string]any{mcptools.ClaimAccountID: fake.BuildFakeID()},
	}}}
}

func buildTestTools(t *testing.T) (*Tools, *mealplanningmock.RepositoryMock) {
	t.Helper()

	repo := &mealplanningmock.RepositoryMock{}

	tools, err := NewTools(repo)
	require.NoError(t, err)

	return tools, repo
}

// listTools opens an in-memory session against a server carrying only this surface and asks
// it what it serves.
func listTools(t *testing.T, tools *Tools) []*sdkmcp.Tool {
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

	listed, err := clientSession.ListTools(t.Context(), &sdkmcp.ListToolsParams{})
	require.NoError(t, err)

	return listed.Tools
}

func TestNewTools(T *testing.T) {
	T.Parallel()

	T.Run("refuses a nil repository", func(t *testing.T) {
		t.Parallel()

		_, err := NewTools(nil)
		require.Error(t, err)
	})
}

func TestTools_RegisterOn(T *testing.T) {
	T.Parallel()

	T.Run("registers every tool under a distinct name", func(t *testing.T) {
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
		}

		assert.Len(t, listed, registeredToolCount)
	})
}

// registeredToolCount is how many tools RegisterOn adds. It is counted by hand because the SDK
// offers no way to ask a server what it holds short of a session, and a tool dropped from
// RegisterOn would otherwise go unnoticed.
const registeredToolCount = 47

func TestTools_GetRecipe(T *testing.T) {
	T.Parallel()

	T.Run("reads the recipe through the repository", func(t *testing.T) {
		t.Parallel()

		tools, repo := buildTestTools(t)
		expected := mealplanningfakes.BuildFakeRecipe()

		repo.GetRecipeFunc = func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
			assert.Equal(t, expected.ID, recipeID)

			return expected, nil
		}

		_, actual, err := tools.GetRecipe()(t.Context(), toolRequest(), &GetRecipeInvocation{RecipeID: expected.ID})
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("refuses a caller with no token", func(t *testing.T) {
		t.Parallel()

		tools, repo := buildTestTools(t)

		_, _, err := tools.GetRecipe()(t.Context(), &sdkmcp.CallToolRequest{}, &GetRecipeInvocation{RecipeID: fake.BuildFakeID()})
		require.Error(t, err)
		assert.Empty(t, repo.GetRecipeCalls())
	})
}

func TestTools_SearchForRecipes(T *testing.T) {
	T.Parallel()

	T.Run("searches through the repository with the caller's query", func(t *testing.T) {
		t.Parallel()

		tools, repo := buildTestTools(t)
		expected := mealplanningfakes.BuildFakeRecipe()
		query := fake.BuildFakeID()

		repo.SearchForRecipesFunc = func(_ context.Context, q string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
			assert.Equal(t, query, q)

			return &filtering.QueryFilteredResult[mealplanning.Recipe]{Data: []*mealplanning.Recipe{expected}}, nil
		}

		_, actual, err := tools.SearchForRecipes()(t.Context(), toolRequest(), &SearchForRecipesInvocation{Query: query})
		require.NoError(t, err)
		require.Len(t, actual.Results, 1)
		assert.Equal(t, expected, actual.Results[0])
	})
}
