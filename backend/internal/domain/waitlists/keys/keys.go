package keys

const (
	idSuffix = ".id"

	// WaitlistIDKey is the standard key for referring to a waitlist ID.
	WaitlistIDKey = "waitlist" + idSuffix
	// WaitlistSignupIDKey is the standard key for referring to a waitlist signup ID.
	WaitlistSignupIDKey = "waitlist_signup" + idSuffix
	// WaitlistSignupStatusKey is the standard key for where a signup stands:
	// waiting, invited, converted or withdrawn.
	WaitlistSignupStatusKey = "waitlist_signup.status"
)
