package privacy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/mock"

	platformdataprivacy "github.com/primandproper/platform-go/v10/dataprivacy"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v10/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollector_Collect(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInvitation := fakes.BuildFakeAccountInvitation()

		repo := &identitymock.RepositoryMock{
			GetUserFunc: func(_ context.Context, actualUserID string) (*identity.User, error) {
				assert.Equal(t, exampleUser.ID, actualUserID)

				return exampleUser, nil
			},
			GetAccountsFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
				return singlePage(exampleAccount, func(a *identity.Account) string { return a.ID }), nil
			},
			GetPendingAccountInvitationsFromUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
				return singlePage(exampleInvitation, func(i *identity.AccountInvitation) string { return i.ID }), nil
			},
			GetPendingAccountInvitationsForUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
				return emptyPage[identity.AccountInvitation](), nil
			},
		}

		collector := NewCollector(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		fragment, err := collector.Collect(t.Context(), subject(exampleUser.ID))
		require.NoError(t, err)
		require.NotNil(t, fragment)

		var collection identity.UserDataCollection
		require.NoError(t, json.Unmarshal(fragment, &collection))

		assert.Equal(t, exampleUser.ID, collection.User.ID)
		assert.Len(t, collection.Accounts, 1)
		assert.Len(t, collection.AccountInvitations, 1)
	})

	T.Run("deduplicates an invitation the subject both sent and received", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleInvitation := fakes.BuildFakeAccountInvitation()

		// The same row on both sides, which is what happens when somebody invites an
		// address that turns out to be their own. An export listing it twice is the sort
		// of thing a subject notices and asks about.
		repo := &identitymock.RepositoryMock{
			GetUserFunc: func(context.Context, string) (*identity.User, error) {
				return exampleUser, nil
			},
			GetAccountsFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
				return emptyPage[identity.Account](), nil
			},
			GetPendingAccountInvitationsFromUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
				return singlePage(exampleInvitation, func(i *identity.AccountInvitation) string { return i.ID }), nil
			},
			GetPendingAccountInvitationsForUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
				return singlePage(exampleInvitation, func(i *identity.AccountInvitation) string { return i.ID }), nil
			},
		}

		collector := NewCollector(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		fragment, err := collector.Collect(t.Context(), subject(exampleUser.ID))
		require.NoError(t, err)

		var collection identity.UserDataCollection
		require.NoError(t, json.Unmarshal(fragment, &collection))

		assert.Len(t, collection.AccountInvitations, 1)
	})

	T.Run("an unreadable user is an error rather than an empty fragment", func(t *testing.T) {
		t.Parallel()

		// Every other section of an export may legitimately be empty. This one may not:
		// a document asserting that nothing is held about a person is the one wrong
		// answer available, so the request fails instead of shipping it.
		repo := &identitymock.RepositoryMock{
			GetUserFunc: func(context.Context, string) (*identity.User, error) {
				return nil, platformerrors.New("blah")
			},
		}

		collector := NewCollector(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		fragment, err := collector.Collect(t.Context(), subject(identifiers.New()))

		require.Error(t, err)
		assert.Nil(t, fragment)
	})
}

func TestResolveAccountIDs(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleUserID := identifiers.New()

		repo := &identitymock.RepositoryMock{
			GetAccountsFunc: func(_ context.Context, actualUserID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
				assert.Equal(t, exampleUserID, actualUserID)

				return singlePage(exampleAccount, func(a *identity.Account) string { return a.ID }), nil
			},
		}

		ids, err := ResolveAccountIDs(repo)(t.Context(), exampleUserID)

		require.NoError(t, err)
		assert.Equal(t, []string{exampleAccount.ID}, ids)
	})

	T.Run("with error fetching accounts", func(t *testing.T) {
		t.Parallel()

		repo := &identitymock.RepositoryMock{
			GetAccountsFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
				return nil, platformerrors.New("blah")
			},
		}

		ids, err := ResolveAccountIDs(repo)(t.Context(), identifiers.New())

		require.Error(t, err)
		assert.Nil(t, ids)
	})
}

func subject(userID string) platformdataprivacy.Subject {
	return platformdataprivacy.Subject{ID: userID, Type: platformdataprivacy.SubjectUser}
}

func singlePage[T any](item *T, idExtractor func(*T) string) *filtering.QueryFilteredResult[T] {
	return filtering.NewQueryFilteredResult([]*T{item}, 1, 1, idExtractor, filtering.DefaultQueryFilter())
}

func emptyPage[T any]() *filtering.QueryFilteredResult[T] {
	return filtering.NewQueryFilteredResult([]*T{}, 0, 0, func(*T) string { return "" }, filtering.DefaultQueryFilter())
}
