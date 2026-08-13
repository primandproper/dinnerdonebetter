package fakes

import (
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/converters"

	"github.com/primandproper/platform-go/v10/filtering"
)

// BuildFakeWaitlist builds a fake waitlist.
func BuildFakeWaitlist() *types.Waitlist {
	return &types.Waitlist{
		CreatedAt:   BuildFakeTime(),
		ID:          BuildFakeID(),
		Name:        buildUniqueString(),
		Description: buildUniqueString(),
		ValidUntil:  time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second),
	}
}

// BuildFakeWaitlistsList builds a fake list of waitlists.
func BuildFakeWaitlistsList() *filtering.QueryFilteredResult[types.Waitlist] {
	var waitlists []*types.Waitlist
	for range exampleQuantity {
		waitlists = append(waitlists, BuildFakeWaitlist())
	}

	return &filtering.QueryFilteredResult[types.Waitlist]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: waitlists,
	}
}

// BuildFakeWaitlistCreationRequestInput builds a fake WaitlistCreationRequestInput.
func BuildFakeWaitlistCreationRequestInput() *types.WaitlistCreationRequestInput {
	waitlist := BuildFakeWaitlist()
	return converters.ConvertWaitlistToWaitlistCreationRequestInput(waitlist)
}

// BuildFakeWaitlistUpdateRequestInput builds a fake WaitlistUpdateRequestInput.
func BuildFakeWaitlistUpdateRequestInput() *types.WaitlistUpdateRequestInput {
	waitlist := BuildFakeWaitlist()
	return converters.ConvertWaitlistToWaitlistUpdateRequestInput(waitlist)
}

// BuildFakeWaitlistSignup builds a fake waitlist signup.
func BuildFakeWaitlistSignup() *types.WaitlistSignup {
	return &types.WaitlistSignup{
		CreatedAt:         BuildFakeTime(),
		ID:                BuildFakeID(),
		Notes:             buildUniqueString(),
		BelongsToWaitlist: BuildFakeID(),
		BelongsToUser:     BuildFakeID(),
		BelongsToAccount:  BuildFakeID(),
	}
}

// BuildFakeWaitlistSignupsList builds a fake list of waitlist signups.
func BuildFakeWaitlistSignupsList() *filtering.QueryFilteredResult[types.WaitlistSignup] {
	var waitlistSignups []*types.WaitlistSignup
	for range exampleQuantity {
		waitlistSignups = append(waitlistSignups, BuildFakeWaitlistSignup())
	}

	return &filtering.QueryFilteredResult[types.WaitlistSignup]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: waitlistSignups,
	}
}

// BuildFakeWaitlistSignupCreationRequestInput builds a fake WaitlistSignupCreationRequestInput.
func BuildFakeWaitlistSignupCreationRequestInput() *types.WaitlistSignupCreationRequestInput {
	waitlistSignup := BuildFakeWaitlistSignup()
	return converters.ConvertWaitlistSignupToWaitlistSignupCreationRequestInput(waitlistSignup)
}

// BuildFakeWaitlistSignupUpdateRequestInput builds a fake WaitlistSignupUpdateRequestInput.
func BuildFakeWaitlistSignupUpdateRequestInput() *types.WaitlistSignupUpdateRequestInput {
	waitlistSignup := BuildFakeWaitlistSignup()
	return converters.ConvertWaitlistSignupToWaitlistSignupUpdateRequestInput(waitlistSignup)
}
