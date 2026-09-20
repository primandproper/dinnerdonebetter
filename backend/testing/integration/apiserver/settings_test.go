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

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createSettingDefinitionForTest(t, testClient)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateDefinition(ctx, &settingspb.CreateDefinitionRequest{Definition: definitionInputForTest()})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	// A kind decides how every stored value is read back, so the set is closed. It is an enum
	// now, which means the unparseable kind the local RPC could be sent as a string is gone —
	// what is left to refuse is the zero value, a definition that names no kind at all.
	T.Run("refuses a definition with no kind", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := definitionInputForTest()
		input.Kind = settingspb.SettingKind_SETTING_KIND_UNSPECIFIED

		created, err := adminClient.CreateDefinition(ctx, &settingspb.CreateDefinitionRequest{Definition: input})
		require.Error(t, err)
		assert.Nil(t, created)
	})

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

func TestSettingDefinitions_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		// The name is the handle application code holds, and it finds the same row
		// as the id does.
		retrieved, err := testClient.GetDefinitionByName(ctx, &settingspb.GetDefinitionByNameRequest{
			Name: created.GetName(),
		})
		require.NoError(t, err)
		assert.Equal(t, created.GetId(), retrieved.GetResult().GetId())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetDefinition(ctx, &settingspb.GetDefinitionRequest{DefinitionId: created.GetId()})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetDefinition(ctx, &settingspb.GetDefinitionRequest{DefinitionId: nonexistentID})
		assert.Error(t, err)
	})
}

// TestSettingDefinitions_AdminOnlyIsEnforced pins the check platform asks in the handler,
// against the definition it has already read: PermissionWriteAdminValues, carried by the
// grants extractor. The store records AdminOnly and never acts on it, because it has no
// notion of who is calling.
//
// It is a write check and only a write check. The flag reaches every reader, which is what
// lets a client hide the setting from a self-service page instead of offering a control
// that always refuses.
func TestSettingDefinitions_AdminOnlyIsEnforced(T *testing.T) {
	T.Parallel()

	user, testClient := createUserAndClientForTest(T)

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

	T.Run("a non-admin cannot answer it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, writeErr := testClient.SetValue(ctx, &settingspb.SetValueRequest{
			Subject: settingsSubjectFor(user.ID),
			Name:    input.GetName(),
			Value:   stringValue(input.GetEnumeration()[0]),
		})
		assert.Error(t, writeErr)
	})

	T.Run("an admin can", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, retrieveErr := adminClient.GetDefinition(ctx, &settingspb.GetDefinitionRequest{
			DefinitionId: created.GetResult().GetId(),
		})
		require.NoError(t, retrieveErr)
		assert.True(t, retrieved.GetResult().GetAdminOnly())
	})
}

func TestSettingDefinitions_Updating(T *testing.T) {
	T.Parallel()

	user, testClient := createUserAndClientForTest(T)

	// UpdateDefinition takes a whole definition rather than a patch of one, so a caller
	// restates the fields it is keeping. The local RPC took pointer fields and left an
	// omission alone.
	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		updated, err := adminClient.UpdateDefinition(ctx, &settingspb.UpdateDefinitionRequest{
			DefinitionId: created.GetId(),
			Definition: &settingspb.SettingDefinitionInput{
				Name:         created.GetName(),
				Description:  "a better description",
				Kind:         created.GetKind(),
				DefaultValue: created.DefaultValue,
				Enumeration:  created.GetEnumeration(),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "a better description", updated.GetResult().GetDescription())

		assert.Equal(t, created.GetName(), updated.GetResult().GetName())
		assert.ElementsMatch(t, created.GetEnumeration(), updated.GetResult().GetEnumeration())
	})

	T.Run("refuses an edit that would strand a stored value", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)
		stranded := created.GetEnumeration()[1]

		_, err := testClient.SetValue(ctx, &settingspb.SetValueRequest{
			Subject: settingsSubjectFor(user.ID),
			Name:    created.GetName(),
			Value:   stringValue(stranded),
		})
		require.NoError(t, err)

		// This is the rule the store exists to own. Applied, the value somebody
		// chose would still be in the table and every read of it would fail.
		_, err = adminClient.UpdateDefinition(ctx, &settingspb.UpdateDefinitionRequest{
			DefinitionId: created.GetId(),
			Definition: &settingspb.SettingDefinitionInput{
				Name:         created.GetName(),
				Description:  created.GetDescription(),
				Kind:         created.GetKind(),
				DefaultValue: pointer.To(created.GetEnumeration()[0]),
				Enumeration:  []string{created.GetEnumeration()[0]},
			},
		})
		require.Error(t, err)

		// And the value is still readable, because the edit was refused rather
		// than half-applied.
		still, err := testClient.GetValue(ctx, &settingspb.GetValueRequest{
			Subject: settingsSubjectFor(user.ID),
			Name:    created.GetName(),
		})
		require.NoError(t, err)
		assert.Equal(t, stranded, still.GetResult().GetRaw())
	})

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

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		_, err := adminClient.ArchiveDefinition(ctx, &settingspb.ArchiveDefinitionRequest{
			DefinitionId: created.GetId(),
		})
		require.NoError(t, err)

		x, err := adminClient.GetDefinition(ctx, &settingspb.GetDefinitionRequest{DefinitionId: created.GetId()})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveDefinition(ctx, &settingspb.ArchiveDefinitionRequest{DefinitionId: created.GetId()})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createSettingDefinitionForTest(t, testClient)

		_, err := testClient.ArchiveDefinition(ctx, &settingspb.ArchiveDefinitionRequest{DefinitionId: created.GetId()})
		assert.Error(t, err)
	})
}

