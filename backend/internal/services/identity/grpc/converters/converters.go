package converters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	uploadedmediaconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc/converters"

	"github.com/primandproper/platform-go/v5/pointer"
)

func convertUserAvatarToGRPC(avatar *uploadedmedia.UploadedMedia) *uploadedmediasvc.UploadedMedia {
	if avatar == nil {
		return nil
	}
	return uploadedmediaconverters.ConvertUploadedMediaToGRPCUploadedMedia(avatar)
}

func convertGRPCAvatarToUser(avatar *uploadedmediasvc.UploadedMedia) *uploadedmedia.UploadedMedia {
	if avatar == nil {
		return nil
	}
	return uploadedmediaconverters.ConvertGRPCUploadedMediaToUploadedMedia(avatar)
}

func ConvertGRPCUserDetailsUpdateRequestInputToUserDetailsUpdateRequestInput(input *identitysvc.UserDetailsUpdateRequestInput) *identity.UserDetailsUpdateRequestInput {
	return &identity.UserDetailsUpdateRequestInput{
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Birthday:        grpcconverters.ConvertPBTimestampToTime(input.Birthday),
		CurrentPassword: input.CurrentPassword,
		TOTPToken:       input.TotpToken,
	}
}

func ConvertUserToGRPCUser(input *identity.User) *identitysvc.User {
	return &identitysvc.User{
		CreatedAt:                  grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		PasswordLastChangedAt:      grpcconverters.ConvertTimePointerToPBTimestamp(input.PasswordLastChangedAt),
		LastUpdatedAt:              grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		LastAcceptedTermsOfService: grpcconverters.ConvertTimePointerToPBTimestamp(input.LastAcceptedTermsOfService),
		LastAcceptedPrivacyPolicy:  grpcconverters.ConvertTimePointerToPBTimestamp(input.LastAcceptedPrivacyPolicy),
		TwoFactorSecretVerifiedAt:  grpcconverters.ConvertTimePointerToPBTimestamp(input.TwoFactorSecretVerifiedAt),
		EmailAddressVerifiedAt:     grpcconverters.ConvertTimePointerToPBTimestamp(input.EmailAddressVerifiedAt),
		Birthday:                   grpcconverters.ConvertTimePointerToPBTimestamp(input.Birthday),
		ArchivedAt:                 grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		HashedPassword:             input.HashedPassword,
		LastName:                   input.LastName,
		AccountStatusExplanation:   input.AccountStatusExplanation,
		Id:                         input.ID,
		AccountStatus:              input.AccountStatus,
		Username:                   input.Username,
		FirstName:                  input.FirstName,
		TwoFactorSecret:            input.TwoFactorSecret,
		EmailAddress:               input.EmailAddress,
		Avatar:                     convertUserAvatarToGRPC(input.Avatar),
		RequiresPasswordChange:     input.RequiresPasswordChange,
	}
}

func ConvertGRPCUserToUser(input *identitysvc.User) *identity.User {
	return &identity.User{
		CreatedAt:                  grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		PasswordLastChangedAt:      grpcconverters.ConvertPBTimestampToTimePointer(input.PasswordLastChangedAt),
		LastUpdatedAt:              grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		LastAcceptedTermsOfService: grpcconverters.ConvertPBTimestampToTimePointer(input.LastAcceptedTermsOfService),
		LastAcceptedPrivacyPolicy:  grpcconverters.ConvertPBTimestampToTimePointer(input.LastAcceptedPrivacyPolicy),
		TwoFactorSecretVerifiedAt:  grpcconverters.ConvertPBTimestampToTimePointer(input.TwoFactorSecretVerifiedAt),
		EmailAddressVerifiedAt:     grpcconverters.ConvertPBTimestampToTimePointer(input.EmailAddressVerifiedAt),
		Birthday:                   grpcconverters.ConvertPBTimestampToTimePointer(input.Birthday),
		ArchivedAt:                 grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		HashedPassword:             input.HashedPassword,
		LastName:                   input.LastName,
		AccountStatusExplanation:   input.AccountStatusExplanation,
		ID:                         input.Id,
		AccountStatus:              input.AccountStatus,
		Username:                   input.Username,
		FirstName:                  input.FirstName,
		TwoFactorSecret:            input.TwoFactorSecret,
		EmailAddress:               input.EmailAddress,
		Avatar:                     convertGRPCAvatarToUser(input.Avatar),
		RequiresPasswordChange:     input.RequiresPasswordChange,
	}
}

