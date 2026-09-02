package integration

import (
	"testing"
	"time"

	waitlistfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/waitlists/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	waitlists "github.com/primandproper/platform-go/v13/waitlists"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func checkWaitlistEquality(t *testing.T, expected, actual *waitlists.List) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected Waitlist to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected Waitlist to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected Waitlist Name")
	assert.Equal(t, expected.Description, actual.Description, "expected Waitlist Description")
	assert.WithinDuration(t, expected.ClosesAt, actual.ClosesAt, time.Second, "expected Waitlist ClosesAt")
}

// createWaitlistForTest opens a waitlist. Lists are administrative rows in one
// global catalog, so it is always the admin client that opens one.
func createWaitlistForTest(t *testing.T, testClient client.Client) *waitlists.List {
	t.Helper()
	ctx := t.Context()

	exampleWaitlist := waitlistfakes.BuildFakeWaitlist()

	createdWaitlist, err := adminClient.CreateWaitlist(ctx, &waitlistssvc.CreateWaitlistRequest{
		Input: grpcconverters.ConvertWaitlistToGRPCWaitlistCreationRequestInput(exampleWaitlist),
	})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCWaitlistToWaitlist(createdWaitlist.GetCreated())
	checkWaitlistEquality(t, exampleWaitlist, converted)

	retrievedWaitlist, err := testClient.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: createdWaitlist.GetCreated().GetId()})
	require.NoError(t, err)
	require.NotNil(t, retrievedWaitlist)

	list := grpcconverters.ConvertGRPCWaitlistToWaitlist(retrievedWaitlist.GetResult())
	checkWaitlistEquality(t, converted, list)

	return list
}

