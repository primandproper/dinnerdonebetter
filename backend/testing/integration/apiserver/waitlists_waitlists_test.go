package integration

import (
	"testing"
	"time"

	waitlistfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	waitlists "github.com/primandproper/platform-go/v14/waitlists"
	waitlistspb "github.com/primandproper/platform-go/v14/waitlists/waitlistspb"
	"github.com/primandproper/primitives-go/v2/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// The waitlists surface is platform's, and three things about it differ from the local one
// these tests were written against.
//
// WaitlistIsOpen is gone and needs no successor: openness is archived_at == null &&
// closes_at > now, both of which are on the wire, so a client computes it from a read it was
// already making. The cases that called it check ListOpenLists instead, which is the read a
// signup page makes anyway.
//
// A signup's notes are no longer part of joining. Join takes a list and an address; notes are
// an operator's annotation set afterwards through UpdateSignupNotes, which is the honest
// shape — the person joining never wrote them.
//
// And a status is an enum rather than a string.

func checkWaitlistEquality(t *testing.T, expected *waitlistspb.WaitlistInput, actual *waitlistspb.Waitlist) {
	t.Helper()

	assert.NotEmpty(t, actual.GetId(), "expected Waitlist to have ID")
	assert.NotNil(t, actual.GetCreatedAt(), "expected Waitlist to have CreatedAt")

	assert.Equal(t, expected.GetName(), actual.GetName(), "expected Waitlist Name")
	assert.Equal(t, expected.GetDescription(), actual.GetDescription(), "expected Waitlist Description")
	assert.WithinDuration(t, expected.GetClosesAt().AsTime(), actual.GetClosesAt().AsTime(), time.Second, "expected Waitlist ClosesAt")
}

// waitlistInputForTest builds a list that is still taking signups.
func waitlistInputForTest() *waitlistspb.WaitlistInput {
	example := waitlistfakes.BuildFakeWaitlist()

	return &waitlistspb.WaitlistInput{
		Name:        example.Name,
		Description: example.Description,
		ClosesAt:    timestamppb.New(example.ClosesAt),
	}
}

// createWaitlistForTest opens a waitlist. Lists are administrative rows in one
// global catalog, so it is always the admin client that opens one.
func createWaitlistForTest(t *testing.T, testClient client.Client) *waitlistspb.Waitlist {
	t.Helper()
	ctx := t.Context()

	input := waitlistInputForTest()

	created, err := adminClient.CreateList(ctx, &waitlistspb.CreateListRequest{List: input})
	require.NoError(t, err)
	checkWaitlistEquality(t, input, created.GetResult())

	retrieved, err := testClient.GetList(ctx, &waitlistspb.GetListRequest{ListId: created.GetResult().GetId()})
	require.NoError(t, err)
	require.NotNil(t, retrieved.GetResult())
	checkWaitlistEquality(t, input, retrieved.GetResult())

	return retrieved.GetResult()
}

// createWaitlistSignupForTest joins a list as testClient's user.
//
// A signup belongs to the person who made it and carries the address off their
// session, so the client that calls this is the one that may read, amend and
// withdraw it afterwards — and one client can join a given list exactly once,
// which is the uniqueness the withdrawal rests on.
//
// The contact sent here is deliberately not the session's, and the assertion below is that
// the server ignored it. platform's Join takes the address from the request by default —
// right for the public form it is written for — and this deployment supplies a
// waitlistsgrpc.ContactResolver, because Join is behind a grant here and a stated address
// would let any authenticated caller sign somebody else up. See
// internal/build/waitlists.ownContact.
func createWaitlistSignupForTest(t *testing.T, testClient client.Client, waitlistID string) *waitlistspb.Signup {
	t.Helper()
	ctx := t.Context()

	statedContact := identifiers.New() + "@stated.example.com"

	_, err := testClient.Join(ctx, &waitlistspb.JoinRequest{
		ListId:  waitlistID,
		Contact: statedContact,
	})
	require.NoError(t, err)

	// Join answers with nothing, deliberately: it is a public RPC upstream, and a response
	// echoing the signup would confirm to an anonymous caller that a given address is on a
	// list. So the signup is read back, through the caller's own subject.
	signup := signupForSubject(t, testClient, waitlistID)

	assert.Equal(t, waitlistID, signup.GetListId(), "expected Signup ListID")
	assert.Equal(t, waitlistspb.SignupStatus_SIGNUP_STATUS_WAITING, signup.GetStatus(), "a signup is born waiting")

	// The address and the subject are the session's, never the request's.
	assert.NotEmpty(t, signup.GetContact(), "expected Signup to carry the session's address")
	assert.NotEqual(t, statedContact, signup.GetContact(), "the request's address was stored instead of the session's")
	assert.Equal(t, string(waitlists.SubjectUser), signup.GetSubject().GetType())
	assert.NotEmpty(t, signup.GetSubject().GetId(), "expected Signup to name its subject")

	retrieved, err := adminClient.GetSignup(ctx, &waitlistspb.GetSignupRequest{
		ListId:   waitlistID,
		SignupId: signup.GetId(),
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved.GetResult())

	assert.Equal(t, signup.GetId(), retrieved.GetResult().GetId())
	assert.Equal(t, signup.GetContact(), retrieved.GetResult().GetContact())
	assert.Equal(t, signup.GetSubject().GetId(), retrieved.GetResult().GetSubject().GetId())

	return retrieved.GetResult()
}

// signupForSubject finds the caller's own signup on a list.
//
// The read a client makes after joining, since Join answers with nothing. The caller makes
// it themselves: of the four signup reads platform puts behind one grant, this is the one a
// member holds, under ReadOwnWaitlistSignupsPermission, and the handler refuses a subject
// that is not the caller's own. The other three stay a service admin's.
func signupForSubject(t *testing.T, testClient client.Client, waitlistID string) *waitlistspb.Signup {
	t.Helper()

	authStatus, err := testClient.GetAuthStatus(t.Context(), &authsvc.GetAuthStatusRequest{})
	require.NoError(t, err)

	mine, err := testClient.ListSignupsForSubject(t.Context(), &waitlistspb.ListSignupsForSubjectRequest{
		Subject: &waitlistspb.SignupSubject{Type: string(waitlists.SubjectUser), Id: authStatus.GetUserId()},
	})
	require.NoError(t, err)

	for _, signup := range mine.GetResults() {
		if signup.GetListId() == waitlistID {
			return signup
		}
	}

	require.FailNow(t, "the caller has no signup on waitlist "+waitlistID)

	return nil
}

// listIsOpen answers what WaitlistIsOpen used to, from the read a client already makes.
func listIsOpen(t *testing.T, testClient client.Client, listID string) bool {
	t.Helper()

	open, err := testClient.ListOpenLists(t.Context(), &waitlistspb.ListOpenListsRequest{})
	require.NoError(t, err)

	for _, list := range open.GetResults() {
		if list.GetId() == listID {
			return true
		}
	}

	return false
}

func TestWaitlists_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, testClient := createUserAndClientForTest(t)
		createWaitlistForTest(t, testClient)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateList(ctx, &waitlistspb.CreateListRequest{})
		require.Error(t, err)
	})

	// A list with no closing time is refused rather than given one: there is no default for
	// "when does this stop taking signups", and guessing one closes somebody's list early or
	// never.
	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.CreateList(ctx, &waitlistspb.CreateListRequest{
			List: &waitlistspb.WaitlistInput{Description: t.Name()},
		})
		require.Error(t, err)
	})
}

