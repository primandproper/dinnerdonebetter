package grpc

import (
	"context"
	"testing"

	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/settings/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/pointer"
	settings "github.com/primandproper/platform-go/v13/settings"
	settingsmock "github.com/primandproper/platform-go/v13/settings/mock"
	"github.com/primandproper/platform-go/v13/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// errStoreUnavailable stands in for a store that cannot answer.
var errStoreUnavailable = platformerrors.New("the store is unavailable")

// pageOfDefinitions returns one full page, as a store would.
func pageOfDefinitions(definitions ...*settings.Definition) *filtering.QueryFilteredResult[settings.Definition] {
	return filtering.NewQueryFilteredResult(definitions, uint64(len(definitions)), uint64(len(definitions)),
		func(d *settings.Definition) string { return d.ID }, filtering.DefaultQueryFilter())
}

// pageOfValues returns one full page, as a store would.
func pageOfValues(values ...*settings.Value) *filtering.QueryFilteredResult[settings.Value] {
	return filtering.NewQueryFilteredResult(values, uint64(len(values)), uint64(len(values)),
		func(v *settings.Value) string { return v.ID }, filtering.DefaultQueryFilter())
}

// codeOf is the gRPC code an error reached a client as.
func codeOf(t *testing.T, err error) codes.Code {
	t.Helper()

	require.Error(t, err)

	return status.Code(err)
}

func TestService_CreateSettingDefinition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := fakes.BuildFakeSettingDefinition()

		store := &settingsmock.StoreMock{
			CreateDefinitionFunc: func(_ context.Context, scope tenancy.Scope, definition *settings.Definition) (*settings.Definition, error) {
				// Every write this service makes is in the one scope the catalog
				// lives in, whatever the request said — it cannot say anything.
				assert.Equal(t, ddbsettings.Scope(), scope)
				assert.Equal(t, example.Name, definition.Name)

				return example, nil
			},
		}

		response, err := buildTestService(t, store).CreateSettingDefinition(who.ctx, &settingssvc.CreateSettingDefinitionRequest{
			Input: &settingssvc.SettingDefinitionCreationRequestInput{
				Name:         example.Name,
				Description:  example.Description,
				Kind:         example.Kind.String(),
				DefaultValue: example.Default,
				Enumeration:  example.Enumeration,
			},
		})
		require.NoError(t, err)
		assert.Equal(t, example.ID, response.GetCreated().GetId())
		assert.Equal(t, example.Kind.String(), response.GetCreated().GetKind())
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)

		_, err := buildTestService(t, nil).CreateSettingDefinition(who.ctx, &settingssvc.CreateSettingDefinitionRequest{})
		assert.Equal(t, codes.InvalidArgument, codeOf(t, err))
	})

	T.Run("with a name already defined", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := fakes.BuildFakeSettingDefinition()

		store := &settingsmock.StoreMock{
			CreateDefinitionFunc: func(context.Context, tenancy.Scope, *settings.Definition) (*settings.Definition, error) {
				return nil, settings.ErrDefinitionNameTaken
			},
		}

		// The mapper is what turns the store's sentinel into something other than
		// "the server broke"; see internal/services/settings/errors.
		_, err := buildTestService(t, store).CreateSettingDefinition(who.ctx, &settingssvc.CreateSettingDefinitionRequest{
			Input: &settingssvc.SettingDefinitionCreationRequestInput{Name: example.Name, Kind: example.Kind.String()},
		})
		assert.Equal(t, codes.AlreadyExists, codeOf(t, err))
	})

	T.Run("without a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).CreateSettingDefinition(t.Context(), &settingssvc.CreateSettingDefinitionRequest{})
		assert.Equal(t, codes.Unauthenticated, codeOf(t, err))
	})
}

