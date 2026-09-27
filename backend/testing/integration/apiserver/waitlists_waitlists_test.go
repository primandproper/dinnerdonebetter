package integration

import (
	"testing"

	waitlistfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	waitlists "github.com/primandproper/platform-go/v14/waitlists"
	waitlistspb "github.com/primandproper/platform-go/v14/waitlists/waitlistspb"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// The waitlists surface is platform's, and conformance/waitlists asserts it: the catalog, the
// public signup page and its double opt-in, the queue's lifecycle, withdrawal and erasure. What
// is left here is what only this deployment can say — who among its roles may do what, the
// confirmation mail it actually queues, and the audit entries its repository records.

// createWaitlistForTest opens a list. Lists are administrative rows in one global catalog, so
// the admin client opens every one.
func createWaitlistForTest(t *testing.T) *waitlistspb.Waitlist {
	t.Helper()

	example := waitlistfakes.BuildFakeWaitlist()

	created, err := adminClient.CreateList(t.Context(), &waitlistspb.CreateListRequest{List: &waitlistspb.WaitlistInput{
		Name:        example.Name,
		Description: example.Description,
		ClosesAt:    timestamppb.New(example.ClosesAt),
	}})
	require.NoError(t, err)

	return created.GetResult()
}

// joinAndConfirmForTest joins a list as testClient's user and follows the confirmation link
// the join mailed, the way the person would: from their inbox, signed in or not.
func joinAndConfirmForTest(t *testing.T, testClient client.Client, listID string) *waitlistspb.Signup {
	t.Helper()
	ctx := t.Context()

	contact := identifiers.New() + "@example.invalid"

	_, err := testClient.Join(ctx, &waitlistspb.JoinRequest{ListId: listID, Contact: contact})
	require.NoError(t, err)

	links, err := conformanceWaitlistLinks(ctx, tenancy.Global(), listID, contact)
	require.NoError(t, err, "the join queued no confirmation mail")

	_, err = buildUnauthenticatedGRPCClientForTest(t).Confirm(ctx, &waitlistspb.ConfirmRequest{Token: links.Confirm})
	require.NoError(t, err)

	return signupForSubject(t, testClient, listID)
}

// signupForSubject finds the caller's own signup on a list, through the one signup read a
// member holds.
func signupForSubject(t *testing.T, testClient client.Client, listID string) *waitlistspb.Signup {
	t.Helper()

	mine, err := testClient.ListSignupsForSubject(t.Context(), subjectRequestFor(t, testClient))
	require.NoError(t, err)

	for _, signup := range mine.GetResults() {
		if signup.GetListId() == listID {
			return signup
		}
	}

	require.FailNow(t, "the caller has no signup on waitlist "+listID)

	return nil
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

// TestWaitlistSignups_Confirmation pins this deployment's half of the double opt-in: the mail it
// queues on the outbound-emails topic carries links that work, a signed-in caller's join is
// attributed to them while it waits on its link, and the repository records the join and the
// confirmation as it records every other signup write.
func TestWaitlistSignups_Confirmation(T *testing.T) {
	T.Parallel()

	T.Run("a signed-in caller's join is held pending until its mailed link is followed", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t)
		contact := identifiers.New() + "@example.invalid"

		_, err := testClient.Join(ctx, &waitlistspb.JoinRequest{ListId: waitlist.GetId(), Contact: contact})
		require.NoError(t, err)

		pending := signupForSubject(t, testClient, waitlist.GetId())
		assert.Equal(t, waitlistspb.SignupStatus_SIGNUP_STATUS_PENDING, pending.GetStatus())
		assert.Equal(t, contact, pending.GetContact(), "the address stated is the one the mail went to")

		links, err := conformanceWaitlistLinks(ctx, tenancy.Global(), waitlist.GetId(), contact)
		require.NoError(t, err, "the join queued no confirmation mail")
		require.NotEmpty(t, links.Unsubscribe, "the confirmation mail carries no way off the list")

		_, err = buildUnauthenticatedGRPCClientForTest(t).Confirm(ctx, &waitlistspb.ConfirmRequest{Token: links.Confirm})
		require.NoError(t, err)

		confirmed := signupForSubject(t, testClient, waitlist.GetId())
		assert.Equal(t, pending.GetId(), confirmed.GetId())
		assert.Equal(t, waitlistspb.SignupStatus_SIGNUP_STATUS_WAITING, confirmed.GetStatus())

		AssertAuditLogContainsFuzzyForResource(t, ctx, "waitlist_signups", confirmed.GetId(), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "waitlist_signups", RelevantID: confirmed.GetId()},
			{EventType: "updated", ResourceType: "waitlist_signups", RelevantID: confirmed.GetId()},
		})
	})
}

