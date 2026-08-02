package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeAccountUserMembership builds a faked AccountUserMembership.
func BuildFakeAccountUserMembership() *types.AccountUserMembership {
	return &types.AccountUserMembership{
		ID:               BuildFakeID(),
		BelongsToUser:    BuildFakeID(),
		BelongsToAccount: fake.UUID(),
		CreatedAt:        BuildFakeTime(),
		ArchivedAt:       nil,
	}
}

// BuildFakeAccountUserMembershipWithUser builds a faked AccountUserMembershipWithUser.
func BuildFakeAccountUserMembershipWithUser() *types.AccountUserMembershipWithUser {
	u := BuildFakeUser()
	u.TwoFactorSecret = ""

	return &types.AccountUserMembershipWithUser{
		ID:               BuildFakeID(),
		BelongsToUser:    u,
		BelongsToAccount: fake.UUID(),
		CreatedAt:        BuildFakeTime(),
		ArchivedAt:       nil,
	}
}