func TestService_GetSettingDefinition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := fakes.BuildFakeSettingDefinition()
		example.AdminOnly = false

		store := &settingsmock.StoreMock{
			GetDefinitionFunc: func(_ context.Context, _ tenancy.Scope, definitionID string) (*settings.Definition, error) {
				assert.Equal(t, example.ID, definitionID)

				return example, nil
			},
		}

		response, err := buildTestService(t, store).GetSettingDefinition(who.ctx, &settingssvc.GetSettingDefinitionRequest{
			SettingDefinitionId: example.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, example.Name, response.GetResult().GetName())
	})

	T.Run("refuses a non-admin an admin-only setting", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := fakes.BuildFakeSettingDefinition()
		example.AdminOnly = true

		store := &settingsmock.StoreMock{
			GetDefinitionFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return example, nil
			},
		}

		// Platform records AdminOnly and deliberately does not enforce it — it has
		// no notion of who is calling. This service is the caller it hands that to.
		_, err := buildTestService(t, store).GetSettingDefinition(who.ctx, &settingssvc.GetSettingDefinitionRequest{
			SettingDefinitionId: example.ID,
		})
		assert.Equal(t, codes.PermissionDenied, codeOf(t, err))
	})

	T.Run("admits an admin to an admin-only setting", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := fakes.BuildFakeSettingDefinition()
		example.AdminOnly = true

		store := &settingsmock.StoreMock{
			GetDefinitionFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return example, nil
			},
		}

		response, err := buildTestService(t, store).GetSettingDefinition(who.ctx, &settingssvc.GetSettingDefinitionRequest{
			SettingDefinitionId: example.ID,
		})
		require.NoError(t, err)
		assert.True(t, response.GetResult().GetAdminOnly())
	})

	T.Run("with no such setting", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &settingsmock.StoreMock{
			GetDefinitionFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return nil, settings.ErrDefinitionNotFound
			},
		}

		_, err := buildTestService(t, store).GetSettingDefinition(who.ctx, &settingssvc.GetSettingDefinitionRequest{})
		assert.Equal(t, codes.NotFound, codeOf(t, err))
	})
}

func TestService_GetSettingDefinitions(T *testing.T) {
	T.Parallel()

	T.Run("hides the admin-only settings from a non-admin", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		open := fakes.BuildFakeSettingDefinition()
		open.AdminOnly = false

		hidden := fakes.BuildFakeSettingDefinition()
		hidden.AdminOnly = true

		store := &settingsmock.StoreMock{
			ListDefinitionsFunc: func(context.Context, tenancy.Scope, *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.Definition], error) {
				return pageOfDefinitions(open, hidden), nil
			},
		}

		response, err := buildTestService(t, store).GetSettingDefinitions(who.ctx, &settingssvc.GetSettingDefinitionsRequest{})
		require.NoError(t, err)
		require.Len(t, response.GetResults(), 1)
		assert.Equal(t, open.ID, response.GetResults()[0].GetId())
	})

	T.Run("shows an admin the whole catalog", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)

		open := fakes.BuildFakeSettingDefinition()
		open.AdminOnly = false

		hidden := fakes.BuildFakeSettingDefinition()
		hidden.AdminOnly = true

		store := &settingsmock.StoreMock{
			ListDefinitionsFunc: func(context.Context, tenancy.Scope, *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.Definition], error) {
				return pageOfDefinitions(open, hidden), nil
			},
		}

		response, err := buildTestService(t, store).GetSettingDefinitions(who.ctx, &settingssvc.GetSettingDefinitionsRequest{})
		require.NoError(t, err)
		assert.Len(t, response.GetResults(), 2)
	})

	T.Run("with a store that cannot answer", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &settingsmock.StoreMock{
			ListDefinitionsFunc: func(context.Context, tenancy.Scope, *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.Definition], error) {
				return nil, errStoreUnavailable
			},
		}

		_, err := buildTestService(t, store).GetSettingDefinitions(who.ctx, &settingssvc.GetSettingDefinitionsRequest{})
		assert.Equal(t, codes.Internal, codeOf(t, err))
	})
}

