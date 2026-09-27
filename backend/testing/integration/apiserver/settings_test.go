package integration

import (
	"testing"

	settingsfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	settingspb "github.com/primandproper/platform-go/v14/settings/settingspb"
	"github.com/primandproper/primitives-go/v2/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The settings surface's behavior — definitions, values, resolution, confinement to the caller's
// own subject, and the admin-only write check — is asserted by platform's settings conformance
// suite, run against this deployment in conformance_test.go, and its reach without a caller by the
// anonymous suite. What remains here is this application's own: that the catalog and the
// "who answered this" read are a service admin's, and the few store rules no suite reaches.
//
// The settings surface is platform's, and the difference that runs through this file is the
// subject.
//
// The local RPCs took a setting name and inferred whose answer it was from the session.
// platform's take the subject explicitly, because a deployment may have several kinds — an
// account's settings, a workspace's — and the library cannot know which the session implies.
// This deployment has one, and internal/build/settings.selfServiceOnly is the authorizer that
// refuses any subject but the caller's own. So every call here names a subject, and one of the
// cases below names somebody else's to check that it is refused.
//
// Two smaller shifts: a kind and a source are enums rather than strings, and SetValue takes a
// TypedValue rather than a raw string — which is what lets a boolean setting be answered with
// a boolean rather than with the word "true".

// subjectFor builds the settings subject for a user.
func settingsSubjectFor(userID string) *settingspb.SettingSubject {
	return &settingspb.SettingSubject{Type: "user", Id: userID}
}

// stringValue wraps a string for SetValue. Every definition these tests build is a string
// setting, so this is the only variant they need.
func stringValue(raw string) *settingspb.TypedValue {
	return &settingspb.TypedValue{Value: &settingspb.TypedValue_StringValue{StringValue: raw}}
}

func checkSettingDefinitionEquality(t *testing.T, expected *settingspb.SettingDefinitionInput, actual *settingspb.SettingDefinition) {
	t.Helper()

	assert.NotEmpty(t, actual.GetId(), "expected SettingDefinition to have ID")
	assert.NotNil(t, actual.GetCreatedAt(), "expected SettingDefinition to have CreatedAt")

	assert.Equal(t, expected.GetName(), actual.GetName(), "expected SettingDefinition Name")
	assert.Equal(t, expected.GetDescription(), actual.GetDescription(), "expected SettingDefinition Description")
	assert.Equal(t, expected.GetKind(), actual.GetKind(), "expected SettingDefinition Kind")
	assert.Equal(t, expected.GetDefaultValue(), actual.GetDefaultValue(), "expected SettingDefinition Default")
	assert.ElementsMatch(t, expected.GetEnumeration(), actual.GetEnumeration(), "expected SettingDefinition Enumeration")
	assert.Equal(t, expected.GetAdminOnly(), actual.GetAdminOnly(), "expected SettingDefinition AdminOnly")
}

// definitionInputForTest builds a string setting with an enumeration and a default drawn from
// it. It is deliberately not admin-only: most of what follows is a signed-in user reading and
// answering it, and an admin-only setting is one they may not see at all.
func definitionInputForTest() *settingspb.SettingDefinitionInput {
	example := settingsfakes.BuildFakeSettingDefinition()

	return &settingspb.SettingDefinitionInput{
		Name:         example.Name,
		Description:  example.Description,
		Kind:         settingspb.SettingKind_SETTING_KIND_STRING,
		DefaultValue: example.Default,
		Enumeration:  example.Enumeration,
		AdminOnly:    false,
	}
}

// createSettingDefinitionForTest adds a setting to the catalog.
//
// Definitions are administrative rows in one global catalog, so it is always the
// admin client that writes one.
func createSettingDefinitionForTest(t *testing.T, testClient client.Client) *settingspb.SettingDefinition {
	t.Helper()
	ctx := t.Context()

	input := definitionInputForTest()

	created, err := adminClient.CreateDefinition(ctx, &settingspb.CreateDefinitionRequest{Definition: input})
	require.NoError(t, err)
	checkSettingDefinitionEquality(t, input, created.GetResult())

	retrieved, err := testClient.GetDefinition(ctx, &settingspb.GetDefinitionRequest{
		DefinitionId: created.GetResult().GetId(),
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved.GetResult())
	checkSettingDefinitionEquality(t, input, retrieved.GetResult())

	return retrieved.GetResult()
}

func TestSettingDefinitions_Creating(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("refuses a default the setting would not admit", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// A default outside its own enumeration answers every subject who has not
		// chosen with a value the setting does not admit.
		input := definitionInputForTest()
		input.DefaultValue = pointer.To("not-in-the-enumeration")

		created, err := adminClient.CreateDefinition(ctx, &settingspb.CreateDefinitionRequest{Definition: input})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created, err := testClient.CreateDefinition(ctx, &settingspb.CreateDefinitionRequest{Definition: definitionInputForTest()})
		require.Error(t, err)
		assert.Nil(t, created)
	})
}

// TestSettingDefinitions_AdminOnlyIsReadable pins the other half of the admin-only check: the
// write refusal is asserted by the settings suite's reserved-settings assertions, and it is a
// write check and only a write check. The flag reaches every reader, which is what lets a client
// hide the setting from a self-service page instead of offering a control that always refuses.
func TestSettingDefinitions_AdminOnlyIsReadable(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	input := definitionInputForTest()
	input.AdminOnly = true

	created, err := adminClient.CreateDefinition(T.Context(), &settingspb.CreateDefinitionRequest{Definition: input})
	require.NoError(T, err)

	// A non-admin reads it, and the flag is what tells them not to offer it.
	//
	// AdminOnly marks a setting only an administrator may *write*; platform puts it on the
	// wire deliberately, so a self-service page can leave it out rather than render a
	// control whose every use is refused. Hiding the definition would leave a client
	// guessing why a name in its own catalog does not resolve.
	T.Run("a non-admin reads it and is told it is reserved", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, readErr := testClient.GetDefinition(ctx, &settingspb.GetDefinitionRequest{
			DefinitionId: created.GetResult().GetId(),
		})
		require.NoError(t, readErr)
		assert.True(t, retrieved.GetResult().GetAdminOnly())
	})
}

