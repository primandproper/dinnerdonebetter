package identity

type (
	// UserDataCollection contains the identity data disclosed to a user for GDPR/CCPA purposes.
	UserDataCollection struct {
		User                   User                    `json:"user"`
		Accounts               []Account               `json:"accounts"`
		AccountInvitations     []AccountInvitation     `json:"account_invitations"`
		AccountUserMemberships []AccountUserMembership `json:"account_user_memberships"`
	}
)