func TestService_UpdateSettingDefinition(T *testing.T) {
	T.Parallel()

	T.Run("leaves an absent field alone", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := fakes.BuildFakeSettingDefinition()

		var written *settings.Definition

		store := &settingsmock.StoreMock{
			GetDefinitionFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return example, nil
			},
			UpdateDefinitionFunc: func(_ context.Context, _ tenancy.Scope, definition *settings.Definition) error {
				written = definition

				return nil
			},
		}

		response, err := buildTestService(t, store).UpdateSettingDefinition(who.ctx, &settingssvc.UpdateSettingDefinitionRequest{
			SettingDefinitionId: example.ID,
			Input:               &settingssvc.SettingDefinitionUpdateRequestInput{Description: pointer.To("a better description")},
		})
		require.NoError(t, err)

		// The edit is a read-modify-write, because platform's UpdateDefinition
		// rewrites the whole row — so an absent field has to arrive as what was
		// already stored rather than as a zero.
		require.NotNil(t, written)
		assert.Equal(t, "a better description", written.Description)
		assert.Equal(t, example.Name, written.Name)
		assert.ElementsMatch(t, example.Enumeration, written.Enumeration)
		assert.Equal(t, "a better description", response.GetUpdated().GetDescription())
	})

	T.Run("replaces an enumeration the request names", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := fakes.BuildFakeSettingDefinition()

		var written *settings.Definition

		store := &settingsmock.StoreMock{
			GetDefinitionFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return example, nil
			},
			UpdateDefinitionFunc: func(_ context.Context, _ tenancy.Scope, definition *settings.Definition) error {
				written = definition

				return nil
			},
		}

		// The wrapper message is what lets an empty list mean "this setting stops
		// enumerating" rather than "leave it alone", which a bare repeated field
		// could not express.
		_, err := buildTestService(t, store).UpdateSettingDefinition(who.ctx, &settingssvc.UpdateSettingDefinitionRequest{
			SettingDefinitionId: example.ID,
			Input: &settingssvc.SettingDefinitionUpdateRequestInput{
				Enumeration: &settingssvc.SettingDefinitionEnumeration{Values: []string{}},
			},
		})
		require.NoError(t, err)

		require.NotNil(t, written)
		assert.Empty(t, written.Enumeration)
	})

	T.Run("with an edit that would strand stored values", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := fakes.BuildFakeSettingDefinition()

		store := &settingsmock.StoreMock{
			GetDefinitionFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return example, nil
			},
			UpdateDefinitionFunc: func(context.Context, tenancy.Scope, *settings.Definition) error {
				return settings.ErrStrandedValues
			},
		}

		// FailedPrecondition rather than InvalidArgument: the request is
		// well-formed, and what refuses it is state an administrator can clear.
		_, err := buildTestService(t, store).UpdateSettingDefinition(who.ctx, &settingssvc.UpdateSettingDefinitionRequest{
			SettingDefinitionId: example.ID,
			Input:               &settingssvc.SettingDefinitionUpdateRequestInput{Name: pointer.To("renamed")},
		})
		assert.Equal(t, codes.FailedPrecondition, codeOf(t, err))
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)

		_, err := buildTestService(t, nil).UpdateSettingDefinition(who.ctx, &settingssvc.UpdateSettingDefinitionRequest{})
		assert.Equal(t, codes.InvalidArgument, codeOf(t, err))
	})
}

func TestService_ArchiveSettingDefinition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := fakes.BuildFakeSettingDefinition()

		store := &settingsmock.StoreMock{
			ArchiveDefinitionFunc: func(_ context.Context, _ tenancy.Scope, definitionID string) error {
				assert.Equal(t, example.ID, definitionID)

				return nil
			},
		}

		response, err := buildTestService(t, store).ArchiveSettingDefinition(who.ctx, &settingssvc.ArchiveSettingDefinitionRequest{
			SettingDefinitionId: example.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, response.GetResponseDetails())
	})
}

