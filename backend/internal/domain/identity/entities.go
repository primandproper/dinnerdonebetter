package identity

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: Account{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `accountID := BuildFakeID()`},
					{Code: `var memberships []*types.AccountUserMembershipWithUser`},
					{Code: `for range exampleQuantity {
	membership := BuildFakeAccountUserMembershipWithUser()
	membership.BelongsToAccount = accountID
	memberships = append(memberships, membership)
}`},
					{Code: `fakeAddress := fake.Address()`},
					{Code: `key := fake.BitcoinPrivateKey()`},
				},
				Fields: []entitydecl.Field{
					{Name: "ID", Expr: `accountID`},
					{Name: "Name", Expr: `fake.UUID()`},
					{Name: "BillingStatus", Expr: `types.UnpaidAccountBillingStatus`},
					{Name: "ContactPhone", Expr: `fake.PhoneFormatted()`},
					{Name: "PaymentProcessorCustomerID", Expr: `fake.UUID()`},
					{Name: "BelongsToUser", Expr: `fake.UUID()`},
					{Name: "Members", Expr: `memberships`},
					{Name: "AddressLine1", Expr: `fakeAddress.Address`},
					{Name: "AddressLine2", Expr: `""`},
					{Name: "City", Expr: `fakeAddress.City`},
					{Name: "State", Expr: `fakeAddress.State`},
					{Name: "ZipCode", Expr: `fakeAddress.Zip`},
					{Name: "Country", Expr: `fakeAddress.Country`},
					{Name: "Latitude", Expr: `new(buildFakeNumber())`},
					{Name: "Longitude", Expr: `new(buildFakeNumber())`},
					{Name: "WebhookEncryptionKey", Expr: `key`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: AccountUpdateRequestInput{}, Converter: "ConvertAccountToAccountUpdateRequestInput"},
					{Type: AccountCreationRequestInput{}, Converter: "ConvertAccountToAccountCreationRequestInput"},
				},
			},
		},
		{
			Type: AccountOwnershipTransferInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Reason", Expr: `fake.Sentence(5)`},
					{Name: "CurrentOwner", Expr: `BuildFakeID()`},
					{Name: "NewOwner", Expr: `BuildFakeID()`},
				},
			},
		},
		{
			Type: AccountInvitation{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "ToEmail", Expr: `fake.Email()`},
					{Name: "ToUser", Expr: `func(s string) *string { return &s }(buildUniqueString())`},
					{Name: "Token", Expr: `fake.UUID()`},
					{Name: "Status", Expr: `string(types.PendingAccountInvitationStatus)`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: AccountInvitationCreationRequestInput{}, Converter: "ConvertAccountInvitationToAccountInvitationCreationInput"},
				},
			},
		},
		{
			Type: AccountInvitationUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Token", Expr: `BuildFakeID()`},
					{Name: "Note", Expr: `fake.Sentence(3)`},
				},
			},
		},
		{
			Type: AccountUserMembership{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "BelongsToAccount", Expr: `fake.UUID()`},
					{Name: "DefaultAccount", Expr: `false`},
				},
			},
		},
		{
			Type: AccountUserMembershipWithUser{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `u := BuildFakeUser()`},
					{Code: `u.TwoFactorSecret = ""`},
				},
				Fields: []entitydecl.Field{
					{Name: "BelongsToUser", Expr: `u`},
					{Name: "BelongsToAccount", Expr: `fake.UUID()`},
					{Name: "DefaultAccount", Expr: `false`},
				},
			},
		},
		{
			Type: User{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `fakeDate := BuildFakeTime()`},
				},
				Fields: []entitydecl.Field{
					{Name: "FirstName", Expr: `fake.FirstName()`},
					{Name: "LastName", Expr: `fake.LastName()`},
					{Name: "EmailAddress", Expr: `fake.Email()`},
					{Name: "Username", Expr: `fmt.Sprintf("%s_%d_%s", fake.Username(), fake.Uint8(), fake.Username())`},
					{Name: "Birthday", Expr: `new(BuildFakeTime())`},
					{Name: "AccountStatus", Expr: `string(types.UnverifiedAccountStatus)`},
					{Name: "TwoFactorSecret", Expr: `base32.StdEncoding.EncodeToString([]byte(fake.Password(false, true, true, false, false, 32)))`},
					{Name: "TwoFactorSecretVerifiedAt", Expr: `&fakeDate`},
					{Name: "AccountStatusExplanation", Expr: `""`},
					{Name: "HashedPassword", Expr: `""`},
					{Name: "RequiresPasswordChange", Expr: `false`},
				},
				List: &entitydecl.List{},
			},
		},
		{
			Type: UserRegistrationInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `user := BuildFakeUser()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Username", Expr: `user.Username`},
					{Name: "FirstName", Expr: `user.FirstName`},
					{Name: "LastName", Expr: `user.LastName`},
					{Name: "EmailAddress", Expr: `user.EmailAddress`},
					{Name: "Password", Expr: `buildFakePassword()`},
					{Name: "Birthday", Expr: `user.Birthday`},
					{Name: "InvitationToken", Expr: `""`},
					{Name: "InvitationID", Expr: `""`},
					{Name: "AccountName", Expr: `""`},
					{Name: "AcceptedTOS", Expr: `false`},
					{Name: "AcceptedPrivacyPolicy", Expr: `false`},
				},
			},
		},
		{
			Type: UserCreationResponse{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `user := BuildFakeUser()`},
				},
				Fields: []entitydecl.Field{
					{Name: "CreatedAt", Expr: `user.CreatedAt`},
					{Name: "Birthday", Expr: `user.Birthday`},
					{Name: "Username", Expr: `user.Username`},
					{Name: "EmailAddress", Expr: `user.EmailAddress`},
					{Name: "TwoFactorQRCode", Expr: `fake.UUID()`},
					{Name: "CreatedUserID", Expr: `user.ID`},
					{Name: "AccountStatus", Expr: `user.AccountStatus`},
					{Name: "TwoFactorSecret", Expr: `user.TwoFactorSecret`},
					{Name: "FirstName", Expr: `user.FirstName`},
					{Name: "LastName", Expr: `user.LastName`},
				},
			},
		},
		{
			Type: AvatarUpdateInput{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: UserDetailsUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "CurrentPassword", Expr: `fake.Password(true, true, true, false, false, 32)`},
					{Name: "TOTPToken", Expr: `"123456"`},
				},
			},
		},
		{
			Type: UserDetailsDatabaseUpdateInput{},
			Fake: entitydecl.Fake{},
		},
	},
}