// ConvertGRPCAccountInvitationUpdateRequestInputToAccountInvitationUpdateRequestInput creates a AccountInvitationDatabaseCreationInput from a AccountInvitationCreationRequestInput.
func ConvertGRPCAccountInvitationUpdateRequestInputToAccountInvitationUpdateRequestInput(input *identitysvc.AccountInvitationUpdateRequestInput) *identity.AccountInvitationUpdateRequestInput {
	x := &identity.AccountInvitationUpdateRequestInput{
		Token: input.Token,
		Note:  input.Note,
	}

	return x
}

func ConvertGRPCAccountCreationRequestInputToAccountCreationRequestInput(input *identitysvc.AccountCreationRequestInput, belongsToUser string) *identity.AccountCreationRequestInput {
	return &identity.AccountCreationRequestInput{
		BelongsToUser: belongsToUser,
		Latitude:      new(float64(pointer.Dereference(input.Latitude))),
		Longitude:     new(float64(pointer.Dereference(input.Longitude))),
		Name:          input.Name,
		ContactPhone:  input.ContactPhone,
		AddressLine1:  input.AddressLine1,
		AddressLine2:  input.AddressLine2,
		City:          input.City,
		State:         input.State,
		ZipCode:       input.ZipCode,
		Country:       input.Country,
	}
}

func ConvertAccountCreationRequestInputToGRPCAccountCreationRequestInput(input *identity.AccountCreationRequestInput) *identitysvc.AccountCreationRequestInput {
	return &identitysvc.AccountCreationRequestInput{
		Latitude:      new(float32(pointer.Dereference(input.Latitude))),
		Longitude:     new(float32(pointer.Dereference(input.Longitude))),
		Name:          input.Name,
		ContactPhone:  input.ContactPhone,
		AddressLine1:  input.AddressLine1,
		AddressLine2:  input.AddressLine2,
		City:          input.City,
		State:         input.State,
		ZipCode:       input.ZipCode,
		Country:       input.Country,
		BelongsToUser: input.BelongsToUser,
	}
}

func ConvertAccountToGRPCAccount(input *identity.Account) *identitysvc.Account {
	var members []*identitysvc.AccountUserMembershipWithUser
	for _, member := range input.Members {
		members = append(members, ConvertAccountUserMembershipWithUserToGRPCAccountUserMembershipWithUser(member))
	}

	return &identitysvc.Account{
		CreatedAt:                  grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt:              grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:                 grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		Longitude:                  new(float32(pointer.Dereference(input.Longitude))),
		Latitude:                   new(float32(pointer.Dereference(input.Latitude))),
		SubscriptionPlanId:         input.SubscriptionPlanID,
		State:                      input.State,
		ContactPhone:               input.ContactPhone,
		City:                       input.City,
		AddressLine1:               input.AddressLine1,
		ZipCode:                    input.ZipCode,
		Country:                    input.Country,
		BillingStatus:              input.BillingStatus,
		AddressLine2:               input.AddressLine2,
		PaymentProcessorCustomerId: input.PaymentProcessorCustomerID,
		BelongsToUser:              input.BelongsToUser,
		Id:                         input.ID,
		Name:                       input.Name,
		WebhookEncryptionKey:       input.WebhookEncryptionKey,
		Members:                    members,
	}
}

func ConvertAccountUserMembershipWithUserToGRPCAccountUserMembershipWithUser(input *identity.AccountUserMembershipWithUser) *identitysvc.AccountUserMembershipWithUser {
	return &identitysvc.AccountUserMembershipWithUser{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		BelongsToUser:    ConvertUserToGRPCUser(input.BelongsToUser),
		Id:               input.ID,
		BelongsToAccount: input.BelongsToAccount,
		DefaultAccount:   input.DefaultAccount,
	}
}