func TestService_SetSettingValue(T *testing.T) {
	T.Parallel()

	T.Run("stores the answer against the session's user", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		definition := fakes.BuildFakeSettingDefinition()
		definition.AdminOnly = false

		chosen := definition.Enumeration[0]
		stored := fakes.BuildFakeSettingValueForUser(who.userID)
		stored.Raw = chosen

		store := &settingsmock.StoreMock{
			GetDefinitionByNameFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return definition, nil
			},
			SetValueFunc: func(_ context.Context, _ tenancy.Scope, subject settings.Subject, name, raw string) (*settings.Value, error) {
				// Whose answer it is comes from the session and never from the
				// request, which is what makes writing somebody else's setting
				// unspellable rather than merely forbidden.
				assert.Equal(t, ddbsettings.SubjectFor(who.userID), subject)
				assert.Equal(t, definition.Name, name)
				assert.Equal(t, chosen, raw)

				return stored, nil
			},
		}

		response, err := buildTestService(t, store).SetSettingValue(who.ctx, &settingssvc.SetSettingValueRequest{
			SettingName: definition.Name,
			Value:       chosen,
		})
		require.NoError(t, err)
		assert.Equal(t, chosen, response.GetValue().GetValue())
		assert.Equal(t, who.userID, response.GetValue().GetBelongsToUser())
	})

	T.Run("refuses a non-admin an admin-only setting", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		definition := fakes.BuildFakeSettingDefinition()
		definition.AdminOnly = true

		store := &settingsmock.StoreMock{
			GetDefinitionByNameFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return definition, nil
			},
		}

		// The definition is read before the write so that "admin-only" means
		// something on this path too. The store's SetValue is never reached, which
		// the unconfigured mock method would panic to prove.
		_, err := buildTestService(t, store).SetSettingValue(who.ctx, &settingssvc.SetSettingValueRequest{
			SettingName: definition.Name,
			Value:       definition.Enumeration[0],
		})
		assert.Equal(t, codes.PermissionDenied, codeOf(t, err))
	})

	T.Run("with a value the setting does not admit", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		definition := fakes.BuildFakeSettingDefinition()
		definition.AdminOnly = false

		store := &settingsmock.StoreMock{
			GetDefinitionByNameFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return definition, nil
			},
			SetValueFunc: func(context.Context, tenancy.Scope, settings.Subject, string, string) (*settings.Value, error) {
				return nil, settings.ErrNotEnumerated
			},
		}

		_, err := buildTestService(t, store).SetSettingValue(who.ctx, &settingssvc.SetSettingValueRequest{
			SettingName: definition.Name,
			Value:       "not-in-the-enumeration",
		})
		assert.Equal(t, codes.InvalidArgument, codeOf(t, err))
	})

	T.Run("with no such setting", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &settingsmock.StoreMock{
			GetDefinitionByNameFunc: func(context.Context, tenancy.Scope, string) (*settings.Definition, error) {
				return nil, settings.ErrDefinitionNotFound
			},
		}

		_, err := buildTestService(t, store).SetSettingValue(who.ctx, &settingssvc.SetSettingValueRequest{SettingName: "no-such-setting"})
		assert.Equal(t, codes.NotFound, codeOf(t, err))
	})
}

func TestService_GetSettingValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		stored := fakes.BuildFakeSettingValueForUser(who.userID)

		store := &settingsmock.StoreMock{
			GetValueFunc: func(_ context.Context, _ tenancy.Scope, subject settings.Subject, _ string) (*settings.Value, error) {
				assert.Equal(t, ddbsettings.SubjectFor(who.userID), subject)

				return stored, nil
			},
		}

		response, err := buildTestService(t, store).GetSettingValue(who.ctx, &settingssvc.GetSettingValueRequest{SettingName: "whatever"})
		require.NoError(t, err)
		assert.Equal(t, stored.ID, response.GetResult().GetId())
	})

	T.Run("when they have not answered", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &settingsmock.StoreMock{
			GetValueFunc: func(context.Context, tenancy.Scope, settings.Subject, string) (*settings.Value, error) {
				return nil, settings.ErrValueNotFound
			},
		}

		// NotFound rather than an empty value: reading the raw row and resolving
		// the setting are different questions, and only the second applies a
		// default.
		_, err := buildTestService(t, store).GetSettingValue(who.ctx, &settingssvc.GetSettingValueRequest{SettingName: "whatever"})
		assert.Equal(t, codes.NotFound, codeOf(t, err))
	})
}

func TestService_GetSettingValues(T *testing.T) {
	T.Parallel()

	T.Run("pages only the requester's own answers", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		stored := fakes.BuildFakeSettingValueForUser(who.userID)

		store := &settingsmock.StoreMock{
			ListValuesForSubjectFunc: func(_ context.Context, _ tenancy.Scope, subject settings.Subject, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.Value], error) {
				assert.Equal(t, ddbsettings.SubjectFor(who.userID), subject)

				return pageOfValues(stored), nil
			},
		}

		response, err := buildTestService(t, store).GetSettingValues(who.ctx, &settingssvc.GetSettingValuesRequest{})
		require.NoError(t, err)
		require.Len(t, response.GetResults(), 1)
		assert.Equal(t, who.userID, response.GetResults()[0].GetBelongsToUser())
	})
}

