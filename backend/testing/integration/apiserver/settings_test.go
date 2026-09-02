package integration

import (
	"testing"

	settingsfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/settings/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	settings "github.com/primandproper/platform-go/v13/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkSettingDefinitionEquality(t *testing.T, expected, actual *settings.Definition) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected SettingDefinition to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected SettingDefinition to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected SettingDefinition Name")
	assert.Equal(t, expected.Description, actual.Description, "expected SettingDefinition Description")
	assert.Equal(t, expected.Kind, actual.Kind, "expected SettingDefinition Kind")
	assert.Equal(t, expected.Default, actual.Default, "expected SettingDefinition Default")
	assert.ElementsMatch(t, expected.Enumeration, actual.Enumeration, "expected SettingDefinition Enumeration")
	assert.Equal(t, expected.AdminOnly, actual.AdminOnly, "expected SettingDefinition AdminOnly")
}

// createSettingDefinitionForTest adds a setting to the catalog.
//
// Definitions are administrative rows in one global catalog, so it is always the
// admin client that writes one. The definition is deliberately not admin-only:
// most of what follows is a signed-in user reading and answering it, and an
// admin-only setting is one they may not see at all.
func createSettingDefinitionForTest(t *testing.T, testClient client.Client) *settings.Definition {
	t.Helper()
	ctx := t.Context()

	example := settingsfakes.BuildFakeSettingDefinition()
	example.AdminOnly = false

	created, err := adminClient.CreateSettingDefinition(ctx, &settingssvc.CreateSettingDefinitionRequest{
		Input: &settingssvc.SettingDefinitionCreationRequestInput{
			Name:         example.Name,
			Description:  example.Description,
			Kind:         example.Kind.String(),
			DefaultValue: example.Default,
			Enumeration:  example.Enumeration,
			AdminOnly:    example.AdminOnly,
		},
	})
	require.NoError(t, err)

	converted := grpcconverters.ConvertGRPCSettingDefinitionToSettingDefinition(created.GetCreated())
	checkSettingDefinitionEquality(t, example, converted)

	retrieved, err := testClient.GetSettingDefinition(ctx, &settingssvc.GetSettingDefinitionRequest{
		SettingDefinitionId: created.GetCreated().GetId(),
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	definition := grpcconverters.ConvertGRPCSettingDefinitionToSettingDefinition(retrieved.GetResult())
	checkSettingDefinitionEquality(t, converted, definition)

	return definition
}

func TestSettingDefinitions_Creating(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createSettingDefinitionForTest(t, testClient)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		example := settingsfakes.BuildFakeSettingDefinition()

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateSettingDefinition(ctx, &settingssvc.CreateSettingDefinitionRequest{
			Input: &settingssvc.SettingDefinitionCreationRequestInput{Name: example.Name, Kind: example.Kind.String()},
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("refuses a kind the store cannot parse", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		example := settingsfakes.BuildFakeSettingDefinition()

		// A kind decides how every stored value is read back, so the set is closed:
		// one the store does not implement is a value nothing could ever parse.
		created, err := adminClient.CreateSettingDefinition(ctx, &settingssvc.CreateSettingDefinitionRequest{
			Input: &settingssvc.SettingDefinitionCreationRequestInput{Name: example.Name, Kind: "duration"},
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("refuses a default the setting would not admit", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		example := settingsfakes.BuildFakeSettingDefinition()

		// A default outside its own enumeration answers every subject who has not
		// chosen with a value the setting does not admit.
		created, err := adminClient.CreateSettingDefinition(ctx, &settingssvc.CreateSettingDefinitionRequest{
			Input: &settingssvc.SettingDefinitionCreationRequestInput{
				Name:         example.Name,
				Kind:         example.Kind.String(),
				Enumeration:  example.Enumeration,
				DefaultValue: new("not-in-the-enumeration"),
			},
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		example := settingsfakes.BuildFakeSettingDefinition()

		created, err := testClient.CreateSettingDefinition(ctx, &settingssvc.CreateSettingDefinitionRequest{
			Input: &settingssvc.SettingDefinitionCreationRequestInput{Name: example.Name, Kind: example.Kind.String()},
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestSettingDefinitions_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		// The name is the handle application code holds, and it finds the same row
		// as the id does.
		retrieved, err := testClient.GetSettingDefinitionByName(ctx, &settingssvc.GetSettingDefinitionByNameRequest{
			SettingName: created.Name,
		})
		require.NoError(t, err)

		checkSettingDefinitionEquality(t, created, grpcconverters.ConvertGRPCSettingDefinitionToSettingDefinition(retrieved.GetResult()))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetSettingDefinition(ctx, &settingssvc.GetSettingDefinitionRequest{SettingDefinitionId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetSettingDefinition(ctx, &settingssvc.GetSettingDefinitionRequest{SettingDefinitionId: nonexistentID})
		assert.Error(t, err)
	})
}

// TestSettingDefinitions_AdminOnlyIsEnforced pins the check platform deliberately
// leaves to its caller: the store records AdminOnly and never acts on it, because
// it has no notion of who is calling.
func TestSettingDefinitions_AdminOnlyIsEnforced(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	example := settingsfakes.BuildFakeSettingDefinition()
	example.AdminOnly = true

	created, err := adminClient.CreateSettingDefinition(T.Context(), &settingssvc.CreateSettingDefinitionRequest{
		Input: &settingssvc.SettingDefinitionCreationRequestInput{
			Name:         example.Name,
			Description:  example.Description,
			Kind:         example.Kind.String(),
			DefaultValue: example.Default,
			Enumeration:  example.Enumeration,
			AdminOnly:    true,
		},
	})
	require.NoError(T, err)

	T.Run("a non-admin cannot read it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, readErr := testClient.GetSettingDefinition(ctx, &settingssvc.GetSettingDefinitionRequest{
			SettingDefinitionId: created.GetCreated().GetId(),
		})
		assert.Error(t, readErr)
	})

	T.Run("a non-admin cannot answer it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, writeErr := testClient.SetSettingValue(ctx, &settingssvc.SetSettingValueRequest{
			SettingName: example.Name,
			Value:       example.Enumeration[0],
		})
		assert.Error(t, writeErr)
	})

	T.Run("an admin can", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, retrieveErr := adminClient.GetSettingDefinition(ctx, &settingssvc.GetSettingDefinitionRequest{
			SettingDefinitionId: created.GetCreated().GetId(),
		})
		require.NoError(t, retrieveErr)
		assert.True(t, retrieved.GetResult().GetAdminOnly())
	})
}

func TestSettingDefinitions_Updating(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		updated, err := adminClient.UpdateSettingDefinition(ctx, &settingssvc.UpdateSettingDefinitionRequest{
			SettingDefinitionId: created.ID,
			Input:               &settingssvc.SettingDefinitionUpdateRequestInput{Description: new("a better description")},
		})
		require.NoError(t, err)
		assert.Equal(t, "a better description", updated.GetUpdated().GetDescription())

		// An absent field is left as it was, which is what makes this a partial
		// edit rather than a replace with holes in it.
		assert.Equal(t, created.Name, updated.GetUpdated().GetName())
		assert.ElementsMatch(t, created.Enumeration, updated.GetUpdated().GetEnumeration())
	})

	T.Run("refuses an edit that would strand a stored value", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)
		stranded := created.Enumeration[1]

		_, err := testClient.SetSettingValue(ctx, &settingssvc.SetSettingValueRequest{
			SettingName: created.Name,
			Value:       stranded,
		})
		require.NoError(t, err)

		// This is the rule the store exists to own. Applied, the value somebody
		// chose would still be in the table and every read of it would fail.
		_, err = adminClient.UpdateSettingDefinition(ctx, &settingssvc.UpdateSettingDefinitionRequest{
			SettingDefinitionId: created.ID,
			Input: &settingssvc.SettingDefinitionUpdateRequestInput{
				Enumeration:  &settingssvc.SettingDefinitionEnumeration{Values: []string{created.Enumeration[0]}},
				DefaultValue: new(created.Enumeration[0]),
			},
		})
		require.Error(t, err)

		// And the value is still readable, because the edit was refused rather
		// than half-applied.
		still, err := testClient.GetSettingValue(ctx, &settingssvc.GetSettingValueRequest{SettingName: created.Name})
		require.NoError(t, err)
		assert.Equal(t, stranded, still.GetResult().GetValue())
	})

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		_, err := testClient.UpdateSettingDefinition(ctx, &settingssvc.UpdateSettingDefinitionRequest{
			SettingDefinitionId: created.ID,
			Input:               &settingssvc.SettingDefinitionUpdateRequestInput{Description: new("nope")},
		})
		assert.Error(t, err)
	})
}