func ConvertGRPCAccountToAccount(input *identitysvc.Account) *identity.Account {
	var members []*identity.AccountUserMembershipWithUser
	for _, member := range input.Members {
		members = append(members, ConvertGRPCAccountUserMembershipWithUserToAccountUserMembershipWithUser(member))
	}

	return &identity.Account{
		CreatedAt:                  grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		LastUpdatedAt:              grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ArchivedAt:                 grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		Longitude:                  new(float64(pointer.Dereference(input.Longitude))),
		Latitude:                   new(float64(pointer.Dereference(input.Latitude))),
		SubscriptionPlanID:         input.SubscriptionPlanId,
		State:                      input.State,
		ContactPhone:               input.ContactPhone,
		City:                       input.City,
		AddressLine1:               input.AddressLine1,
		ZipCode:                    input.ZipCode,
		Country:                    input.Country,
		BillingStatus:              input.BillingStatus,
		AddressLine2:               input.AddressLine2,
		PaymentProcessorCustomerID: input.PaymentProcessorCustomerId,
		BelongsToUser:              input.BelongsToUser,
		ID:                         input.Id,
		Name:                       input.Name,
		WebhookEncryptionKey:       input.WebhookEncryptionKey,
		Members:                    members,
	}
}

func ConvertGRPCAccountUserMembershipWithUserToAccountUserMembershipWithUser(input *identitysvc.AccountUserMembershipWithUser) *identity.AccountUserMembershipWithUser {
	return &identity.AccountUserMembershipWithUser{
		CreatedAt:        grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		LastUpdatedAt:    grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ArchivedAt:       grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		BelongsToUser:    ConvertGRPCUserToUser(input.BelongsToUser),
		ID:               input.Id,
		BelongsToAccount: input.BelongsToAccount,
		DefaultAccount:   input.DefaultAccount,
	}
}

func ConvertGRPCAccountInvitationCreationRequestInputToAccountInvitationCreationRequestInput(input *identitysvc.AccountInvitationCreationRequestInput) *identity.AccountInvitationCreationRequestInput {
	return &identity.AccountInvitationCreationRequestInput{
		ExpiresAt: grpcconverters.ConvertPBTimestampToTimePointer(input.ExpiresAt),
		Note:      input.Note,
		ToEmail:   input.ToEmail,
		ToName:    input.ToName,
	}
}

func ConvertAccountInvitationToGRPCAccountInvitation(input *identity.AccountInvitation) *identitysvc.AccountInvitation {
	return &identitysvc.AccountInvitation{
		CreatedAt:          grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt:      grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:         grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		ToUser:             input.ToUser,
		Status:             input.Status,
		ToEmail:            input.ToEmail,
		StatusNote:         input.StatusNote,
		Token:              input.Token,
		Id:                 input.ID,
		Note:               input.Note,
		ToName:             input.ToName,
		ExpiresAt:          grpcconverters.ConvertTimeToPBTimestamp(input.ExpiresAt),
		DestinationAccount: ConvertAccountToGRPCAccount(&input.DestinationAccount),
		FromUser:           ConvertUserToGRPCUser(&input.FromUser),
	}
}

func ConvertGRPCAccountInvitationToAccountInvitation(input *identitysvc.AccountInvitation) *identity.AccountInvitation {
	return &identity.AccountInvitation{
		CreatedAt:          grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		LastUpdatedAt:      grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ArchivedAt:         grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		ToUser:             input.ToUser,
		Status:             input.Status,
		ToEmail:            input.ToEmail,
		StatusNote:         input.StatusNote,
		Token:              input.Token,
		ID:                 input.Id,
		Note:               input.Note,
		ToName:             input.ToName,
		ExpiresAt:          grpcconverters.ConvertPBTimestampToTime(input.ExpiresAt),
		DestinationAccount: *ConvertGRPCAccountToAccount(input.DestinationAccount),
		FromUser:           *ConvertGRPCUserToUser(input.FromUser),
	}
}

func ConvertGRPCUserRegistrationInputToUserRegistrationInput(input *identitysvc.UserRegistrationInput) *identity.UserRegistrationInput {
	return &identity.UserRegistrationInput{
		Birthday:              grpcconverters.ConvertPBTimestampToTimePointer(input.Birthday),
		Password:              input.Password,
		EmailAddress:          input.EmailAddress,
		InvitationToken:       input.InvitationToken,
		InvitationID:          input.InvitationId,
		Username:              input.Username,
		FirstName:             input.FirstName,
		LastName:              input.LastName,
		AccountName:           input.AccountName,
		AcceptedTOS:           input.AcceptedTos,
		AcceptedPrivacyPolicy: input.AcceptedPrivacyPolicy,
	}
}