func TestService_GetSettingValuesForDefinition(T *testing.T) {
	T.Parallel()

	T.Run("refuses a non-admin", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		// This is the read the table this replaced handed to any account member who
		// asked for the account's configurations: every row it returns is somebody
		// else's preference.
		_, err := buildTestService(t, nil).GetSettingValuesForDefinition(who.ctx, &settingssvc.GetSettingValuesForDefinitionRequest{
			SettingName: "whatever",
		})
		assert.Equal(t, codes.PermissionDenied, codeOf(t, err))
	})

	T.Run("answers an admin", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		stored := fakes.BuildFakeSettingValueForUser("somebody-else")

		store := &settingsmock.StoreMock{
			ListValuesForDefinitionFunc: func(_ context.Context, _ tenancy.Scope, name string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.Value], error) {
				assert.Equal(t, "whatever", name)

				return pageOfValues(stored), nil
			},
		}

		response, err := buildTestService(t, store).GetSettingValuesForDefinition(who.ctx, &settingssvc.GetSettingValuesForDefinitionRequest{
			SettingName: "whatever",
		})
		require.NoError(t, err)
		assert.Len(t, response.GetResults(), 1)
	})
}

func TestService_ClearSettingValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &settingsmock.StoreMock{
			ClearValueFunc: func(_ context.Context, _ tenancy.Scope, subject settings.Subject, name string) error {
				assert.Equal(t, ddbsettings.SubjectFor(who.userID), subject)
				assert.Equal(t, "whatever", name)

				return nil
			},
		}

		response, err := buildTestService(t, store).ClearSettingValue(who.ctx, &settingssvc.ClearSettingValueRequest{SettingName: "whatever"})
		require.NoError(t, err)
		assert.NotNil(t, response.GetResponseDetails())
	})
}

func TestService_ResolveSetting(T *testing.T) {
	T.Parallel()

	T.Run("reports where the value came from", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		definition := fakes.BuildFakeSettingDefinition()
		definition.AdminOnly = false

		store := &settingsmock.StoreMock{
			ResolveFunc: func(context.Context, tenancy.Scope, settings.Subject, string) (*settings.Resolution, error) {
				return &settings.Resolution{
					Definition: definition,
					Raw:        *definition.Default,
					Source:     settings.SourceDefault,
				}, nil
			},
		}

		response, err := buildTestService(t, store).ResolveSetting(who.ctx, &settingssvc.ResolveSettingRequest{SettingName: definition.Name})
		require.NoError(t, err)
		assert.Equal(t, settings.SourceDefault.String(), response.GetResult().GetSource())
		assert.Equal(t, *definition.Default, response.GetResult().GetRaw())
		// The value is absent because the default answered, which is what the
		// source is there to say.
		assert.Nil(t, response.GetResult().GetValue())
	})

	T.Run("carries the unset case rather than guessing", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		definition := fakes.BuildFakeSettingDefinition()
		definition.AdminOnly = false
		definition.Default = nil

		store := &settingsmock.StoreMock{
			ResolveFunc: func(context.Context, tenancy.Scope, settings.Subject, string) (*settings.Resolution, error) {
				return &settings.Resolution{Definition: definition, Source: settings.SourceUnset}, nil
			},
		}

		response, err := buildTestService(t, store).ResolveSetting(who.ctx, &settingssvc.ResolveSettingRequest{SettingName: definition.Name})
		require.NoError(t, err)
		assert.Equal(t, settings.SourceUnset.String(), response.GetResult().GetSource())
		assert.Empty(t, response.GetResult().GetRaw())
	})

	T.Run("refuses a non-admin an admin-only setting", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		definition := fakes.BuildFakeSettingDefinition()
		definition.AdminOnly = true

		store := &settingsmock.StoreMock{
			ResolveFunc: func(context.Context, tenancy.Scope, settings.Subject, string) (*settings.Resolution, error) {
				return &settings.Resolution{Definition: definition, Source: settings.SourceDefault, Raw: *definition.Default}, nil
			},
		}

		_, err := buildTestService(t, store).ResolveSetting(who.ctx, &settingssvc.ResolveSettingRequest{SettingName: definition.Name})
		assert.Equal(t, codes.PermissionDenied, codeOf(t, err))
	})
}