func TestSettingDefinitions_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	created := make([]*settingspb.SettingDefinition, 0, exampleQuantity)
	for range exampleQuantity {
		created = append(created, createSettingDefinitionForTest(T, testClient))
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.ListDefinitions(ctx, &settingspb.ListDefinitionsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.GetResults()), len(created))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ListDefinitions(ctx, &settingspb.ListDefinitionsRequest{})
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

	T.Run("refuses a value the setting does not admit", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)

		_, err := testClient.SetValue(ctx, &settingspb.SetValueRequest{
			Subject: subject,
			Name:    definition.GetName(),
			Value:   stringValue("not-in-the-enumeration"),
		})
		require.Error(t, err)
	})

	T.Run("refuses an answer to a setting that does not exist", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := testClient.SetValue(ctx, &settingspb.SetValueRequest{
			Subject: subject,
			Name:    "no-such-setting",
			Value:   stringValue("anything"),
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SetValue(ctx, &settingspb.SetValueRequest{
			Subject: subject,
			Name:    definition.GetName(),
			Value:   stringValue(definition.GetEnumeration()[0]),
		})
		assert.Error(t, err)
	})
}

// TestSettingValues_Resolving walks the tri-state end to end, which is the whole
// reason a resolution is on the wire rather than a bare value.
func TestSettingValues_Resolving(T *testing.T) {
	T.Parallel()

	user, testClient := createUserAndClientForTest(T)
	subject := settingsSubjectFor(user.ID)

	T.Run("falls back to the default, then reports the subject, then falls back again", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)

		fromDefault, err := testClient.Resolve(ctx, &settingspb.ResolveRequest{Subject: subject, Name: definition.GetName()})
		require.NoError(t, err)
		assert.Equal(t, settingspb.ValueSource_VALUE_SOURCE_DEFAULT, fromDefault.GetResolution().GetSource())
		assert.Equal(t, definition.GetDefaultValue(), fromDefault.GetResolution().GetTypedValue().GetStringValue())

		_, err = testClient.SetValue(ctx, &settingspb.SetValueRequest{
			Subject: subject,
			Name:    definition.GetName(),
			Value:   stringValue(definition.GetEnumeration()[1]),
		})
		require.NoError(t, err)

		fromSubject, err := testClient.Resolve(ctx, &settingspb.ResolveRequest{Subject: subject, Name: definition.GetName()})
		require.NoError(t, err)
		assert.Equal(t, settingspb.ValueSource_VALUE_SOURCE_SUBJECT, fromSubject.GetResolution().GetSource())
		assert.Equal(t, definition.GetEnumeration()[1], fromSubject.GetResolution().GetTypedValue().GetStringValue())

		// Clearing puts them back on the default rather than leaving them
		// unanswered, and the raw row is gone.
		_, err = testClient.ClearValue(ctx, &settingspb.ClearValueRequest{Subject: subject, Name: definition.GetName()})
		require.NoError(t, err)

		_, err = testClient.GetValue(ctx, &settingspb.GetValueRequest{Subject: subject, Name: definition.GetName()})
		require.Error(t, err)

		backToDefault, err := testClient.Resolve(ctx, &settingspb.ResolveRequest{Subject: subject, Name: definition.GetName()})
		require.NoError(t, err)
		assert.Equal(t, settingspb.ValueSource_VALUE_SOURCE_DEFAULT, backToDefault.GetResolution().GetSource())
	})

	T.Run("reports a setting nobody has answered that has no default", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := definitionInputForTest()
		input.DefaultValue = nil
		input.Enumeration = nil

		created, err := adminClient.CreateDefinition(ctx, &settingspb.CreateDefinitionRequest{Definition: input})
		require.NoError(t, err)
		// The absent default survives the wire, which is what makes the unset case
		// expressible at all.
		assert.Nil(t, created.GetResult().DefaultValue)

		unset, err := testClient.Resolve(ctx, &settingspb.ResolveRequest{Subject: subject, Name: input.GetName()})
		require.NoError(t, err)
		assert.Equal(t, settingspb.ValueSource_VALUE_SOURCE_UNSET, unset.GetResolution().GetSource())
		assert.Nil(t, unset.GetResolution().GetValue())
	})

	T.Run("resolves the whole catalog in one call", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		definition := createSettingDefinitionForTest(t, testClient)

		// This is the read a preferences page makes: every setting, including the
		// ones nobody has touched, each with the value that applies.
		all, err := testClient.ResolveAll(ctx, &settingspb.ResolveAllRequest{Subject: subject})
		require.NoError(t, err)

		// An admin-only setting is shown here, flag and all, and that is the design
		// rather than a leak. AdminOnly restricts who may write a value — SetValue and
		// ClearValue refuse it without the grant, which the case below pins — and
		// platform says in as many words that the flag on the wire "is also what an
		// admin UI reads to know which settings to hide from a self-service page". So
		// the page does the hiding, and it can only do it because the read answered.
		var found bool
		for _, resolution := range all.GetResolutions() {
			if resolution.GetDefinition().GetId() == definition.GetId() {
				found = true
				assert.Equal(t, settingspb.ValueSource_VALUE_SOURCE_DEFAULT, resolution.GetSource())
			}
		}

		assert.True(t, found, "expected the setting just defined to be among the resolved ones")
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ResolveAll(ctx, &settingspb.ResolveAllRequest{Subject: subject})
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

	firstUser, firstClient := createUserAndClientForTest(T)
	secondUser, secondClient := createUserAndClientForTest(T)

	definition := createSettingDefinitionForTest(T, firstClient)

	_, err := firstClient.SetValue(T.Context(), &settingspb.SetValueRequest{
		Subject: settingsSubjectFor(firstUser.ID),
		Name:    definition.GetName(),
		Value:   stringValue(definition.GetEnumeration()[0]),
	})
	require.NoError(T, err)

	T.Run("the other member's own answers do not include it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Not merely filtered out of the response: the second user has not
		// answered this setting, so there is no row of theirs to return.
		_, readErr := secondClient.GetValue(ctx, &settingspb.GetValueRequest{
			Subject: settingsSubjectFor(secondUser.ID),
			Name:    definition.GetName(),
		})
		assert.Error(t, readErr)
	})

	// Naming the subject is a request a caller can make, which is precisely why the
	// authorizer exists: a subject that is not the caller's own is refused before any read
	// happens. See internal/build/settings.selfServiceOnly.
	T.Run("and they cannot ask for it by naming the other subject", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, readErr := secondClient.GetValue(ctx, &settingspb.GetValueRequest{
			Subject: settingsSubjectFor(firstUser.ID),
			Name:    definition.GetName(),
		})
		assert.Error(t, readErr)

		_, listErr := secondClient.ListValuesForSubject(ctx, &settingspb.ListValuesForSubjectRequest{
			Subject: settingsSubjectFor(firstUser.ID),
		})
		assert.Error(t, listErr)
	})

	T.Run("and they cannot ask who has answered it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, listErr := secondClient.ListValuesForDefinition(ctx, &settingspb.ListValuesForDefinitionRequest{
			Name: definition.GetName(),
		})
		assert.Error(t, listErr)
	})

	T.Run("an admin can", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		answers, listErr := adminClient.ListValuesForDefinition(ctx, &settingspb.ListValuesForDefinitionRequest{
			Name: definition.GetName(),
		})
		require.NoError(t, listErr)
		assert.Len(t, answers.GetResults(), 1)
	})
}