func TestWaitlists_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWaitlistForTest(t, testClient)

		retrieved, err := testClient.GetList(ctx, &waitlistspb.GetListRequest{ListId: created.GetId()})
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetList(ctx, &waitlistspb.GetListRequest{ListId: nonexistentID})
		require.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetList(ctx, &waitlistspb.GetListRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		created := []*waitlistspb.Waitlist{}
		for range exampleQuantity {
			created = append(created, createWaitlistForTest(t, testClient))
		}

		results, err := testClient.ListLists(ctx, &waitlistspb.ListListsRequest{})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), len(created))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ListLists(ctx, &waitlistspb.ListListsRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_ListingOpen(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		created := []*waitlistspb.Waitlist{}
		for range exampleQuantity {
			created = append(created, createWaitlistForTest(t, testClient))
		}

		results, err := testClient.ListOpenLists(ctx, &waitlistspb.ListOpenListsRequest{})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), len(created))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ListOpenLists(ctx, &waitlistspb.ListOpenListsRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_Updating(T *testing.T) {
	T.Parallel()

	// UpdateList replaces rather than merges: the input is a whole list, not a patch of one,
	// so a caller restates the fields it is keeping. The local RPC took *string fields and
	// left an omission alone.
	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWaitlistForTest(t, testClient)

		const newName = "renamed"

		_, err := adminClient.UpdateList(ctx, &waitlistspb.UpdateListRequest{
			List: &waitlistspb.WaitlistInput{
				Id:          created.GetId(),
				Name:        newName,
				Description: created.GetDescription(),
				ClosesAt:    created.GetClosesAt(),
			},
		})
		require.NoError(t, err)

		retrieved, err := testClient.GetList(ctx, &waitlistspb.GetListRequest{ListId: created.GetId()})
		require.NoError(t, err)
		assert.Equal(t, newName, retrieved.GetResult().GetName())
		assert.Equal(t, created.GetDescription(), retrieved.GetResult().GetDescription())
		assert.NotNil(t, retrieved.GetResult().GetLastUpdatedAt())
	})

	// Moving the closing time into the past is how a list is closed early, and it is not
	// guarded against the signups already on it: they keep their places.
	T.Run("closing a list early", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWaitlistForTest(t, testClient)
		require.True(t, listIsOpen(t, testClient, created.GetId()))

		_, err := adminClient.UpdateList(ctx, &waitlistspb.UpdateListRequest{
			List: &waitlistspb.WaitlistInput{
				Id:          created.GetId(),
				Name:        created.GetName(),
				Description: created.GetDescription(),
				ClosesAt:    timestamppb.New(time.Now().Add(-time.Hour).UTC()),
			},
		})
		require.NoError(t, err)

		// Off the open page, still in the catalog.
		assert.False(t, listIsOpen(t, testClient, created.GetId()))

		stillListed, err := testClient.GetList(ctx, &waitlistspb.GetListRequest{ListId: created.GetId()})
		require.NoError(t, err)
		assert.NotNil(t, stillListed.GetResult())
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.UpdateList(ctx, &waitlistspb.UpdateListRequest{
			List: &waitlistspb.WaitlistInput{
				Id:       nonexistentID,
				Name:     t.Name(),
				ClosesAt: timestamppb.New(time.Now().Add(time.Hour).UTC()),
			},
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateList(ctx, &waitlistspb.UpdateListRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWaitlistForTest(t, testClient)

		_, err := adminClient.ArchiveList(ctx, &waitlistspb.ArchiveListRequest{ListId: created.GetId()})
		require.NoError(t, err)

		_, err = testClient.GetList(ctx, &waitlistspb.GetListRequest{ListId: created.GetId()})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		// Archiving closes a list whatever its closing time says.
		assert.False(t, listIsOpen(t, testClient, created.GetId()))
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveList(ctx, &waitlistspb.ArchiveListRequest{ListId: nonexistentID})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveList(ctx, &waitlistspb.ArchiveListRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlistSignups_Joining(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		createWaitlistSignupForTest(t, testClient, waitlist.GetId())
	})

	// One address, one place on a list. This is the uniqueness the withdrawal rests on,
	// and it is enforced quietly: a second join answers success and writes nothing, because
	// a refusal a caller could see would tell whoever named an address whether it was
	// already on the list. The outcome goes on the operation instead. So the assertion is
	// that the queue did not grow, not that the call failed.
	T.Run("a second signup from the same person adds nothing", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		first := createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		_, err := testClient.Join(ctx, &waitlistspb.JoinRequest{ListId: waitlist.GetId()})
		require.NoError(t, err, "a duplicate join is answered rather than refused")

		mine, err := testClient.ListSignupsForSubject(ctx, subjectRequestFor(t, testClient))
		require.NoError(t, err)

		var onThisList int
		for _, signup := range mine.GetResults() {
			if signup.GetListId() == waitlist.GetId() {
				onThisList++
				assert.Equal(t, first.GetId(), signup.GetId())
			}
		}
		assert.Equal(t, 1, onThisList, "a duplicate join wrote a second place on the list")
	})

	// The narrowing, from the other side: two callers stating the same address are still two
	// people, because neither of their signups carries the address they stated.
	T.Run("a stated address cannot impersonate another caller", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		_, otherClient := createUserAndClientForTest(t)

		waitlist := createWaitlistForTest(t, testClient)
		mine := createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		// Somebody else joining while claiming my address joins as themselves.
		_, err := otherClient.Join(ctx, &waitlistspb.JoinRequest{
			ListId:  waitlist.GetId(),
			Contact: mine.GetContact(),
		})
		require.NoError(t, err)

		theirs := signupForSubject(t, otherClient, waitlist.GetId())
		assert.NotEqual(t, mine.GetContact(), theirs.GetContact())
		assert.NotEqual(t, mine.GetSubject().GetId(), theirs.GetSubject().GetId())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.Join(ctx, &waitlistspb.JoinRequest{})
		require.Error(t, err)
	})

	T.Run("nonexistent waitlist ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.Join(ctx, &waitlistspb.JoinRequest{ListId: nonexistentID})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestWaitlistSignups_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		created := createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		retrieved, err := adminClient.GetSignup(ctx, &waitlistspb.GetSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: created.GetId(),
		})
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)

		retrieved, err := adminClient.GetSignup(ctx, &waitlistspb.GetSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetSignup(ctx, &waitlistspb.GetSignupRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlistSignups_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)

		// One signup per person: an address can hold one place on a list, so a
		// queue of five is five people.
		created := []*waitlistspb.Signup{createWaitlistSignupForTest(t, testClient, waitlist.GetId())}
		for range exampleQuantity - 1 {
			_, joiner := createUserAndClientForTest(t)
			created = append(created, createWaitlistSignupForTest(t, joiner, waitlist.GetId()))
		}

		// the waitlist-wide signup listing is reserved for service admins: it hands
		// back every signatory's address.
		results, err := adminClient.ListSignups(ctx, &waitlistspb.ListSignupsRequest{ListId: waitlist.GetId()})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), len(created))
	})

	T.Run("denied for regular users", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		_, err := testClient.ListSignups(ctx, &waitlistspb.ListSignupsRequest{ListId: waitlist.GetId()})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ListSignups(ctx, &waitlistspb.ListSignupsRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlistSignups_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		created := createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		const newNotes = "Updated notes"

		// The admin client, because the note is the operator's: platform defines this
		// grant as "rewriting the operator's note against a signup", and neither it nor
		// ArchiveSignup has an ownership seam, so a member holding either could rewrite
		// or retire anybody's signup by naming its id. See internal/authorization.
		_, err := adminClient.UpdateSignupNotes(ctx, &waitlistspb.UpdateSignupNotesRequest{
			ListId:   waitlist.GetId(),
			SignupId: created.GetId(),
			Notes:    newNotes,
		})
		require.NoError(t, err)

		retrieved, err := adminClient.GetSignup(ctx, &waitlistspb.GetSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: created.GetId(),
		})
		require.NoError(t, err)
		assert.Equal(t, newNotes, retrieved.GetResult().GetNotes())

		// A note moves nobody: the signup is still waiting and its lifecycle stamp
		// is still unset.
		assert.Equal(t, waitlistspb.SignupStatus_SIGNUP_STATUS_WAITING, retrieved.GetResult().GetStatus())
		assert.Nil(t, retrieved.GetResult().GetStatusChangedAt())
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)

		_, err := adminClient.UpdateSignupNotes(ctx, &waitlistspb.UpdateSignupNotesRequest{
			ListId:   waitlist.GetId(),
			SignupId: nonexistentID,
			Notes:    "Updated notes",
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateSignupNotes(ctx, &waitlistspb.UpdateSignupNotesRequest{})
		assert.Error(t, err)
	})
}