func TestSettingDefinitions_Archiving(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		_, err := adminClient.ArchiveSettingDefinition(ctx, &settingssvc.ArchiveSettingDefinitionRequest{
			SettingDefinitionId: created.ID,
		})
		require.NoError(t, err)

		x, err := adminClient.GetSettingDefinition(ctx, &settingssvc.GetSettingDefinitionRequest{SettingDefinitionId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveSettingDefinition(ctx, &settingssvc.ArchiveSettingDefinitionRequest{SettingDefinitionId: created.ID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		_, err := testClient.ArchiveSettingDefinition(ctx, &settingssvc.ArchiveSettingDefinitionRequest{SettingDefinitionId: created.ID})
		assert.Error(t, err)
	})
}

func TestSettingDefinitions_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	created := make([]*settings.Definition, 0, exampleQuantity)
	for range exampleQuantity {
		created = append(created, createSettingDefinitionForTest(T, testClient))
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetSettingDefinitions(ctx, &settingssvc.GetSettingDefinitionsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.GetResults()), len(created))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetSettingDefinitions(ctx, &settingssvc.GetSettingDefinitionsRequest{})
		assert.Error(t, err)
	})
}

func TestSettingValues_Answering(T *testing.T) {
	T.Parallel()

	user, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)
		chosen := definition.Enumeration[0]

		set, err := testClient.SetSettingValue(ctx, &settingssvc.SetSettingValueRequest{
			SettingName: definition.Name,
			Value:       chosen,
		})
		require.NoError(t, err)
		assert.Equal(t, chosen, set.GetValue().GetValue())
		// Whose answer it is comes from the session, never from the request.
		assert.Equal(t, user.ID, set.GetValue().GetBelongsToUser())

		read, err := testClient.GetSettingValue(ctx, &settingssvc.GetSettingValueRequest{SettingName: definition.Name})
		require.NoError(t, err)
		assert.Equal(t, set.GetValue().GetId(), read.GetResult().GetId())

		// A second answer converges on the same row rather than writing another.
		second := definition.Enumeration[1]

		changed, err := testClient.SetSettingValue(ctx, &settingssvc.SetSettingValueRequest{
			SettingName: definition.Name,
			Value:       second,
		})
		require.NoError(t, err)
		assert.Equal(t, set.GetValue().GetId(), changed.GetValue().GetId())
		assert.Equal(t, second, changed.GetValue().GetValue())

		mine, err := testClient.GetSettingValues(ctx, &settingssvc.GetSettingValuesRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, mine.GetResults())
		for _, value := range mine.GetResults() {
			assert.Equal(t, user.ID, value.GetBelongsToUser(), "this read is the requester's own answers and nobody else's")
		}
	})

	T.Run("refuses a value the setting does not admit", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)

		_, err := testClient.SetSettingValue(ctx, &settingssvc.SetSettingValueRequest{
			SettingName: definition.Name,
			Value:       "not-in-the-enumeration",
		})
		require.Error(t, err)
	})

	T.Run("refuses an answer to a setting that does not exist", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := testClient.SetSettingValue(ctx, &settingssvc.SetSettingValueRequest{
			SettingName: "no-such-setting",
			Value:       "anything",
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SetSettingValue(ctx, &settingssvc.SetSettingValueRequest{
			SettingName: definition.Name,
			Value:       definition.Enumeration[0],
		})
		assert.Error(t, err)
	})
}

