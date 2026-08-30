package fakes

import (
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// BuildFakeWaitlist builds a fake waitlist.
func BuildFakeWaitlist() *types.Waitlist {
	waitlist := fake.BuildFakeRecord[types.Waitlist]()

	// A waitlist people can still sign up for. An arbitrary timestamp is as likely to
	// be in the past as the future, and a closed waitlist takes no signups — so half of
	// the tests that create one would be testing the closed path without saying so.
	waitlist.ValidUntil = time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)

	return waitlist
}

// BuildFakeWaitlistsList builds a fake list of waitlists.
func BuildFakeWaitlistsList() *filtering.QueryFilteredResult[types.Waitlist] {
	return fake.BuildFakePage(BuildFakeWaitlist)
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
	return fake.BuildFakeRecord[types.WaitlistSignup]()
}

// BuildFakeWaitlistSignupsList builds a fake list of waitlist signups.
func BuildFakeWaitlistSignupsList() *filtering.QueryFilteredResult[types.WaitlistSignup] {
	return fake.BuildFakePage(BuildFakeWaitlistSignup)
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
