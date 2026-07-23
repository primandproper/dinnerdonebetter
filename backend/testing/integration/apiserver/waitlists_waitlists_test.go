package integration

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/converters"
	waitlistfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/waitlists/grpc/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func checkWaitlistEquality(t *testing.T, expected, actual *waitlists.Waitlist) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected Waitlist to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected Waitlist to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected Waitlist Name")
	assert.Equal(t, expected.Description, actual.Description, "expected Waitlist Description")
	assert.WithinDuration(t, expected.ValidUntil, actual.ValidUntil, time.Second, "expected Waitlist ValidUntil")
}

func checkWaitlistSignupEquality(t *testing.T, expected, actual *waitlists.WaitlistSignup) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected WaitlistSignup to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected WaitlistSignup to have CreatedAt")

	assert.Equal(t, expected.Notes, actual.Notes, "expected WaitlistSignup Notes")
	assert.Equal(t, expected.BelongsToWaitlist, actual.BelongsToWaitlist, "expected WaitlistSignup BelongsToWaitlist")
	assert.NotEmpty(t, actual.BelongsToUser, "expected WaitlistSignup to have BelongsToUser")
	assert.NotEmpty(t, actual.BelongsToAccount, "expected WaitlistSignup to have BelongsToAccount")
}

func createWaitlistForTest(t *testing.T, testClient client.Client) *waitlists.Waitlist {
	t.Helper()
	ctx := t.Context()

	exampleWaitlist := waitlistfakes.BuildFakeWaitlist()
	exampleWaitlistInput := converters.ConvertWaitlistToWaitlistCreationRequestInput(exampleWaitlist)

	input := &waitlistssvc.WaitlistCreationRequestInput{
		Name:        exampleWaitlistInput.Name,
		Description: exampleWaitlistInput.Description,
		ValidUntil:  timestamppb.New(exampleWaitlistInput.ValidUntil),
	}

	createdWaitlist, err := adminClient.CreateWaitlist(ctx, &waitlistssvc.CreateWaitlistRequest{Input: input})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCWaitlistToWaitlist(createdWaitlist.Created)
	checkWaitlistEquality(t, exampleWaitlist, converted)

	retrievedWaitlist, err := testClient.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: createdWaitlist.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, retrievedWaitlist)

	waitlist := grpcconverters.ConvertGRPCWaitlistToWaitlist(retrievedWaitlist.Result)
	checkWaitlistEquality(t, converted, waitlist)

	return waitlist
}