// TestWaitlistSignups_Lifecycle walks the queue end to end: waiting, invited,
// converted — with the second invitation refused, which is what makes an
// invitation email go out once.
func TestWaitlistSignups_Lifecycle(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		signup := createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		invited, err := adminClient.Invite(ctx, &waitlistspb.InviteRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.NoError(t, err)
		assert.Equal(t, waitlistspb.SignupStatus_SIGNUP_STATUS_INVITED, invited.GetResult().GetStatus())
		assert.NotNil(t, invited.GetResult().GetStatusChangedAt())

		_, err = adminClient.Invite(ctx, &waitlistspb.InviteRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))

		converted, err := adminClient.Convert(ctx, &waitlistspb.ConvertRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.NoError(t, err)
		assert.Equal(t, waitlistspb.SignupStatus_SIGNUP_STATUS_CONVERTED, converted.GetResult().GetStatus())
	})

	T.Run("denied for regular users", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		signup := createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		// Being on a list does not entitle somebody to invite themselves off it.
		_, err := testClient.Invite(ctx, &waitlistspb.InviteRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		_, err = testClient.Convert(ctx, &waitlistspb.ConvertRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.Invite(ctx, &waitlistspb.InviteRequest{})
		require.Error(t, err)

		_, err = c.Convert(ctx, &waitlistspb.ConvertRequest{})
		require.Error(t, err)
	})
}