// TestSettingValues_Resolving walks the tri-state end to end, which is the whole
// reason a resolution is on the wire rather than a bare value.
func TestSettingValues_Resolving(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("falls back to the default, then reports the subject, then falls back again", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)

		fromDefault, err := testClient.ResolveSetting(ctx, &settingssvc.ResolveSettingRequest{SettingName: definition.Name})
		require.NoError(t, err)
		assert.Equal(t, settings.SourceDefault.String(), fromDefault.GetResult().GetSource())
		assert.Equal(t, *definition.Default, fromDefault.GetResult().GetRaw())

		_, err = testClient.SetSettingValue(ctx, &settingssvc.SetSettingValueRequest{
			SettingName: definition.Name,
			Value:       definition.Enumeration[1],
		})
		require.NoError(t, err)

		fromSubject, err := testClient.ResolveSetting(ctx, &settingssvc.ResolveSettingRequest{SettingName: definition.Name})
		require.NoError(t, err)
		assert.Equal(t, settings.SourceSubject.String(), fromSubject.GetResult().GetSource())
		assert.Equal(t, definition.Enumeration[1], fromSubject.GetResult().GetRaw())

		// Clearing puts them back on the default rather than leaving them
		// unanswered, and the raw row is gone.
		_, err = testClient.ClearSettingValue(ctx, &settingssvc.ClearSettingValueRequest{SettingName: definition.Name})
		require.NoError(t, err)

		_, err = testClient.GetSettingValue(ctx, &settingssvc.GetSettingValueRequest{SettingName: definition.Name})
		require.Error(t, err)

		backToDefault, err := testClient.ResolveSetting(ctx, &settingssvc.ResolveSettingRequest{SettingName: definition.Name})
		require.NoError(t, err)
		assert.Equal(t, settings.SourceDefault.String(), backToDefault.GetResult().GetSource())
	})

	T.Run("reports a setting nobody has answered that has no default", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		example := settingsfakes.BuildFakeSettingDefinition()

		created, err := adminClient.CreateSettingDefinition(ctx, &settingssvc.CreateSettingDefinitionRequest{
			Input: &settingssvc.SettingDefinitionCreationRequestInput{
				Name: example.Name,
				Kind: example.Kind.String(),
			},
		})
		require.NoError(t, err)
		// The absent default survives the wire, which is what makes the unset case
		// expressible at all.
		assert.Nil(t, created.GetCreated().DefaultValue)

		unset, err := testClient.ResolveSetting(ctx, &settingssvc.ResolveSettingRequest{SettingName: example.Name})
		require.NoError(t, err)
		assert.Equal(t, settings.SourceUnset.String(), unset.GetResult().GetSource())
		assert.Empty(t, unset.GetResult().GetRaw())
	})

	T.Run("resolves the whole catalog in one call", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)

		// This is the read a preferences page makes: every setting, including the
		// ones nobody has touched, each with the value that applies.
		all, err := testClient.ResolveSettings(ctx, &settingssvc.ResolveSettingsRequest{})
		require.NoError(t, err)

		var found bool
		for _, resolution := range all.GetResults() {
			assert.False(t, resolution.GetDefinition().GetAdminOnly(), "a non-admin is shown no admin-only setting")

			if resolution.GetDefinition().GetId() == definition.ID {
				found = true
				assert.Equal(t, settings.SourceDefault.String(), resolution.GetSource())
			}
		}

		assert.True(t, found, "expected the setting just defined to be among the resolved ones")
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ResolveSettings(ctx, &settingssvc.ResolveSettingsRequest{})
		assert.Error(t, err)
	})
}