// TestWaitlists_ThisDeploymentsAuthorization pins how this deployment divides the surface among
// its roles. The suites make each reserved call as an operator and never ask whether an ordinary
// member is refused it, and the three authorizers this deployment supplies are its own rules.
//
// Signups are global rows carrying the address the list writes to, so the listing is a read of
// every signatory's email and the operator's grant is what keeps it private.
func TestWaitlists_ThisDeploymentsAuthorization(T *testing.T) {
	T.Parallel()

	T.Run("a member may not work a list or its queue", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, owner := createUserAndClientForTest(t)
		_, member := createUserAndClientForTest(t)

		waitlist := createWaitlistForTest(t)
		signup := joinAndConfirmForTest(t, owner, waitlist.GetId())
		list, signupID := waitlist.GetId(), signup.GetId()

		refusals := map[string]func() error{
			"CreateList": func() error {
				_, err := member.CreateList(ctx, &waitlistspb.CreateListRequest{List: &waitlistspb.WaitlistInput{Name: t.Name()}})
				return err
			},
			"UpdateList": func() error {
				_, err := member.UpdateList(ctx, &waitlistspb.UpdateListRequest{List: &waitlistspb.WaitlistInput{Id: list, Name: t.Name()}})
				return err
			},
			"ArchiveList": func() error {
				_, err := member.ArchiveList(ctx, &waitlistspb.ArchiveListRequest{ListId: list})
				return err
			},
			"GetSignup": func() error {
				_, err := member.GetSignup(ctx, &waitlistspb.GetSignupRequest{ListId: list, SignupId: signupID})
				return err
			},
			"ListSignups": func() error {
				_, err := member.ListSignups(ctx, &waitlistspb.ListSignupsRequest{ListId: list})
				return err
			},
			"UpdateSignupNotes": func() error {
				_, err := member.UpdateSignupNotes(ctx, &waitlistspb.UpdateSignupNotesRequest{ListId: list, SignupId: signupID, Notes: t.Name()})
				return err
			},
			"Invite": func() error {
				_, err := member.Invite(ctx, &waitlistspb.InviteRequest{ListId: list, SignupId: signupID})
				return err
			},
			"Convert": func() error {
				_, err := member.Convert(ctx, &waitlistspb.ConvertRequest{ListId: list, SignupId: signupID})
				return err
			},
			"ArchiveSignup": func() error {
				_, err := member.ArchiveSignup(ctx, &waitlistspb.ArchiveSignupRequest{ListId: list, SignupId: signupID})
				return err
			},
		}

		for name, call := range refusals {
			assert.Equal(t, codes.PermissionDenied, status.Code(call()), "a member was not refused %s", name)
		}

		// Including the person on the list: a place in the queue is not a seat at its desk.
		_, err := owner.GetSignup(ctx, &waitlistspb.GetSignupRequest{ListId: list, SignupId: signupID})
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	// ListSignupsForSubject is a member's under ReadOwnWaitlistSignupsPermission, and the grant
	// cannot say whose: the subject comes off the request. The SubjectRead authorizer does,
	// before any row is read, and answers NotFound so that somebody else's subject and nobody's
	// are the same answer.
	T.Run("a member reads their own signups and nobody else's", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, owner := createUserAndClientForTest(t)
		_, member := createUserAndClientForTest(t)

		waitlist := createWaitlistForTest(t)
		joinAndConfirmForTest(t, owner, waitlist.GetId())

		own, err := owner.ListSignupsForSubject(ctx, subjectRequestFor(t, owner))
		require.NoError(t, err)
		assert.NotEmpty(t, own.GetResults())

		_, err = member.ListSignupsForSubject(ctx, subjectRequestFor(t, owner))
		assert.Equal(t, codes.NotFound, status.Code(err))

		theirs, err := adminClient.ListSignupsForSubject(ctx, subjectRequestFor(t, owner))
		require.NoError(t, err)
		assert.NotEmpty(t, theirs.GetResults())
	})

	// Withdraw by signup id is the signed-in door: its own signup's person or a service admin.
	// Somebody holding only an identifier — signed in or not — is refused as though it named
	// nothing; the mailed unsubscribe link is the anonymous door, and it is Unsubscribe's.
	T.Run("a withdrawal by id is the signup's own person's", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, owner := createUserAndClientForTest(t)
		_, member := createUserAndClientForTest(t)

		waitlist := createWaitlistForTest(t)
		signup := joinAndConfirmForTest(t, owner, waitlist.GetId())
		request := &waitlistspb.WithdrawRequest{ListId: waitlist.GetId(), SignupId: signup.GetId()}

		_, err := member.Withdraw(ctx, request)
		assert.Equal(t, codes.NotFound, status.Code(err))

		_, err = buildUnauthenticatedGRPCClientForTest(t).Withdraw(ctx, request)
		assert.Equal(t, codes.NotFound, status.Code(err))

		_, err = owner.Withdraw(ctx, request)
		require.NoError(t, err)

		AssertAuditLogContainsFuzzyForResource(t, ctx, "waitlist_signups", signup.GetId(), 10, []*ExpectedAuditEntry{
			{EventType: "updated", ResourceType: "waitlist_signups", RelevantID: signup.GetId()},
		})
	})
}
