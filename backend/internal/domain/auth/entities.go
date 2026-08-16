package auth

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: ChangeActiveAccountInput{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: UserStatusResponse{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "AccountStatus", Expr: `identity.GoodStandingUserAccountStatus.String()`},
					{Name: "ActiveAccount", Expr: `BuildFakeID()`},
				},
			},
		},
		{
			Type: TokenResponse{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "AccessToken", Expr: `fake.UUID()`},
				},
			},
		},
		{
			Type: UserLoginInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Username", Expr: `BuildFakeID()`},
					{Name: "Password", Expr: `buildFakePassword()`},
					{Name: "TOTPToken", Expr: `buildFakeTOTPToken()`},
				},
			},
		},
		{
			Type: PasswordUpdateInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "NewPassword", Expr: `buildFakePassword()`},
					{Name: "CurrentPassword", Expr: `buildFakePassword()`},
					{Name: "TOTPToken", Expr: `fmt.Sprintf("0%s", fake.Zip())`},
				},
			},
		},
		{
			Type: TOTPSecretRefreshInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "CurrentPassword", Expr: `buildFakePassword()`},
					{Name: "TOTPToken", Expr: `fmt.Sprintf("0%s", fake.Zip())`},
				},
			},
		},
		{
			Type: TOTPSecretRefreshResponse{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "TwoFactorQRCode", Expr: `fake.UUID()`},
					{Name: "TwoFactorSecret", Expr: `fake.UUID()`},
				},
			},
		},
		{
			Type: UserPermissionsRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Permissions", Expr: `[]string{
	buildUniqueString(),
	buildUniqueString(),
	buildUniqueString(),
}`},
				},
			},
		},
		{
			Type: PasswordResetToken{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `fakeDate := BuildFakeTime()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Token", Expr: `fake.UUID()`},
					{Name: "ExpiresAt", Expr: `fakeDate.Add(30 * time.Minute)`},
					{Name: "CreatedAt", Expr: `fakeDate`},
				},
			},
		},
		{
			Type: UsernameReminderRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "EmailAddress", Expr: `fake.Email()`},
				},
			},
		},
		{
			Type: PasswordResetTokenCreationRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "EmailAddress", Expr: `fake.Email()`},
				},
			},
		},
		{
			Type: PasswordResetTokenRedemptionRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "NewPassword", Expr: `buildFakePassword()`},
				},
			},
		},
		{
			Type: EmailAddressVerificationRequestInput{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: UsernameUpdateInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "CurrentPassword", Expr: `fake.Password(true, true, true, false, false, 32)`},
					{Name: "TOTPToken", Expr: `"123456"`},
				},
			},
		},
		{
			Type: UserEmailAddressUpdateInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "NewEmailAddress", Expr: `fake.Email()`},
					{Name: "CurrentPassword", Expr: `fake.Password(true, true, true, false, false, 32)`},
					{Name: "TOTPToken", Expr: `"123456"`},
				},
			},
		},
		{
			Type: PasswordResetResponse{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: UserPermissionsResponse{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Permissions", Expr: `map[string]bool{
	string(authorization.CreateWebhooksPermission): true,
}`},
				},
			},
		},
	},
}