func TestService_ResolveSettings(T *testing.T) {
	T.Parallel()

	T.Run("answers the whole catalog for a non-admin, less the admin-only ones", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		open := fakes.BuildFakeSettingDefinition()
		open.AdminOnly = false

		hidden := fakes.BuildFakeSettingDefinition()
		hidden.AdminOnly = true

		store := &settingsmock.StoreMock{
			ResolveAllFunc: func(_ context.Context, _ tenancy.Scope, subject settings.Subject) ([]*settings.Resolution, error) {
				assert.Equal(t, ddbsettings.SubjectFor(who.userID), subject)

				return []*settings.Resolution{
					{Definition: open, Raw: *open.Default, Source: settings.SourceDefault},
					{Definition: hidden, Raw: *hidden.Default, Source: settings.SourceDefault},
				}, nil
			},
		}

		response, err := buildTestService(t, store).ResolveSettings(who.ctx, &settingssvc.ResolveSettingsRequest{})
		require.NoError(t, err)
		require.Len(t, response.GetResults(), 1)
		assert.Equal(t, open.Name, response.GetResults()[0].GetDefinition().GetName())
	})

	T.Run("with a store that cannot answer", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &settingsmock.StoreMock{
			ResolveAllFunc: func(context.Context, tenancy.Scope, settings.Subject) ([]*settings.Resolution, error) {
				return nil, errStoreUnavailable
			},
		}

		_, err := buildTestService(t, store).ResolveSettings(who.ctx, &settingssvc.ResolveSettingsRequest{})
		assert.Equal(t, codes.Internal, codeOf(t, err))
	})
}

func TestConverters_RoundTrip(T *testing.T) {
	T.Parallel()

	T.Run("a definition survives the wire", func(t *testing.T) {
		t.Parallel()

		example := fakes.BuildFakeSettingDefinition()

		back := converters.ConvertGRPCSettingDefinitionToSettingDefinition(
			converters.ConvertSettingDefinitionToGRPCSettingDefinition(example))

		assert.Equal(t, example.ID, back.ID)
		assert.Equal(t, example.Name, back.Name)
		assert.Equal(t, example.Kind, back.Kind)
		assert.Equal(t, example.Default, back.Default)
		assert.Equal(t, example.Enumeration, back.Enumeration)
		assert.Equal(t, example.AdminOnly, back.AdminOnly)
	})

	T.Run("a definition with no default keeps the absence", func(t *testing.T) {
		t.Parallel()

		example := fakes.BuildFakeSettingDefinition()
		example.Default = nil

		// The whole reason default_value is optional on the wire: a setting
		// defaulting to "" answers everybody who has not chosen, and one with no
		// default answers nobody.
		back := converters.ConvertGRPCSettingDefinitionToSettingDefinition(
			converters.ConvertSettingDefinitionToGRPCSettingDefinition(example))

		assert.Nil(t, back.Default)
	})

	T.Run("a value survives the wire", func(t *testing.T) {
		t.Parallel()

		example := fakes.BuildFakeSettingValueForUser("user_1")

		back := converters.ConvertGRPCSettingValueToSettingValue(
			converters.ConvertSettingValueToGRPCSettingValue(example))

		assert.Equal(t, example.ID, back.ID)
		assert.Equal(t, example.DefinitionID, back.DefinitionID)
		assert.Equal(t, example.Raw, back.Raw)
		assert.Equal(t, example.Subject, back.Subject)
	})

	T.Run("nil in, nil out", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, converters.ConvertSettingDefinitionToGRPCSettingDefinition(nil))
		assert.Nil(t, converters.ConvertSettingValueToGRPCSettingValue(nil))
		assert.Nil(t, converters.ConvertSettingResolutionToGRPCSettingResolution(nil))
	})
}