// TestWaitlistSignups_Withdrawing is the opt-out this adoption was for: a
// suppression that outlives the address it is about.
func TestWaitlistSignups_Withdrawing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		signup := createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		_, err := testClient.Withdraw(ctx, &waitlistspb.WithdrawRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.NoError(t, err)

		// Filling the form in again does not put them back on the list. This is the whole
		// obligation: the local table it replaced had no way to express it.
		//
		// It is honored quietly. The call answers success and writes nothing, because a
		// visible refusal would tell whoever named the address that somebody with it had
		// asked to be left alone — which is the one fact a withdrawal exists to keep. So
		// the assertion is that they are still withdrawn, not that the call failed.
		_, err = testClient.Join(ctx, &waitlistspb.JoinRequest{ListId: waitlist.GetId()})
		require.NoError(t, err, "a suppressed join is answered rather than refused")

		stillWithdrawn, err := adminClient.GetSignup(ctx, &waitlistspb.GetSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.NoError(t, err)
		assert.Equal(t, waitlistspb.SignupStatus_SIGNUP_STATUS_WITHDRAWN, stillWithdrawn.GetResult().GetStatus())

		// A second withdrawal is refused, and the code is worth pinning because it is
		// the anonymization rather than the lifecycle guard that refuses it: the row
		// no longer names anybody, so the service can no longer tell that this caller
		// is the person it used to be about. The store's own answer to a replayed
		// withdrawal is ErrAlreadyWithdrawn (see the repository suite); nothing gets
		// that far from here, and that is the design working rather than around it.
		_, err = testClient.Withdraw(ctx, &waitlistspb.WithdrawRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// One person's withdrawal does not keep another person off the list.
	//
	// The suppression is on the address, so without the contact resolver somebody who typed
	// a withdrawn address would be suppressed by it — a stranger's opt-out becoming a denial
	// of service against them. With the resolver the stated address does not matter: they
	// join as themselves.
	//
	// This is not about disclosure. Whether an address has withdrawn is already unlearnable
	// from the reply: platform answers a suppressed join as success and records the outcome
	// on the operation. See waitlists/grpc.quietJoinOutcome.
	T.Run("another person's opt-out does not keep this caller off the list", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		_, otherClient := createUserAndClientForTest(t)

		waitlist := createWaitlistForTest(t, testClient)
		signup := createWaitlistSignupForTest(t, testClient, waitlist.GetId())
		withdrawnContact := signup.GetContact()

		_, err := testClient.Withdraw(ctx, &waitlistspb.WithdrawRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.NoError(t, err)

		// Joining while claiming the withdrawn address puts them on the list under their
		// own, which is the resolver working: the suppression is about an address nobody
		// here is.
		_, err = otherClient.Join(ctx, &waitlistspb.JoinRequest{
			ListId:  waitlist.GetId(),
			Contact: withdrawnContact,
		})
		require.NoError(t, err)

		joined := signupForSubject(t, otherClient, waitlist.GetId())
		assert.NotEqual(t, withdrawnContact, joined.GetContact())
		assert.Equal(t, waitlistspb.SignupStatus_SIGNUP_STATUS_WAITING, joined.GetStatus(),
			"a stranger's withdrawal suppressed this caller")
	})

	T.Run("denied for somebody else's signup", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		_, otherClient := createUserAndClientForTest(t)

		waitlist := createWaitlistForTest(t, testClient)
		signup := createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		// NotFound rather than PermissionDenied, and platform is emphatic that this is
		// not a rounding of one to the other. The caller here is frequently anonymous and
		// holds an identifier somebody handed them; a PermissionDenied on a signup that
		// is not theirs and a NotFound on one that does not exist would be two answers a
		// caller walking identifiers could tell apart, which is the enumeration this
		// service refuses to be.
		_, err := otherClient.Withdraw(ctx, &waitlistspb.WithdrawRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.Withdraw(ctx, &waitlistspb.WithdrawRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlistSignups_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		created := createWaitlistSignupForTest(t, testClient, waitlist.GetId())

		// The admin client: archiving is the queue being tidied, not an opt-out being
		// honored, and platform puts it behind an operator's grant for that reason.
		_, err := adminClient.ArchiveSignup(ctx, &waitlistspb.ArchiveSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: created.GetId(),
		})
		require.NoError(t, err)

		// Archiving is not withdrawing: the row is hidden and the address is still stored,
		// so a second attempt is a duplicate rather than an honored opt-out. Both are quiet,
		// so what separates them here is that no new place appears either way — the
		// difference between the two lives on the operation, not in the reply.
		_, err = testClient.Join(ctx, &waitlistspb.JoinRequest{ListId: waitlist.GetId()})
		require.NoError(t, err, "a duplicate join is answered rather than refused")

		mine, err := testClient.ListSignupsForSubject(ctx, subjectRequestFor(t, testClient))
		require.NoError(t, err)
		for _, signup := range mine.GetResults() {
			assert.NotEqual(t, waitlist.GetId(), signup.GetListId(), "an archived signup came back")
		}
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)

		_, err := adminClient.ArchiveSignup(ctx, &waitlistspb.ArchiveSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: nonexistentID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveSignup(ctx, &waitlistspb.ArchiveSignupRequest{})
		assert.Error(t, err)
	})
}

// subjectRequestFor builds the caller's own subject read.
func subjectRequestFor(t *testing.T, testClient client.Client) *waitlistspb.ListSignupsForSubjectRequest {
	t.Helper()

	authStatus, err := testClient.GetAuthStatus(t.Context(), &authsvc.GetAuthStatusRequest{})
	require.NoError(t, err)

	return &waitlistspb.ListSignupsForSubjectRequest{
		Subject: &waitlistspb.SignupSubject{Type: string(waitlists.SubjectUser), Id: authStatus.GetUserId()},
	}
}