// createWaitlistSignupForTest creates a signup as testClient's user; signups are user-owned, so the
// creating client is the owner for subsequent reads/updates/archives.
func createWaitlistSignupForTest(t *testing.T, testClient client.Client, waitlistID string) *waitlists.WaitlistSignup {
	t.Helper()
	ctx := t.Context()

	exampleSignup := waitlistfakes.BuildFakeWaitlistSignup()
	exampleSignup.BelongsToWaitlist = waitlistID
	exampleSignupInput := converters.ConvertWaitlistSignupToWaitlistSignupCreationRequestInput(exampleSignup)

	input := &waitlistssvc.WaitlistSignupCreationRequestInput{
		Notes:             exampleSignupInput.Notes,
		BelongsToWaitlist: exampleSignupInput.BelongsToWaitlist,
		BelongsToUser:     exampleSignupInput.BelongsToUser,
		BelongsToAccount:  exampleSignupInput.BelongsToAccount,
	}

	createdSignup, err := testClient.CreateWaitlistSignup(ctx, &waitlistssvc.CreateWaitlistSignupRequest{
		Input: input,
	})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCWaitlistSignupToWaitlistSignup(createdSignup.Created)
	// Note: BelongsToUser and BelongsToAccount are set by the service from session context,
	// so we only check that they're not empty rather than matching the fake data
	assert.Equal(t, exampleSignup.Notes, converted.Notes, "expected WaitlistSignup Notes")
	assert.Equal(t, exampleSignup.BelongsToWaitlist, converted.BelongsToWaitlist, "expected WaitlistSignup BelongsToWaitlist")
	assert.NotEmpty(t, converted.BelongsToUser, "expected WaitlistSignup to have BelongsToUser")
	assert.NotEmpty(t, converted.BelongsToAccount, "expected WaitlistSignup to have BelongsToAccount")

	retrievedSignup, err := testClient.GetWaitlistSignup(ctx, &waitlistssvc.GetWaitlistSignupRequest{
		WaitlistSignupId: createdSignup.Created.Id,
		WaitlistId:       waitlistID,
	})
	require.NoError(t, err)
	require.NotNil(t, retrievedSignup)

	signup := grpcconverters.ConvertGRPCWaitlistSignupToWaitlistSignup(retrievedSignup.Result)
	// Verify retrieved signup matches created signup
	assert.Equal(t, converted.ID, signup.ID, "expected WaitlistSignup ID to match")
	assert.Equal(t, converted.Notes, signup.Notes, "expected WaitlistSignup Notes to match")
	assert.Equal(t, converted.BelongsToWaitlist, signup.BelongsToWaitlist, "expected WaitlistSignup BelongsToWaitlist to match")
	assert.Equal(t, converted.BelongsToUser, signup.BelongsToUser, "expected WaitlistSignup BelongsToUser to match")
	assert.Equal(t, converted.BelongsToAccount, signup.BelongsToAccount, "expected WaitlistSignup BelongsToAccount to match")

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

		input := &waitlistssvc.WaitlistCreationRequestInput{
			Name:        "", // Invalid: empty name
			Description: t.Name(),
			ValidUntil:  timestamppb.New(time.Now().Add(-24 * time.Hour)), // Invalid: in the past
		}

		_, err := adminClient.CreateWaitlist(ctx, &waitlistssvc.CreateWaitlistRequest{Input: input})
		assert.Error(t, err)
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
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
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

		createdWaitlists := []*waitlists.Waitlist{}
		for range exampleQuantity {
			createdWaitlists = append(createdWaitlists, createWaitlistForTest(t, testClient))
		}

		results, err := testClient.GetWaitlists(ctx, &waitlistssvc.GetWaitlistsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.True(t, len(results.Results) >= len(createdWaitlists))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetWaitlists(ctx, &waitlistssvc.GetWaitlistsRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_ListingActive(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdWaitlists := []*waitlists.Waitlist{}
		for range exampleQuantity {
			createdWaitlists = append(createdWaitlists, createWaitlistForTest(t, testClient))
		}

		results, err := testClient.GetActiveWaitlists(ctx, &waitlistssvc.GetActiveWaitlistsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		// All created waitlists should be active (not expired)
		assert.True(t, len(results.Results) >= len(createdWaitlists))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetActiveWaitlists(ctx, &waitlistssvc.GetActiveWaitlistsRequest{})
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
		newValidUntil := time.Now().Add(48 * time.Hour)

		_, err := adminClient.UpdateWaitlist(ctx, &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: createdWaitlist.ID,
			Input: &waitlistssvc.WaitlistUpdateRequestInput{
				Name:        &newName,
				Description: &newDescription,
				ValidUntil:  timestamppb.New(newValidUntil),
			},
		})
		assert.NoError(t, err)

		retrieved, err := testClient.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: createdWaitlist.ID})
		require.NoError(t, err)
		assert.Equal(t, newName, retrieved.Result.Name)
		assert.Equal(t, newDescription, retrieved.Result.Description)
		assert.WithinDuration(t, newValidUntil, grpcconverters.ConvertGRPCWaitlistToWaitlist(retrieved.Result).ValidUntil, time.Second)
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
		assert.Error(t, err)
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
		assert.NoError(t, err)
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWaitlistForTest(t, testClient)

		_, err := adminClient.ArchiveWaitlist(ctx, &waitlistssvc.ArchiveWaitlistRequest{WaitlistId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveWaitlist(ctx, &waitlistssvc.ArchiveWaitlistRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlists_IsNotExpired(T *testing.T) {
	T.Parallel()

	T.Run("happy path - not expired", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWaitlist := createWaitlistForTest(t, testClient)

		result, err := testClient.WaitlistIsNotExpired(ctx, &waitlistssvc.WaitlistIsNotExpiredRequest{WaitlistId: createdWaitlist.ID})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsNotExpired)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.WaitlistIsNotExpired(ctx, &waitlistssvc.WaitlistIsNotExpiredRequest{WaitlistId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WaitlistIsNotExpired(ctx, &waitlistssvc.WaitlistIsNotExpiredRequest{})
		assert.Error(t, err)
	})
}

func TestWaitlistSignups_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		createWaitlistSignupForTest(t, testClient, waitlist.ID)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateWaitlistSignup(ctx, &waitlistssvc.CreateWaitlistSignupRequest{})
		require.Error(t, err)
	})

	T.Run("nonexistent waitlist ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleSignup := waitlistfakes.BuildFakeWaitlistSignup()
		exampleSignupInput := converters.ConvertWaitlistSignupToWaitlistSignupCreationRequestInput(exampleSignup)

		input := &waitlistssvc.WaitlistSignupCreationRequestInput{
			Notes:             exampleSignupInput.Notes,
			BelongsToWaitlist: nonexistentID,
			BelongsToUser:     exampleSignupInput.BelongsToUser,
			BelongsToAccount:  exampleSignupInput.BelongsToAccount,
		}

		_, err := adminClient.CreateWaitlistSignup(ctx, &waitlistssvc.CreateWaitlistSignupRequest{
			Input: input,
		})
		assert.Error(t, err)
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
			WaitlistSignupId: createdSignup.ID,
			WaitlistId:       waitlist.ID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)

		retrieved, err := testClient.GetWaitlistSignup(ctx, &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistSignupId: nonexistentID,
			WaitlistId:       waitlist.ID,
		})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
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

		createdSignups := []*waitlists.WaitlistSignup{}
		for range exampleQuantity {
			createdSignups = append(createdSignups, createWaitlistSignupForTest(t, testClient, waitlist.ID))
		}

		// the waitlist-wide signup listing is reserved for service admins.
		results, err := adminClient.GetWaitlistSignupsForWaitlist(ctx, &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: waitlist.ID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.True(t, len(results.Results) >= len(createdSignups))
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
			WaitlistSignupId: createdSignup.ID,
			WaitlistId:       waitlist.ID,
			Input: &waitlistssvc.WaitlistSignupUpdateRequestInput{
				Notes: &newNotes,
			},
		})
		assert.NoError(t, err)

		retrieved, err := testClient.GetWaitlistSignup(ctx, &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistSignupId: createdSignup.ID,
			WaitlistId:       waitlist.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, newNotes, retrieved.Result.Notes)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)

		newNotes := "Updated notes"
		_, err := adminClient.UpdateWaitlistSignup(ctx, &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistSignupId: nonexistentID,
			WaitlistId:       waitlist.ID,
			Input: &waitlistssvc.WaitlistSignupUpdateRequestInput{
				Notes: &newNotes,
			},
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateWaitlistSignup(ctx, &waitlistssvc.UpdateWaitlistSignupRequest{})
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
			WaitlistSignupId: createdSignup.ID,
		})
		assert.NoError(t, err)
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		waitlist := createWaitlistForTest(t, testClient)
		createWaitlistSignupForTest(t, testClient, waitlist.ID)

		_, err := adminClient.ArchiveWaitlistSignup(ctx, &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistSignupId: nonexistentID,
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveWaitlistSignup(ctx, &waitlistssvc.ArchiveWaitlistSignupRequest{})
		assert.Error(t, err)
	})
}
