package grpc

import (
	"context"
	"testing"
	"time"

	ddbwaitlists "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/waitlists/grpc/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/pointer"
	"github.com/primandproper/platform-go/v13/tenancy"
	waitlists "github.com/primandproper/platform-go/v13/waitlists"
	waitlistsmock "github.com/primandproper/platform-go/v13/waitlists/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// assertCode asserts that err reached the client as the given gRPC code.
func assertCode(t *testing.T, err error, expected codes.Code) {
	t.Helper()

	require.Error(t, err)
	assert.Equal(t, expected, status.Code(err), "got %v", err)
}

func TestServiceImpl_CreateWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("opens the list in the global catalog", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlist()

		var written *waitlists.List
		store := &waitlistsmock.StoreMock{
			CreateListFunc: func(_ context.Context, scope tenancy.Scope, list *waitlists.List) (*waitlists.List, error) {
				assert.Equal(t, ddbwaitlists.Scope(), scope)

				list.ID = fake.BuildFakeID()
				written = list

				return list, nil
			},
		}

		res, err := buildTestService(t, store).CreateWaitlist(who.ctx, &waitlistssvc.CreateWaitlistRequest{
			Input: converters.ConvertWaitlistToGRPCWaitlistCreationRequestInput(example),
		})
		require.NoError(t, err)
		require.NotNil(t, res.GetCreated())

		assert.Equal(t, example.Name, written.Name)
		assert.Equal(t, example.Description, written.Description)
		assert.WithinDuration(t, example.ClosesAt, written.ClosesAt, time.Second)
		assert.Equal(t, written.ID, res.GetCreated().GetId())
	})

	t.Run("refuses a list the store will not take", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)

		store := &waitlistsmock.StoreMock{
			CreateListFunc: func(context.Context, tenancy.Scope, *waitlists.List) (*waitlists.List, error) {
				return nil, waitlists.ErrEmptyClosesAt
			},
		}

		_, err := buildTestService(t, store).CreateWaitlist(who.ctx, &waitlistssvc.CreateWaitlistRequest{
			Input: &waitlistssvc.WaitlistCreationRequestInput{Name: "launch"},
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("refuses a request with no input", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)

		_, err := buildTestService(t, nil).CreateWaitlist(who.ctx, &waitlistssvc.CreateWaitlistRequest{})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).CreateWaitlist(t.Context(), &waitlistssvc.CreateWaitlistRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_GetWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("reads one list", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlist()

		store := &waitlistsmock.StoreMock{
			GetListFunc: func(_ context.Context, scope tenancy.Scope, listID string) (*waitlists.List, error) {
				assert.Equal(t, ddbwaitlists.Scope(), scope)
				assert.Equal(t, example.ID, listID)

				return example, nil
			},
		}

		res, err := buildTestService(t, store).GetWaitlist(who.ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: example.ID})
		require.NoError(t, err)
		assert.Equal(t, example.ID, res.GetResult().GetId())
	})

	t.Run("reports an absent list as not found", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &waitlistsmock.StoreMock{
			GetListFunc: func(context.Context, tenancy.Scope, string) (*waitlists.List, error) {
				return nil, waitlists.ErrListNotFound
			},
		}

		_, err := buildTestService(t, store).GetWaitlist(who.ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: fake.BuildFakeID()})
		assertCode(t, err, codes.NotFound)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).GetWaitlist(t.Context(), &waitlistssvc.GetWaitlistRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_GetWaitlists(t *testing.T) {
	t.Parallel()

	t.Run("pages the catalog", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		page := waitlistfakes.BuildFakeWaitlistList()

		store := &waitlistsmock.StoreMock{
			ListListsFunc: func(_ context.Context, scope tenancy.Scope, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.List], error) {
				assert.Equal(t, ddbwaitlists.Scope(), scope)

				return page, nil
			},
		}

		res, err := buildTestService(t, store).GetWaitlists(who.ctx, &waitlistssvc.GetWaitlistsRequest{})
		require.NoError(t, err)
		assert.Len(t, res.GetResults(), len(page.Data))
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).GetWaitlists(t.Context(), &waitlistssvc.GetWaitlistsRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

// TestServiceImpl_GetOpenWaitlists pins that the open page is its own read.
//
// It is not GetWaitlists with the closed ones dropped afterwards: a page filtered
// after the fact is a page whose size the caller cannot rely on.
func TestServiceImpl_GetOpenWaitlists(t *testing.T) {
	t.Parallel()

	t.Run("pages the lists still taking signups", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		page := waitlistfakes.BuildFakeWaitlistList()

		store := &waitlistsmock.StoreMock{
			ListOpenListsFunc: func(_ context.Context, scope tenancy.Scope, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.List], error) {
				assert.Equal(t, ddbwaitlists.Scope(), scope)

				return page, nil
			},
		}

		res, err := buildTestService(t, store).GetOpenWaitlists(who.ctx, &waitlistssvc.GetOpenWaitlistsRequest{})
		require.NoError(t, err)
		assert.Len(t, res.GetResults(), len(page.Data))

		// The catalog read is not what answered.
		assert.Empty(t, store.ListListsCalls())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).GetOpenWaitlists(t.Context(), &waitlistssvc.GetOpenWaitlistsRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_UpdateWaitlist(t *testing.T) {
	t.Parallel()

	// The store's update takes a whole list, so an input carrying one field has to
	// be merged into the row as read. Writing the request straight through would
	// blank whatever the client left out.
	t.Run("merges the input into the list as read", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlist()

		var written *waitlists.List
		store := &waitlistsmock.StoreMock{
			GetListFunc: func(context.Context, tenancy.Scope, string) (*waitlists.List, error) {
				return example, nil
			},
			UpdateListFunc: func(_ context.Context, _ tenancy.Scope, list *waitlists.List) error {
				written = list

				return nil
			},
		}

		res, err := buildTestService(t, store).UpdateWaitlist(who.ctx, &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: example.ID,
			Input:      &waitlistssvc.WaitlistUpdateRequestInput{Name: pointer.To("renamed")},
		})
		require.NoError(t, err)

		assert.Equal(t, "renamed", written.Name)
		assert.Equal(t, example.Description, written.Description)
		assert.Equal(t, example.ClosesAt, written.ClosesAt)
		assert.Equal(t, "renamed", res.GetUpdated().GetName())
	})

	t.Run("reports an absent list as not found", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)

		store := &waitlistsmock.StoreMock{
			GetListFunc: func(context.Context, tenancy.Scope, string) (*waitlists.List, error) {
				return nil, waitlists.ErrListNotFound
			},
		}

		_, err := buildTestService(t, store).UpdateWaitlist(who.ctx, &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: fake.BuildFakeID(),
			Input:      &waitlistssvc.WaitlistUpdateRequestInput{Name: pointer.To("renamed")},
		})
		assertCode(t, err, codes.NotFound)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).UpdateWaitlist(t.Context(), &waitlistssvc.UpdateWaitlistRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_ArchiveWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("retires the list", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		listID := fake.BuildFakeID()

		store := &waitlistsmock.StoreMock{
			ArchiveListFunc: func(_ context.Context, scope tenancy.Scope, id string) error {
				assert.Equal(t, ddbwaitlists.Scope(), scope)
				assert.Equal(t, listID, id)

				return nil
			},
		}

		res, err := buildTestService(t, store).ArchiveWaitlist(who.ctx, &waitlistssvc.ArchiveWaitlistRequest{WaitlistId: listID})
		require.NoError(t, err)
		assert.NotNil(t, res.GetResponseDetails())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).ArchiveWaitlist(t.Context(), &waitlistssvc.ArchiveWaitlistRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

// TestServiceImpl_WaitlistIsOpen pins that the answer is List.OpenAt's rather
// than a comparison spelled in the handler — archived counts as closed, which a
// bare closes_at comparison would get wrong.
func TestServiceImpl_WaitlistIsOpen(t *testing.T) {
	t.Parallel()

	t.Run("says a live future-dated list is open", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlist()

		store := &waitlistsmock.StoreMock{
			GetListFunc: func(context.Context, tenancy.Scope, string) (*waitlists.List, error) {
				return example, nil
			},
		}

		res, err := buildTestService(t, store).WaitlistIsOpen(who.ctx, &waitlistssvc.WaitlistIsOpenRequest{WaitlistId: example.ID})
		require.NoError(t, err)
		assert.True(t, res.GetIsOpen())
	})

	t.Run("says a past-dated list is closed", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlist()
		example.ClosesAt = time.Now().Add(-time.Hour).UTC()

		store := &waitlistsmock.StoreMock{
			GetListFunc: func(context.Context, tenancy.Scope, string) (*waitlists.List, error) {
				return example, nil
			},
		}

		res, err := buildTestService(t, store).WaitlistIsOpen(who.ctx, &waitlistssvc.WaitlistIsOpenRequest{WaitlistId: example.ID})
		require.NoError(t, err)
		assert.False(t, res.GetIsOpen())
	})

	t.Run("says an archived list is closed whatever its closing time says", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlist()
		example.ArchivedAt = pointer.To(time.Now().UTC())

		store := &waitlistsmock.StoreMock{
			GetListFunc: func(context.Context, tenancy.Scope, string) (*waitlists.List, error) {
				return example, nil
			},
		}

		res, err := buildTestService(t, store).WaitlistIsOpen(who.ctx, &waitlistssvc.WaitlistIsOpenRequest{WaitlistId: example.ID})
		require.NoError(t, err)
		assert.False(t, res.GetIsOpen())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).WaitlistIsOpen(t.Context(), &waitlistssvc.WaitlistIsOpenRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_JoinWaitlist(t *testing.T) {
	t.Parallel()

	// The address and the subject come off the session. A signup that could name
	// either is one anybody could make on anybody's behalf — and, because a
	// withdrawal suppresses an address, a suppression anybody could evade.
	t.Run("signs the caller up under the session's address", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		listID := fake.BuildFakeID()

		var written *waitlists.Signup
		store := &waitlistsmock.StoreMock{
			JoinFunc: func(_ context.Context, scope tenancy.Scope, id string, signup *waitlists.Signup) (*waitlists.Signup, error) {
				assert.Equal(t, ddbwaitlists.Scope(), scope)
				assert.Equal(t, listID, id)

				signup.ID = fake.BuildFakeID()
				signup.Status = waitlists.StatusWaiting
				written = signup

				return signup, nil
			},
		}

		res, err := buildTestService(t, store).JoinWaitlist(who.ctx, &waitlistssvc.JoinWaitlistRequest{
			WaitlistId: listID,
			Input:      &waitlistssvc.WaitlistSignupCreationRequestInput{Notes: "please"},
		})
		require.NoError(t, err)

		assert.Equal(t, who.email, written.Contact)
		assert.Equal(t, ddbwaitlists.SubjectFor(who.userID), written.Subject)
		assert.Equal(t, "please", written.Notes)
		assert.Equal(t, waitlists.StatusWaiting.String(), res.GetCreated().GetStatus())
	})

	// The suppression is the whole reason this store was adopted, so the code it
	// reaches a client as is worth pinning: PermissionDenied, not a conflict. A
	// client that retried on a conflict would be a client that re-subscribed
	// somebody who asked to be left alone.
	t.Run("refuses a contact that has withdrawn", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &waitlistsmock.StoreMock{
			JoinFunc: func(context.Context, tenancy.Scope, string, *waitlists.Signup) (*waitlists.Signup, error) {
				return nil, waitlists.ErrContactWithdrawn
			},
		}

		_, err := buildTestService(t, store).JoinWaitlist(who.ctx, &waitlistssvc.JoinWaitlistRequest{
			WaitlistId: fake.BuildFakeID(),
			Input:      &waitlistssvc.WaitlistSignupCreationRequestInput{},
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("refuses a contact already on the list", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &waitlistsmock.StoreMock{
			JoinFunc: func(context.Context, tenancy.Scope, string, *waitlists.Signup) (*waitlists.Signup, error) {
				return nil, waitlists.ErrAlreadySignedUp
			},
		}

		_, err := buildTestService(t, store).JoinWaitlist(who.ctx, &waitlistssvc.JoinWaitlistRequest{
			WaitlistId: fake.BuildFakeID(),
			Input:      &waitlistssvc.WaitlistSignupCreationRequestInput{},
		})
		assertCode(t, err, codes.AlreadyExists)
	})

	t.Run("refuses a closed list", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		store := &waitlistsmock.StoreMock{
			JoinFunc: func(context.Context, tenancy.Scope, string, *waitlists.Signup) (*waitlists.Signup, error) {
				return nil, waitlists.ErrListClosed
			},
		}

		_, err := buildTestService(t, store).JoinWaitlist(who.ctx, &waitlistssvc.JoinWaitlistRequest{
			WaitlistId: fake.BuildFakeID(),
			Input:      &waitlistssvc.WaitlistSignupCreationRequestInput{},
		})
		assertCode(t, err, codes.FailedPrecondition)
	})

	t.Run("refuses a request with no input", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)

		_, err := buildTestService(t, nil).JoinWaitlist(who.ctx, &waitlistssvc.JoinWaitlistRequest{WaitlistId: fake.BuildFakeID()})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).JoinWaitlist(t.Context(), &waitlistssvc.JoinWaitlistRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_GetWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("reads the caller's own signup", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(who.userID)

		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(_ context.Context, scope tenancy.Scope, listID, signupID string) (*waitlists.Signup, error) {
				assert.Equal(t, ddbwaitlists.Scope(), scope)
				assert.Equal(t, example.ListID, listID)
				assert.Equal(t, example.ID, signupID)

				return example, nil
			},
		}

		res, err := buildTestService(t, store).GetWaitlistSignup(who.ctx, &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, example.ID, res.GetResult().GetId())
		assert.Equal(t, example.Contact, res.GetResult().GetContact())
	})

	// A signup is user-owned, which the scope cannot express: every list here is
	// global, so what makes a signup somebody's is its subject.
	t.Run("refuses somebody else's signup", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(fake.BuildFakeID())

		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
		}

		_, err := buildTestService(t, store).GetWaitlistSignup(who.ctx, &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		assertCode(t, err, codes.PermissionDenied)
	})

	t.Run("lets a service admin read anybody's", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(fake.BuildFakeID())

		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
		}

		res, err := buildTestService(t, store).GetWaitlistSignup(who.ctx, &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, example.ID, res.GetResult().GetId())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).GetWaitlistSignup(t.Context(), &waitlistssvc.GetWaitlistSignupRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

// TestServiceImpl_GetWaitlistSignupsForWaitlist pins that the list-wide read is
// service-admin-only. It hands back every signatory's address, which is exactly
// the read a member of the list must not have.
func TestServiceImpl_GetWaitlistSignupsForWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("pages one list's signups for a service admin", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		page := waitlistfakes.BuildFakeWaitlistSignupList()
		listID := fake.BuildFakeID()

		store := &waitlistsmock.StoreMock{
			ListSignupsFunc: func(_ context.Context, scope tenancy.Scope, id string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.Signup], error) {
				assert.Equal(t, ddbwaitlists.Scope(), scope)
				assert.Equal(t, listID, id)

				return page, nil
			},
		}

		res, err := buildTestService(t, store).GetWaitlistSignupsForWaitlist(who.ctx, &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: listID,
		})
		require.NoError(t, err)
		assert.Len(t, res.GetResults(), len(page.Data))
	})

	t.Run("refuses an ordinary user before it reads anything", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		store := &waitlistsmock.StoreMock{}

		_, err := buildTestService(t, store).GetWaitlistSignupsForWaitlist(who.ctx, &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: fake.BuildFakeID(),
		})
		assertCode(t, err, codes.PermissionDenied)
		assert.Empty(t, store.ListSignupsCalls())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).GetWaitlistSignupsForWaitlist(t.Context(), &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_UpdateWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("rewrites the note on the caller's own signup", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(who.userID)

		var written string
		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
			UpdateSignupNotesFunc: func(_ context.Context, _ tenancy.Scope, _, _, notes string) error {
				written = notes

				return nil
			},
		}

		res, err := buildTestService(t, store).UpdateWaitlistSignup(who.ctx, &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
			Input:            &waitlistssvc.WaitlistSignupUpdateRequestInput{Notes: pointer.To("changed my mind")},
		})
		require.NoError(t, err)
		assert.Equal(t, "changed my mind", written)
		assert.Equal(t, "changed my mind", res.GetUpdated().GetNotes())
	})

	// An input that names no note leaves the stored one alone, which is what the
	// optional field means. Writing the zero value through would erase a note by
	// omission.
	t.Run("leaves the note alone when the input names none", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(who.userID)

		var written string
		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
			UpdateSignupNotesFunc: func(_ context.Context, _ tenancy.Scope, _, _, notes string) error {
				written = notes

				return nil
			},
		}

		_, err := buildTestService(t, store).UpdateWaitlistSignup(who.ctx, &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
			Input:            &waitlistssvc.WaitlistSignupUpdateRequestInput{},
		})
		require.NoError(t, err)
		assert.Equal(t, example.Notes, written)
	})

	t.Run("refuses somebody else's signup", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(fake.BuildFakeID())

		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
		}

		_, err := buildTestService(t, store).UpdateWaitlistSignup(who.ctx, &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
			Input:            &waitlistssvc.WaitlistSignupUpdateRequestInput{Notes: pointer.To("nope")},
		})
		assertCode(t, err, codes.PermissionDenied)
		assert.Empty(t, store.UpdateSignupNotesCalls())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).UpdateWaitlistSignup(t.Context(), &waitlistssvc.UpdateWaitlistSignupRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

// TestServiceImpl_InviteWaitlistSignup pins that working the queue is the
// operator's. Being on a list does not entitle somebody to invite themselves off
// it, which is the whole difference between a queue and a form.
func TestServiceImpl_InviteWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("moves the signup and reads back where it landed", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(fake.BuildFakeID())

		invited := *example
		invited.Status = waitlists.StatusInvited
		invited.StatusChangedAt = pointer.To(time.Now().UTC())

		store := &waitlistsmock.StoreMock{
			InviteFunc: func(_ context.Context, scope tenancy.Scope, listID, signupID string) error {
				assert.Equal(t, ddbwaitlists.Scope(), scope)
				assert.Equal(t, example.ListID, listID)
				assert.Equal(t, example.ID, signupID)

				return nil
			},
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return &invited, nil
			},
		}

		res, err := buildTestService(t, store).InviteWaitlistSignup(who.ctx, &waitlistssvc.InviteWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, waitlists.StatusInvited.String(), res.GetUpdated().GetStatus())
		assert.NotNil(t, res.GetUpdated().GetStatusChangedAt())
	})

	// The guard is what makes an invitation happen once. Two operators inviting
	// the same person send one email between them, and the second is told so.
	t.Run("reports a lost guard as a failed precondition", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)

		store := &waitlistsmock.StoreMock{
			InviteFunc: func(context.Context, tenancy.Scope, string, string) error {
				return waitlists.ErrWrongStatus
			},
		}

		_, err := buildTestService(t, store).InviteWaitlistSignup(who.ctx, &waitlistssvc.InviteWaitlistSignupRequest{
			WaitlistId:       fake.BuildFakeID(),
			WaitlistSignupId: fake.BuildFakeID(),
		})
		assertCode(t, err, codes.FailedPrecondition)
	})

	t.Run("refuses an ordinary user before it writes anything", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		store := &waitlistsmock.StoreMock{}

		_, err := buildTestService(t, store).InviteWaitlistSignup(who.ctx, &waitlistssvc.InviteWaitlistSignupRequest{
			WaitlistId:       fake.BuildFakeID(),
			WaitlistSignupId: fake.BuildFakeID(),
		})
		assertCode(t, err, codes.PermissionDenied)
		assert.Empty(t, store.InviteCalls())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).InviteWaitlistSignup(t.Context(), &waitlistssvc.InviteWaitlistSignupRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_ConvertWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("marks the invitation taken up", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(fake.BuildFakeID())

		converted := *example
		converted.Status = waitlists.StatusConverted

		store := &waitlistsmock.StoreMock{
			ConvertFunc: func(context.Context, tenancy.Scope, string, string) error {
				return nil
			},
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return &converted, nil
			},
		}

		res, err := buildTestService(t, store).ConvertWaitlistSignup(who.ctx, &waitlistssvc.ConvertWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, waitlists.StatusConverted.String(), res.GetUpdated().GetStatus())
	})

	t.Run("refuses an ordinary user before it writes anything", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		store := &waitlistsmock.StoreMock{}

		_, err := buildTestService(t, store).ConvertWaitlistSignup(who.ctx, &waitlistssvc.ConvertWaitlistSignupRequest{
			WaitlistId:       fake.BuildFakeID(),
			WaitlistSignupId: fake.BuildFakeID(),
		})
		assertCode(t, err, codes.PermissionDenied)
		assert.Empty(t, store.ConvertCalls())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).ConvertWaitlistSignup(t.Context(), &waitlistssvc.ConvertWaitlistSignupRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

// TestServiceImpl_WithdrawFromWaitlist is the opt-out, which is what this
// adoption was for. The local domain had no equivalent at all.
func TestServiceImpl_WithdrawFromWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("takes the caller off the list and hands back the blanked row", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(who.userID)

		withdrawn := *example
		withdrawn.Status = waitlists.StatusWithdrawn
		withdrawn.Contact = ""
		withdrawn.Notes = ""
		withdrawn.Subject = waitlists.Subject{}

		var read int
		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				read++
				if read == 1 {
					return example, nil
				}

				return &withdrawn, nil
			},
			WithdrawFunc: func(_ context.Context, scope tenancy.Scope, listID, signupID string) error {
				assert.Equal(t, ddbwaitlists.Scope(), scope)
				assert.Equal(t, example.ListID, listID)
				assert.Equal(t, example.ID, signupID)

				return nil
			},
		}

		res, err := buildTestService(t, store).WithdrawFromWaitlist(who.ctx, &waitlistssvc.WithdrawFromWaitlistRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		require.NoError(t, err)

		// The response is what an unsubscribe page renders: off the list, and
		// holding nothing about the person any more.
		assert.Equal(t, waitlists.StatusWithdrawn.String(), res.GetUpdated().GetStatus())
		assert.Empty(t, res.GetUpdated().GetContact())
		assert.Empty(t, res.GetUpdated().GetSubjectId())
	})

	t.Run("reports a second withdrawal rather than restamping it", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(who.userID)

		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
			WithdrawFunc: func(context.Context, tenancy.Scope, string, string) error {
				return waitlists.ErrAlreadyWithdrawn
			},
		}

		_, err := buildTestService(t, store).WithdrawFromWaitlist(who.ctx, &waitlistssvc.WithdrawFromWaitlistRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		assertCode(t, err, codes.FailedPrecondition)
	})

	t.Run("refuses to withdraw somebody else", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(fake.BuildFakeID())

		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
		}

		_, err := buildTestService(t, store).WithdrawFromWaitlist(who.ctx, &waitlistssvc.WithdrawFromWaitlistRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		assertCode(t, err, codes.PermissionDenied)
		assert.Empty(t, store.WithdrawCalls())
	})

	// Somebody who has asked to come off a list by another channel has still
	// asked, so an operator may do it for them.
	t.Run("lets a service admin withdraw on somebody's behalf", func(t *testing.T) {
		t.Parallel()

		who := adminContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(fake.BuildFakeID())

		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
			WithdrawFunc: func(context.Context, tenancy.Scope, string, string) error {
				return nil
			},
		}

		_, err := buildTestService(t, store).WithdrawFromWaitlist(who.ctx, &waitlistssvc.WithdrawFromWaitlistRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		require.NoError(t, err)
		assert.Len(t, store.WithdrawCalls(), 1)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).WithdrawFromWaitlist(t.Context(), &waitlistssvc.WithdrawFromWaitlistRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_ArchiveWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("retires the caller's own signup", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(who.userID)

		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
			ArchiveSignupFunc: func(_ context.Context, scope tenancy.Scope, listID, signupID string) error {
				assert.Equal(t, ddbwaitlists.Scope(), scope)
				assert.Equal(t, example.ListID, listID)
				assert.Equal(t, example.ID, signupID)

				return nil
			},
		}

		res, err := buildTestService(t, store).ArchiveWaitlistSignup(who.ctx, &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, res.GetResponseDetails())
	})

	t.Run("refuses somebody else's signup", func(t *testing.T) {
		t.Parallel()

		who := userContextForTest(t)
		example := waitlistfakes.BuildFakeWaitlistSignupForUser(fake.BuildFakeID())

		store := &waitlistsmock.StoreMock{
			GetSignupFunc: func(context.Context, tenancy.Scope, string, string) (*waitlists.Signup, error) {
				return example, nil
			},
		}

		_, err := buildTestService(t, store).ArchiveWaitlistSignup(who.ctx, &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistId:       example.ListID,
			WaitlistSignupId: example.ID,
		})
		assertCode(t, err, codes.PermissionDenied)
		assert.Empty(t, store.ArchiveSignupCalls())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil).ArchiveWaitlistSignup(t.Context(), &waitlistssvc.ArchiveWaitlistSignupRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}
