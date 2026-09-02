// Package fakes builds the randomized waitlists and signups this application's
// tests write.
package fakes

import (
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	platformwaitlists "github.com/primandproper/platform-go/v13/waitlists"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeWaitlist builds a faked List that is still taking signups.
//
// The closing time is forced into the future rather than randomized. An
// arbitrary timestamp is as likely to be in the past as ahead of it, and a
// closed list takes no signups — so half of the tests that opened one would be
// testing the closed path without saying so.
func BuildFakeWaitlist() *platformwaitlists.List {
	list := fake.BuildFakeRecord[platformwaitlists.List]()
	list.ClosesAt = time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	list.Scope = waitlists.Scope()

	return list
}

// BuildFakeWaitlistList builds a faked page of Lists.
func BuildFakeWaitlistList() *filtering.QueryFilteredResult[platformwaitlists.List] {
	return fake.BuildFakePage(BuildFakeWaitlist)
}

// BuildFakeWaitlistSignup builds a faked Signup: waiting, anonymous, and with a
// contact that looks like the address a list would write to.
//
// The status is fixed rather than randomized because a signup is born waiting
// and the store refuses one that arrives in any other status — a randomized
// status would build a signup that could never be written. The subject is left
// empty for the same reason the contact is not: a caller that wants one names
// the user it belongs to, and a random one would name a user that does not
// exist.
func BuildFakeWaitlistSignup() *platformwaitlists.Signup {
	signup := fake.BuildFakeRecord[platformwaitlists.Signup]()
	signup.Contact = gofakeit.Email()
	signup.ContactDigest = ""
	signup.Status = platformwaitlists.StatusWaiting
	signup.StatusChangedAt = nil
	signup.Subject = platformwaitlists.Subject{}
	signup.Scope = waitlists.Scope()

	return signup
}

// BuildFakeWaitlistSignupForUser builds a faked Signup belonging to one user,
// which is the only shape this application writes.
func BuildFakeWaitlistSignupForUser(userID string) *platformwaitlists.Signup {
	signup := BuildFakeWaitlistSignup()
	signup.Subject = waitlists.SubjectFor(userID)

	return signup
}

// BuildFakeWaitlistSignupList builds a faked page of Signups.
func BuildFakeWaitlistSignupList() *filtering.QueryFilteredResult[platformwaitlists.Signup] {
	return fake.BuildFakePage(BuildFakeWaitlistSignup)
}