// createWaitlistSignupForTest joins a list as testClient's user.
//
// A signup belongs to the person who made it and carries the address off their
// session, so the client that calls this is the one that may read, amend and
// withdraw it afterwards — and one client can join a given list exactly once,
// which is the uniqueness the withdrawal rests on.
func createWaitlistSignupForTest(t *testing.T, testClient client.Client, waitlistID string) *waitlists.Signup {
	t.Helper()
	ctx := t.Context()

	exampleSignup := waitlistfakes.BuildFakeWaitlistSignup()

	createdSignup, err := testClient.JoinWaitlist(ctx, &waitlistssvc.JoinWaitlistRequest{
		WaitlistId: waitlistID,
		Input:      &waitlistssvc.WaitlistSignupCreationRequestInput{Notes: exampleSignup.Notes},
	})
	require.NoError(t, err)

	converted := grpcconverters.ConvertGRPCWaitlistSignupToWaitlistSignup(createdSignup.GetCreated())
	assert.Equal(t, exampleSignup.Notes, converted.Notes, "expected WaitlistSignup Notes")
	assert.Equal(t, waitlistID, converted.ListID, "expected WaitlistSignup ListID")
	assert.Equal(t, waitlists.StatusWaiting, converted.Status, "a signup is born waiting")

	// The address and the subject are the session's, never the request's.
	assert.NotEmpty(t, converted.Contact, "expected WaitlistSignup to carry the session's address")
	assert.Equal(t, waitlists.SubjectUser, converted.Subject.Type)
	assert.NotEmpty(t, converted.Subject.ID, "expected WaitlistSignup to name its subject")

	retrievedSignup, err := testClient.GetWaitlistSignup(ctx, &waitlistssvc.GetWaitlistSignupRequest{
		WaitlistId:       waitlistID,
		WaitlistSignupId: createdSignup.GetCreated().GetId(),
	})
	require.NoError(t, err)
	require.NotNil(t, retrievedSignup)

	signup := grpcconverters.ConvertGRPCWaitlistSignupToWaitlistSignup(retrievedSignup.GetResult())
	assert.Equal(t, converted.ID, signup.ID)
	assert.Equal(t, converted.Notes, signup.Notes)
	assert.Equal(t, converted.Contact, signup.Contact)
	assert.Equal(t, converted.Subject, signup.Subject)

	return signup
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

		_, err := c.CreateWaitlist(ctx, &waitlistssvc.CreateWaitlistRequest{})
		require.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// No name and no closing time: the store refuses both, and the refusal
		// reaches the client as an invalid argument rather than as a server error.
		_, err := adminClient.CreateWaitlist(ctx, &waitlistssvc.CreateWaitlistRequest{
			Input: &waitlistssvc.WaitlistCreationRequestInput{Description: t.Name()},
		})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestWaitlists_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWaitlist := createWaitlistForTest(t, testClient)

		retrieved, err := testClient.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: createdWaitlist.ID})
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: nonexistentID})
		require.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdWaitlists := []*waitlists.List{}
		for range exampleQuantity {
			createdWaitlists = append(createdWaitlists, createWaitlistForTest(t, testClient))
		}

		results, err := testClient.GetWaitlists(ctx, &waitlistssvc.GetWaitlistsRequest{})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), len(createdWaitlists))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetWaitlists(ctx, &waitlistssvc.GetWaitlistsRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_ListingOpen(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdWaitlists := []*waitlists.List{}
		for range exampleQuantity {
			createdWaitlists = append(createdWaitlists, createWaitlistForTest(t, testClient))
		}

		results, err := testClient.GetOpenWaitlists(ctx, &waitlistssvc.GetOpenWaitlistsRequest{})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), len(createdWaitlists))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetOpenWaitlists(ctx, &waitlistssvc.GetOpenWaitlistsRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWaitlist := createWaitlistForTest(t, testClient)

		newName := "Updated Name"
		newDescription := "Updated Description"
		newClosesAt := time.Now().Add(48 * time.Hour)

		_, err := adminClient.UpdateWaitlist(ctx, &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: createdWaitlist.ID,
			Input: &waitlistssvc.WaitlistUpdateRequestInput{
				Name:        &newName,
				Description: &newDescription,
				ClosesAt:    timestamppb.New(newClosesAt),
			},
		})
		require.NoError(t, err)

		retrieved, err := testClient.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: createdWaitlist.ID})
		require.NoError(t, err)
		assert.Equal(t, newName, retrieved.GetResult().GetName())
		assert.Equal(t, newDescription, retrieved.GetResult().GetDescription())
		assert.WithinDuration(t, newClosesAt, grpcconverters.ConvertGRPCWaitlistToWaitlist(retrieved.GetResult()).ClosesAt, time.Second)
	})

	// An update naming one field leaves the others alone, because the store's
	// update takes a whole list and the service merges into the row as read.
	T.Run("leaves unnamed fields alone", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWaitlist := createWaitlistForTest(t, testClient)

		newName := "Only The Name"
		_, err := adminClient.UpdateWaitlist(ctx, &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: createdWaitlist.ID,
			Input:      &waitlistssvc.WaitlistUpdateRequestInput{Name: &newName},
		})
		require.NoError(t, err)

		retrieved, err := testClient.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: createdWaitlist.ID})
		require.NoError(t, err)
		assert.Equal(t, newName, retrieved.GetResult().GetName())
		assert.Equal(t, createdWaitlist.Description, retrieved.GetResult().GetDescription())
		assert.WithinDuration(t, createdWaitlist.ClosesAt,
			grpcconverters.ConvertGRPCWaitlistToWaitlist(retrieved.GetResult()).ClosesAt, time.Second)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		newName := "Updated Name"
		_, err := adminClient.UpdateWaitlist(ctx, &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: nonexistentID,
			Input: &waitlistssvc.WaitlistUpdateRequestInput{
				Name: &newName,
			},
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateWaitlist(ctx, &waitlistssvc.UpdateWaitlistRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWaitlist := createWaitlistForTest(t, testClient)

		_, err := adminClient.ArchiveWaitlist(ctx, &waitlistssvc.ArchiveWaitlistRequest{WaitlistId: createdWaitlist.ID})
		require.NoError(t, err)

		// Archiving closes the list immediately, whatever its closing time says.
		_, err = testClient.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: createdWaitlist.ID})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveWaitlist(ctx, &waitlistssvc.ArchiveWaitlistRequest{WaitlistId: nonexistentID})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveWaitlist(ctx, &waitlistssvc.ArchiveWaitlistRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_IsOpen(T *testing.T) {
	T.Parallel()

	T.Run("happy path - open", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWaitlist := createWaitlistForTest(t, testClient)

		result, err := testClient.WaitlistIsOpen(ctx, &waitlistssvc.WaitlistIsOpenRequest{WaitlistId: createdWaitlist.ID})
		require.NoError(t, err)
		assert.True(t, result.GetIsOpen())
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.WaitlistIsOpen(ctx, &waitlistssvc.WaitlistIsOpenRequest{WaitlistId: nonexistentID})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WaitlistIsOpen(ctx, &waitlistssvc.WaitlistIsOpenRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlistSignups_Joining(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		createWaitlistSignupForTest(t, testClient, waitlist.ID)
	})

	// One address, one place on a list. This is the uniqueness the withdrawal
	// rests on, so it is worth pinning from the outside.
	T.Run("refuses a second signup from the same person", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		createWaitlistSignupForTest(t, testClient, waitlist.ID)

		_, err := testClient.JoinWaitlist(ctx, &waitlistssvc.JoinWaitlistRequest{
			WaitlistId: waitlist.ID,
			Input:      &waitlistssvc.WaitlistSignupCreationRequestInput{Notes: "again"},
		})
		require.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.JoinWaitlist(ctx, &waitlistssvc.JoinWaitlistRequest{})
		require.Error(t, err)
	})

	T.Run("nonexistent waitlist ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.JoinWaitlist(ctx, &waitlistssvc.JoinWaitlistRequest{
			WaitlistId: nonexistentID,
			Input:      &waitlistssvc.WaitlistSignupCreationRequestInput{Notes: t.Name()},
		})
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
		createdSignup := createWaitlistSignupForTest(t, testClient, waitlist.ID)

		retrieved, err := testClient.GetWaitlistSignup(ctx, &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: createdSignup.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)

		retrieved, err := testClient.GetWaitlistSignup(ctx, &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetWaitlistSignup(ctx, &waitlistssvc.GetWaitlistSignupRequest{})
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
		createdSignups := []*waitlists.Signup{createWaitlistSignupForTest(t, testClient, waitlist.ID)}
		for range exampleQuantity - 1 {
			_, joiner := createUserAndClientForTest(t)
			createdSignups = append(createdSignups, createWaitlistSignupForTest(t, joiner, waitlist.ID))
		}

		// the waitlist-wide signup listing is reserved for service admins: it hands
		// back every signatory's address.
		results, err := adminClient.GetWaitlistSignupsForWaitlist(ctx, &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: waitlist.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), len(createdSignups))
	})

	T.Run("denied for regular users", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		createWaitlistSignupForTest(t, testClient, waitlist.ID)

		_, err := testClient.GetWaitlistSignupsForWaitlist(ctx, &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: waitlist.ID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetWaitlistSignupsForWaitlist(ctx, &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{})
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
		createdSignup := createWaitlistSignupForTest(t, testClient, waitlist.ID)

		newNotes := "Updated notes"

		_, err := testClient.UpdateWaitlistSignup(ctx, &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: createdSignup.ID,
			Input: &waitlistssvc.WaitlistSignupUpdateRequestInput{
				Notes: &newNotes,
			},
		})
		require.NoError(t, err)

		retrieved, err := testClient.GetWaitlistSignup(ctx, &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: createdSignup.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, newNotes, retrieved.GetResult().GetNotes())

		// A note moves nobody: the signup is still waiting and its lifecycle stamp
		// is still unset.
		assert.Equal(t, waitlists.StatusWaiting.String(), retrieved.GetResult().GetStatus())
		assert.Nil(t, retrieved.GetResult().GetStatusChangedAt())
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)

		newNotes := "Updated notes"
		_, err := adminClient.UpdateWaitlistSignup(ctx, &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: nonexistentID,
			Input: &waitlistssvc.WaitlistSignupUpdateRequestInput{
				Notes: &newNotes,
			},
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateWaitlistSignup(ctx, &waitlistssvc.UpdateWaitlistSignupRequest{})
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
		signup := createWaitlistSignupForTest(t, testClient, waitlist.ID)

		invited, err := adminClient.InviteWaitlistSignup(ctx, &waitlistssvc.InviteWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: signup.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, waitlists.StatusInvited.String(), invited.GetUpdated().GetStatus())
		assert.NotNil(t, invited.GetUpdated().GetStatusChangedAt())

		_, err = adminClient.InviteWaitlistSignup(ctx, &waitlistssvc.InviteWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: signup.ID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))

		converted, err := adminClient.ConvertWaitlistSignup(ctx, &waitlistssvc.ConvertWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: signup.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, waitlists.StatusConverted.String(), converted.GetUpdated().GetStatus())
	})

	T.Run("denied for regular users", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		signup := createWaitlistSignupForTest(t, testClient, waitlist.ID)

		// Being on a list does not entitle somebody to invite themselves off it.
		_, err := testClient.InviteWaitlistSignup(ctx, &waitlistssvc.InviteWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: signup.ID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		_, err = testClient.ConvertWaitlistSignup(ctx, &waitlistssvc.ConvertWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: signup.ID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.InviteWaitlistSignup(ctx, &waitlistssvc.InviteWaitlistSignupRequest{})
		require.Error(t, err)

		_, err = c.ConvertWaitlistSignup(ctx, &waitlistssvc.ConvertWaitlistSignupRequest{})
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
		signup := createWaitlistSignupForTest(t, testClient, waitlist.ID)

		withdrawn, err := testClient.WithdrawFromWaitlist(ctx, &waitlistssvc.WithdrawFromWaitlistRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: signup.ID,
		})
		require.NoError(t, err)

		// The row is still live and no longer says who it was about, which is what
		// an unsubscribe page renders.
		assert.Equal(t, waitlists.StatusWithdrawn.String(), withdrawn.GetUpdated().GetStatus())
		assert.Empty(t, withdrawn.GetUpdated().GetContact())
		assert.Empty(t, withdrawn.GetUpdated().GetNotes())
		assert.Empty(t, withdrawn.GetUpdated().GetSubjectId())

		// Filling the form in again does not put them back on the list. This is the
		// whole obligation: the local table it replaced had no way to express it.
		_, err = testClient.JoinWaitlist(ctx, &waitlistssvc.JoinWaitlistRequest{
			WaitlistId: waitlist.ID,
			Input:      &waitlistssvc.WaitlistSignupCreationRequestInput{Notes: "changed my mind"},
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// A second withdrawal is refused, and the code is worth pinning because it is
		// the anonymization rather than the lifecycle guard that refuses it: the row
		// no longer names anybody, so the service can no longer tell that this caller
		// is the person it used to be about. The store's own answer to a replayed
		// withdrawal is ErrAlreadyWithdrawn (see the repository suite); nothing gets
		// that far from here, and that is the design working rather than around it.
		_, err = testClient.WithdrawFromWaitlist(ctx, &waitlistssvc.WithdrawFromWaitlistRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: signup.ID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("denied for somebody else's signup", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		_, otherClient := createUserAndClientForTest(t)

		waitlist := createWaitlistForTest(t, testClient)
		signup := createWaitlistSignupForTest(t, testClient, waitlist.ID)

		_, err := otherClient.WithdrawFromWaitlist(ctx, &waitlistssvc.WithdrawFromWaitlistRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: signup.ID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WithdrawFromWaitlist(ctx, &waitlistssvc.WithdrawFromWaitlistRequest{})
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
		createdSignup := createWaitlistSignupForTest(t, testClient, waitlist.ID)

		_, err := testClient.ArchiveWaitlistSignup(ctx, &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: createdSignup.ID,
		})
		require.NoError(t, err)

		// Archiving is not withdrawing: the row is hidden and the address is still
		// stored, so a second attempt is a duplicate rather than an honored opt-out.
		_, err = testClient.JoinWaitlist(ctx, &waitlistssvc.JoinWaitlistRequest{
			WaitlistId: waitlist.ID,
			Input:      &waitlistssvc.WaitlistSignupCreationRequestInput{Notes: "again"},
		})
		require.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)

		_, err := adminClient.ArchiveWaitlistSignup(ctx, &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistId:       waitlist.ID,
			WaitlistSignupId: nonexistentID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveWaitlistSignup(ctx, &waitlistssvc.ArchiveWaitlistSignupRequest{})
		assert.Error(t, err)
	})
}