// TestSettingValues_AreNotReadableByOtherMembers is the leak this adoption closed.
//
// The table it replaced filed a configuration against a user and an account at
// once, and the account read filtered on the account alone — so any member
// holding read.service_setting_configurations was handed every other member's
// personal preferences. There is no read that does that now: a person's answers
// are theirs, and the administrative "who has overridden this" is a service
// admin's.
func TestSettingValues_AreNotReadableByOtherMembers(T *testing.T) {
	T.Parallel()

	_, firstClient := createUserAndClientForTest(T)
	_, secondClient := createUserAndClientForTest(T)

	definition := createSettingDefinitionForTest(T, firstClient)

	_, err := firstClient.SetSettingValue(T.Context(), &settingssvc.SetSettingValueRequest{
		SettingName: definition.Name,
		Value:       definition.Enumeration[0],
	})
	require.NoError(T, err)

	T.Run("the other member's own answers do not include it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Not merely filtered out of the response: the second user has not
		// answered this setting, so there is no row of theirs to return.
		_, readErr := secondClient.GetSettingValue(ctx, &settingssvc.GetSettingValueRequest{SettingName: definition.Name})
		assert.Error(t, readErr)
	})

	T.Run("and they cannot ask who has answered it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, listErr := secondClient.GetSettingValuesForDefinition(ctx, &settingssvc.GetSettingValuesForDefinitionRequest{
			SettingName: definition.Name,
		})
		assert.Error(t, listErr)
	})

	T.Run("an admin can", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		answers, listErr := adminClient.GetSettingValuesForDefinition(ctx, &settingssvc.GetSettingValuesForDefinitionRequest{
			SettingName: definition.Name,
		})
		require.NoError(t, listErr)
		assert.Len(t, answers.GetResults(), 1)
	})
}