func TestSettingDefinitions_Updating(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		_, err := testClient.UpdateDefinition(ctx, &settingspb.UpdateDefinitionRequest{
			DefinitionId: created.GetId(),
			Definition: &settingspb.SettingDefinitionInput{
				Name:        created.GetName(),
				Description: "nope",
				Kind:        created.GetKind(),
			},
		})
		assert.Error(t, err)
	})
}

func TestSettingDefinitions_Archiving(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		_, err := testClient.ArchiveDefinition(ctx, &settingspb.ArchiveDefinitionRequest{DefinitionId: created.GetId()})
		assert.Error(t, err)
	})
}

func TestSettingValues_Answering(T *testing.T) {
	T.Parallel()

	user, testClient := createUserAndClientForTest(T)
	subject := settingsSubjectFor(user.ID)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)
		chosen := definition.GetEnumeration()[0]

		set, err := testClient.SetValue(ctx, &settingspb.SetValueRequest{
			Subject: subject,
			Name:    definition.GetName(),
			Value:   stringValue(chosen),
		})
		require.NoError(t, err)
		assert.Equal(t, chosen, set.GetResolution().GetValue().GetRaw())
		// Whose answer it is comes from the subject, and the authorizer has already
		// refused any subject but the caller's own.
		assert.Equal(t, user.ID, set.GetResolution().GetValue().GetSubject().GetId())

		read, err := testClient.GetValue(ctx, &settingspb.GetValueRequest{Subject: subject, Name: definition.GetName()})
		require.NoError(t, err)
		assert.Equal(t, set.GetResolution().GetValue().GetId(), read.GetResult().GetId())

		// A second answer converges on the same row rather than writing another.
		second := definition.GetEnumeration()[1]

		changed, err := testClient.SetValue(ctx, &settingspb.SetValueRequest{
			Subject: subject,
			Name:    definition.GetName(),
			Value:   stringValue(second),
		})
		require.NoError(t, err)
		assert.Equal(t, set.GetResolution().GetValue().GetId(), changed.GetResolution().GetValue().GetId())
		assert.Equal(t, second, changed.GetResolution().GetValue().GetRaw())

		mine, err := testClient.ListValuesForSubject(ctx, &settingspb.ListValuesForSubjectRequest{Subject: subject})
		require.NoError(t, err)
		require.NotEmpty(t, mine.GetResults())
		for _, value := range mine.GetResults() {
			assert.Equal(t, user.ID, value.GetSubject().GetId(), "this read is the requester's own answers and nobody else's")
		}
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

	firstUser, firstClient := createUserAndClientForTest(T)
	_, secondClient := createUserAndClientForTest(T)

	definition := createSettingDefinitionForTest(T, firstClient)

	_, err := firstClient.SetValue(T.Context(), &settingspb.SetValueRequest{
		Subject: settingsSubjectFor(firstUser.ID),
		Name:    definition.GetName(),
		Value:   stringValue(definition.GetEnumeration()[0]),
	})
	require.NoError(T, err)

	T.Run("and they cannot ask who has answered it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, listErr := secondClient.ListValuesForDefinition(ctx, &settingspb.ListValuesForDefinitionRequest{
			Name: definition.GetName(),
		})
		assert.Error(t, listErr)
	})
}