func ConvertUserRegistrationInputToGRPCUserRegistrationInput(input *identity.UserRegistrationInput) *identitysvc.UserRegistrationInput {
	return &identitysvc.UserRegistrationInput{
		Birthday:              grpcconverters.ConvertTimePointerToPBTimestamp(input.Birthday),
		Password:              input.Password,
		EmailAddress:          input.EmailAddress,
		InvitationToken:       input.InvitationToken,
		InvitationId:          input.InvitationID,
		Username:              input.Username,
		FirstName:             input.FirstName,
		LastName:              input.LastName,
		AccountName:           input.AccountName,
		AcceptedTos:           input.AcceptedTOS,
		AcceptedPrivacyPolicy: input.AcceptedPrivacyPolicy,
	}
}

func ConvertUserCreationResponseToGRPCUserCreationResponse(input *identity.UserCreationResponse) *identitysvc.UserCreationResponse {
	return &identitysvc.UserCreationResponse{
		CreatedAt:       grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		Birthday:        grpcconverters.ConvertTimePointerToPBTimestamp(input.Birthday),
		Username:        input.Username,
		EmailAddress:    input.EmailAddress,
		TwoFactorQrCode: input.TwoFactorQRCode,
		CreatedUserId:   input.CreatedUserID,
		AccountStatus:   input.AccountStatus,
		TwoFactorSecret: input.TwoFactorSecret,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
	}
}

func ConvertGRPCAccountOwnershipTransferInputToAccountOwnershipTransferInput(input *identitysvc.AccountOwnershipTransferInput) *identity.AccountOwnershipTransferInput {
	return &identity.AccountOwnershipTransferInput{
		Reason:       input.Reason,
		CurrentOwner: input.CurrentOwner,
		NewOwner:     input.NewOwner,
	}
}

func convertFloat32PointerToFloat64Pointer(input *float32) *float64 {
	if input == nil {
		return nil
	}
	return new(float64(*input))
}

func ConvertGRPCAccountUpdateRequestInputToAccountUpdateRequestInput(input *identitysvc.AccountUpdateRequestInput) *identity.AccountUpdateRequestInput {
	return &identity.AccountUpdateRequestInput{
		Name:          input.Name,
		ContactPhone:  input.ContactPhone,
		AddressLine1:  input.AddressLine1,
		AddressLine2:  input.AddressLine2,
		City:          input.City,
		State:         input.State,
		ZipCode:       input.ZipCode,
		Country:       input.Country,
		Latitude:      convertFloat32PointerToFloat64Pointer(input.Latitude),
		Longitude:     convertFloat32PointerToFloat64Pointer(input.Longitude),
		BelongsToUser: input.BelongsToUser,
	}
}

func convertFloat64PointerToFloat32Pointer(input *float64) *float32 {
	if input == nil {
		return nil
	}
	return new(float32(*input))
}

func ConvertAccountUpdateRequestInputToGRPCAccountUpdateRequestInput(input *identity.AccountUpdateRequestInput) *identitysvc.AccountUpdateRequestInput {
	return &identitysvc.AccountUpdateRequestInput{
		Name:          input.Name,
		ContactPhone:  input.ContactPhone,
		AddressLine1:  input.AddressLine1,
		AddressLine2:  input.AddressLine2,
		City:          input.City,
		State:         input.State,
		ZipCode:       input.ZipCode,
		Country:       input.Country,
		Latitude:      convertFloat64PointerToFloat32Pointer(input.Latitude),
		Longitude:     convertFloat64PointerToFloat32Pointer(input.Longitude),
		BelongsToUser: input.BelongsToUser,
	}
}

func ConvertGRPCModifyUserPermissionsInputToModifyUserPermissionsInput(input *identitysvc.ModifyUserPermissionsInput) *identity.ModifyUserPermissionsInput {
	return &identity.ModifyUserPermissionsInput{
		Reason:  input.Reason,
		NewRole: input.NewRole,
	}
}

// ConvertUserDataCollectionToGRPCDataCollection converts a domain identity UserDataCollection to a proto DataCollection.
func ConvertUserDataCollectionToGRPCDataCollection(input *identity.UserDataCollection) *identitysvc.DataCollection {
	result := &identitysvc.DataCollection{
		User: ConvertUserToGRPCUser(&input.User),
	}

	for i := range input.Accounts {
		result.Accounts = append(result.Accounts, ConvertAccountToGRPCAccount(&input.Accounts[i]))
	}

	for i := range input.AccountInvitations {
		result.ReceivedInvites = append(result.ReceivedInvites, ConvertAccountInvitationToGRPCAccountInvitation(&input.AccountInvitations[i]))
	}

	return result
}
