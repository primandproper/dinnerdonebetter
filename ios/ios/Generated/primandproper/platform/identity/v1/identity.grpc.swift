/// Package primandproper.platform.identity.v1 is the wire schema for the four
/// nouns this module's identity package owns -- users, accounts, memberships and
/// invitations -- and for the operations over them that are more than one write.
///
/// This file is shipped inside the published Go module, and it is the file
/// itself that is shipped -- not a copy for you to keep in sync. A consumer puts
/// the module's proto directory on protoc's path and imports this file by its
/// canonical name, exactly as filtering.proto already works:
///
///	PLATFORM_PROTO := $(shell go list -m -f '{{.Dir}}' github.com/primandproper/platform-go/v14)
///
///	protoc --proto_path proto/ \
///	    --proto_path $(PLATFORM_PROTO)/identity/proto \
///	    --proto_path $(PLATFORM_PROTO)/filtering/proto \
///	    --go_opt=Mprimandproper/platform/identity/v1/identity.proto=github.com/primandproper/platform-go/v14/identity/identitypb \
///	    --go_opt=Mprimandproper/platform/filtering/v1/filtering.proto=github.com/primandproper/primitives-go/v2/filtering/filteringpb \
///	    $(CONSUMER_PROTO_FILES)   # both platform files deliberately absent from that list
///
/// Go links against the bindings this module already generated, in
/// github.com/primandproper/platform-go/v14/identity/identitypb, rather than
/// making a second copy: the server and the typed client here are built against
/// those types, and a consumer generating its own would hold types that look
/// identical and satisfy nothing. Swift, TypeScript and Kotlin have no such
/// bindings to link against and generate this file directly, which is the whole
/// point of shipping the schema rather than a Go package.
///
/// Field numbers are the compatibility promise, and they are the promise across
/// every language a consumer generates into. Numbers are never reused and never
/// repurposed: a field that goes away is reserved.
///
/// # What is not here, and why
///
/// No scope field, anywhere, and the name is reserved so there cannot be one.
/// Every row this schema describes carries a tenancy.Scope, and every read
/// filters on it -- but a scope a client could put in a request is a
/// cross-tenant read hiding behind a request field. The scope comes off the
/// principal the consumer's authentication interceptor resolved, and from
/// nowhere else. See identity/grpc for that seam.
///
/// Reserving the name rather than only saying so is audit.proto's pattern:
/// `reserved "scope";` is a schema protoc refuses to accept a scope field into,
/// in this repository and in a consumer's fork of the file alike, whereas a
/// comment is a request to the next author. It is reserved on all twenty-nine
/// request messages, on the four inputs they are built from, and on the nine
/// messages a response is built from -- a scope on one of those would be
/// answering a client with something the client supplied. The response wrappers
/// hold nothing but those messages and reserve nothing.
///
/// No credentials, in either direction. There is no hashed_password,
/// two_factor_secret or email_address_verification_token on User, and no token
/// on Invitation -- an invitation's token appears only as a request field on the
/// two RPCs that answer one, because that is where it arrives from, on a link.
/// A schema with no field for a secret is a stronger guarantee than a converter
/// that remembers to clear one.
///
/// No credential RPCs either: setting a password, enrolling a second factor and
/// verifying an email address are the sign-in service's, not the directory's,
/// and RegisterRequest carries no password for the same reason. The identity
/// package never hashes -- it stores what an engine produced -- so a plaintext
/// password on this wire would put the choice of hashing engine in the
/// transport.
///
/// SetUserRequiresPasswordChange is the one write that looks like a credential
/// RPC and is not one. It carries no secret in either direction -- it assigns a
/// boolean on a directory row, which the sign-in service reads on its status
/// call and which SignInService.UpdatePassword clears -- so the sentence above
/// does not reach it: there is nothing here that a hashing engine produced. It
/// is an operator write on a directory column and belongs with ArchiveUser,
/// UpdateUserAccountStatus and SetUserServiceRoles, which is where it sits.
///
/// The sign-in service is deliberately not where it lives, even though that is
/// the service that enforces the flag. signin.Directory is the narrowest
/// interface the component holding everybody's passwords can be given, and an
/// interface that could also impose a forced change on any user is one that
/// could be made to.
///
/// Registration here therefore mints the passwordless user that package already
/// treats as first-class. A registration that carries a credential is
/// SignInService.Register, in signin.proto: that service holds the authenticator,
/// hashes what arrives, and comes back through this package's own registration on
/// one transaction. Which of the two a consumer calls is the question of whether
/// the registrant is choosing a password at that moment -- a directory being
/// filled from elsewhere is this one, and somebody signing up is that one.
///
/// No avatar. The media registry is this module's, but identity has no avatar
/// column and joining one is a contract between two packages that has not been
/// designed. It is absent rather than guessed at.

// DO NOT EDIT.
// swift-format-ignore-file
// swiftlint:disable all
//
// Generated by the gRPC Swift generator plugin for the protocol buffer compiler.
// Source: primandproper/platform/identity/v1/identity.proto
//
// For information on using the generated types, please see the documentation:
//   https://github.com/grpc/grpc-swift

import GRPCCore
import GRPCProtobuf

// MARK: - primandproper.platform.identity.v1.IdentityService

/// Namespace containing generated types for the "primandproper.platform.identity.v1.IdentityService" service.
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
internal enum Primandproper_Platform_Identity_V1_IdentityService {
    /// Service descriptor for the "primandproper.platform.identity.v1.IdentityService" service.
    internal static let descriptor = GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService")
    /// Namespace for method metadata.
    internal enum Method {
        /// Namespace for "Register" metadata.
        internal enum Register {
            /// Request type for "Register".
            internal typealias Input = Primandproper_Platform_Identity_V1_RegisterRequest
            /// Response type for "Register".
            internal typealias Output = Primandproper_Platform_Identity_V1_RegisterResponse
            /// Descriptor for "Register".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "Register"
            )
        }
        /// Namespace for "UpdateProfile" metadata.
        internal enum UpdateProfile {
            /// Request type for "UpdateProfile".
            internal typealias Input = Primandproper_Platform_Identity_V1_UpdateProfileRequest
            /// Response type for "UpdateProfile".
            internal typealias Output = Primandproper_Platform_Identity_V1_UpdateProfileResponse
            /// Descriptor for "UpdateProfile".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "UpdateProfile"
            )
        }
        /// Namespace for "UpdateAccount" metadata.
        internal enum UpdateAccount {
            /// Request type for "UpdateAccount".
            internal typealias Input = Primandproper_Platform_Identity_V1_UpdateAccountRequest
            /// Response type for "UpdateAccount".
            internal typealias Output = Primandproper_Platform_Identity_V1_UpdateAccountResponse
            /// Descriptor for "UpdateAccount".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "UpdateAccount"
            )
        }
        /// Namespace for "RecordAgreement" metadata.
        internal enum RecordAgreement {
            /// Request type for "RecordAgreement".
            internal typealias Input = Primandproper_Platform_Identity_V1_RecordAgreementRequest
            /// Response type for "RecordAgreement".
            internal typealias Output = Primandproper_Platform_Identity_V1_RecordAgreementResponse
            /// Descriptor for "RecordAgreement".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "RecordAgreement"
            )
        }
        /// Namespace for "Invite" metadata.
        internal enum Invite {
            /// Request type for "Invite".
            internal typealias Input = Primandproper_Platform_Identity_V1_InviteRequest
            /// Response type for "Invite".
            internal typealias Output = Primandproper_Platform_Identity_V1_InviteResponse
            /// Descriptor for "Invite".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "Invite"
            )
        }
        /// Namespace for "AcceptInvitation" metadata.
        internal enum AcceptInvitation {
            /// Request type for "AcceptInvitation".
            internal typealias Input = Primandproper_Platform_Identity_V1_AcceptInvitationRequest
            /// Response type for "AcceptInvitation".
            internal typealias Output = Primandproper_Platform_Identity_V1_AcceptInvitationResponse
            /// Descriptor for "AcceptInvitation".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "AcceptInvitation"
            )
        }
        /// Namespace for "RejectInvitation" metadata.
        internal enum RejectInvitation {
            /// Request type for "RejectInvitation".
            internal typealias Input = Primandproper_Platform_Identity_V1_RejectInvitationRequest
            /// Response type for "RejectInvitation".
            internal typealias Output = Primandproper_Platform_Identity_V1_RejectInvitationResponse
            /// Descriptor for "RejectInvitation".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "RejectInvitation"
            )
        }
        /// Namespace for "CancelInvitation" metadata.
        internal enum CancelInvitation {
            /// Request type for "CancelInvitation".
            internal typealias Input = Primandproper_Platform_Identity_V1_CancelInvitationRequest
            /// Response type for "CancelInvitation".
            internal typealias Output = Primandproper_Platform_Identity_V1_CancelInvitationResponse
            /// Descriptor for "CancelInvitation".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "CancelInvitation"
            )
        }
        /// Namespace for "CreateAccount" metadata.
        internal enum CreateAccount {
            /// Request type for "CreateAccount".
            internal typealias Input = Primandproper_Platform_Identity_V1_CreateAccountRequest
            /// Response type for "CreateAccount".
            internal typealias Output = Primandproper_Platform_Identity_V1_CreateAccountResponse
            /// Descriptor for "CreateAccount".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "CreateAccount"
            )
        }
        /// Namespace for "TransferAccountOwnership" metadata.
        internal enum TransferAccountOwnership {
            /// Request type for "TransferAccountOwnership".
            internal typealias Input = Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest
            /// Response type for "TransferAccountOwnership".
            internal typealias Output = Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse
            /// Descriptor for "TransferAccountOwnership".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "TransferAccountOwnership"
            )
        }
        /// Namespace for "SetDefaultAccount" metadata.
        internal enum SetDefaultAccount {
            /// Request type for "SetDefaultAccount".
            internal typealias Input = Primandproper_Platform_Identity_V1_SetDefaultAccountRequest
            /// Response type for "SetDefaultAccount".
            internal typealias Output = Primandproper_Platform_Identity_V1_SetDefaultAccountResponse
            /// Descriptor for "SetDefaultAccount".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "SetDefaultAccount"
            )
        }
        /// Namespace for "SetMembershipRoles" metadata.
        internal enum SetMembershipRoles {
            /// Request type for "SetMembershipRoles".
            internal typealias Input = Primandproper_Platform_Identity_V1_SetMembershipRolesRequest
            /// Response type for "SetMembershipRoles".
            internal typealias Output = Primandproper_Platform_Identity_V1_SetMembershipRolesResponse
            /// Descriptor for "SetMembershipRoles".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "SetMembershipRoles"
            )
        }
        /// Namespace for "RemoveMembership" metadata.
        internal enum RemoveMembership {
            /// Request type for "RemoveMembership".
            internal typealias Input = Primandproper_Platform_Identity_V1_RemoveMembershipRequest
            /// Response type for "RemoveMembership".
            internal typealias Output = Primandproper_Platform_Identity_V1_RemoveMembershipResponse
            /// Descriptor for "RemoveMembership".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "RemoveMembership"
            )
        }
        /// Namespace for "ArchiveUser" metadata.
        internal enum ArchiveUser {
            /// Request type for "ArchiveUser".
            internal typealias Input = Primandproper_Platform_Identity_V1_ArchiveUserRequest
            /// Response type for "ArchiveUser".
            internal typealias Output = Primandproper_Platform_Identity_V1_ArchiveUserResponse
            /// Descriptor for "ArchiveUser".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "ArchiveUser"
            )
        }
        /// Namespace for "ArchiveAccount" metadata.
        internal enum ArchiveAccount {
            /// Request type for "ArchiveAccount".
            internal typealias Input = Primandproper_Platform_Identity_V1_ArchiveAccountRequest
            /// Response type for "ArchiveAccount".
            internal typealias Output = Primandproper_Platform_Identity_V1_ArchiveAccountResponse
            /// Descriptor for "ArchiveAccount".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "ArchiveAccount"
            )
        }
        /// Namespace for "UpdateUserAccountStatus" metadata.
        internal enum UpdateUserAccountStatus {
            /// Request type for "UpdateUserAccountStatus".
            internal typealias Input = Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest
            /// Response type for "UpdateUserAccountStatus".
            internal typealias Output = Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse
            /// Descriptor for "UpdateUserAccountStatus".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "UpdateUserAccountStatus"
            )
        }
        /// Namespace for "SetUserServiceRoles" metadata.
        internal enum SetUserServiceRoles {
            /// Request type for "SetUserServiceRoles".
            internal typealias Input = Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest
            /// Response type for "SetUserServiceRoles".
            internal typealias Output = Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse
            /// Descriptor for "SetUserServiceRoles".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "SetUserServiceRoles"
            )
        }
        /// Namespace for "SetUserRequiresPasswordChange" metadata.
        internal enum SetUserRequiresPasswordChange {
            /// Request type for "SetUserRequiresPasswordChange".
            internal typealias Input = Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest
            /// Response type for "SetUserRequiresPasswordChange".
            internal typealias Output = Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse
            /// Descriptor for "SetUserRequiresPasswordChange".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "SetUserRequiresPasswordChange"
            )
        }
        /// Namespace for "GetPrincipal" metadata.
        internal enum GetPrincipal {
            /// Request type for "GetPrincipal".
            internal typealias Input = Primandproper_Platform_Identity_V1_GetPrincipalRequest
            /// Response type for "GetPrincipal".
            internal typealias Output = Primandproper_Platform_Identity_V1_GetPrincipalResponse
            /// Descriptor for "GetPrincipal".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "GetPrincipal"
            )
        }
        /// Namespace for "GetUser" metadata.
        internal enum GetUser {
            /// Request type for "GetUser".
            internal typealias Input = Primandproper_Platform_Identity_V1_GetUserRequest
            /// Response type for "GetUser".
            internal typealias Output = Primandproper_Platform_Identity_V1_GetUserResponse
            /// Descriptor for "GetUser".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "GetUser"
            )
        }
        /// Namespace for "ListUsers" metadata.
        internal enum ListUsers {
            /// Request type for "ListUsers".
            internal typealias Input = Primandproper_Platform_Identity_V1_ListUsersRequest
            /// Response type for "ListUsers".
            internal typealias Output = Primandproper_Platform_Identity_V1_ListUsersResponse
            /// Descriptor for "ListUsers".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "ListUsers"
            )
        }
        /// Namespace for "SearchUsersByUsername" metadata.
        internal enum SearchUsersByUsername {
            /// Request type for "SearchUsersByUsername".
            internal typealias Input = Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest
            /// Response type for "SearchUsersByUsername".
            internal typealias Output = Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse
            /// Descriptor for "SearchUsersByUsername".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "SearchUsersByUsername"
            )
        }
        /// Namespace for "GetAccount" metadata.
        internal enum GetAccount {
            /// Request type for "GetAccount".
            internal typealias Input = Primandproper_Platform_Identity_V1_GetAccountRequest
            /// Response type for "GetAccount".
            internal typealias Output = Primandproper_Platform_Identity_V1_GetAccountResponse
            /// Descriptor for "GetAccount".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "GetAccount"
            )
        }
        /// Namespace for "ListAccounts" metadata.
        internal enum ListAccounts {
            /// Request type for "ListAccounts".
            internal typealias Input = Primandproper_Platform_Identity_V1_ListAccountsRequest
            /// Response type for "ListAccounts".
            internal typealias Output = Primandproper_Platform_Identity_V1_ListAccountsResponse
            /// Descriptor for "ListAccounts".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "ListAccounts"
            )
        }
        /// Namespace for "ListAccountsForUser" metadata.
        internal enum ListAccountsForUser {
            /// Request type for "ListAccountsForUser".
            internal typealias Input = Primandproper_Platform_Identity_V1_ListAccountsForUserRequest
            /// Response type for "ListAccountsForUser".
            internal typealias Output = Primandproper_Platform_Identity_V1_ListAccountsForUserResponse
            /// Descriptor for "ListAccountsForUser".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "ListAccountsForUser"
            )
        }
        /// Namespace for "GetMembership" metadata.
        internal enum GetMembership {
            /// Request type for "GetMembership".
            internal typealias Input = Primandproper_Platform_Identity_V1_GetMembershipRequest
            /// Response type for "GetMembership".
            internal typealias Output = Primandproper_Platform_Identity_V1_GetMembershipResponse
            /// Descriptor for "GetMembership".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "GetMembership"
            )
        }
        /// Namespace for "ListMembershipsForUser" metadata.
        internal enum ListMembershipsForUser {
            /// Request type for "ListMembershipsForUser".
            internal typealias Input = Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest
            /// Response type for "ListMembershipsForUser".
            internal typealias Output = Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse
            /// Descriptor for "ListMembershipsForUser".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "ListMembershipsForUser"
            )
        }
        /// Namespace for "ListAccountMembers" metadata.
        internal enum ListAccountMembers {
            /// Request type for "ListAccountMembers".
            internal typealias Input = Primandproper_Platform_Identity_V1_ListAccountMembersRequest
            /// Response type for "ListAccountMembers".
            internal typealias Output = Primandproper_Platform_Identity_V1_ListAccountMembersResponse
            /// Descriptor for "ListAccountMembers".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "ListAccountMembers"
            )
        }
        /// Namespace for "GetInvitation" metadata.
        internal enum GetInvitation {
            /// Request type for "GetInvitation".
            internal typealias Input = Primandproper_Platform_Identity_V1_GetInvitationRequest
            /// Response type for "GetInvitation".
            internal typealias Output = Primandproper_Platform_Identity_V1_GetInvitationResponse
            /// Descriptor for "GetInvitation".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "GetInvitation"
            )
        }
        /// Namespace for "ListInvitationsFromUser" metadata.
        internal enum ListInvitationsFromUser {
            /// Request type for "ListInvitationsFromUser".
            internal typealias Input = Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest
            /// Response type for "ListInvitationsFromUser".
            internal typealias Output = Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse
            /// Descriptor for "ListInvitationsFromUser".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "ListInvitationsFromUser"
            )
        }
        /// Namespace for "ListInvitationsForEmailAddress" metadata.
        internal enum ListInvitationsForEmailAddress {
            /// Request type for "ListInvitationsForEmailAddress".
            internal typealias Input = Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest
            /// Response type for "ListInvitationsForEmailAddress".
            internal typealias Output = Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse
            /// Descriptor for "ListInvitationsForEmailAddress".
            internal static let descriptor = GRPCCore.MethodDescriptor(
                service: GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService"),
                method: "ListInvitationsForEmailAddress"
            )
        }
        /// Descriptors for all methods in the "primandproper.platform.identity.v1.IdentityService" service.
        internal static let descriptors: [GRPCCore.MethodDescriptor] = [
            Register.descriptor,
            UpdateProfile.descriptor,
            UpdateAccount.descriptor,
            RecordAgreement.descriptor,
            Invite.descriptor,
            AcceptInvitation.descriptor,
            RejectInvitation.descriptor,
            CancelInvitation.descriptor,
            CreateAccount.descriptor,
            TransferAccountOwnership.descriptor,
            SetDefaultAccount.descriptor,
            SetMembershipRoles.descriptor,
            RemoveMembership.descriptor,
            ArchiveUser.descriptor,
            ArchiveAccount.descriptor,
            UpdateUserAccountStatus.descriptor,
            SetUserServiceRoles.descriptor,
            SetUserRequiresPasswordChange.descriptor,
            GetPrincipal.descriptor,
            GetUser.descriptor,
            ListUsers.descriptor,
            SearchUsersByUsername.descriptor,
            GetAccount.descriptor,
            ListAccounts.descriptor,
            ListAccountsForUser.descriptor,
            GetMembership.descriptor,
            ListMembershipsForUser.descriptor,
            ListAccountMembers.descriptor,
            GetInvitation.descriptor,
            ListInvitationsFromUser.descriptor,
            ListInvitationsForEmailAddress.descriptor
        ]
    }
}

@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
extension GRPCCore.ServiceDescriptor {
    /// Service descriptor for the "primandproper.platform.identity.v1.IdentityService" service.
    internal static let primandproper_platform_identity_v1_IdentityService = GRPCCore.ServiceDescriptor(fullyQualifiedService: "primandproper.platform.identity.v1.IdentityService")
}

// MARK: primandproper.platform.identity.v1.IdentityService (client)

@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
extension Primandproper_Platform_Identity_V1_IdentityService {
    /// Generated client protocol for the "primandproper.platform.identity.v1.IdentityService" service.
    ///
    /// You don't need to implement this protocol directly, use the generated
    /// implementation, ``Client``.
    ///
    /// > Source IDL Documentation:
    /// >
    /// > IdentityService is the directory on the wire: the four nouns, their
    /// > lifecycle, and the reads a client needs to render them.
    /// > 
    /// > Every write here is one call into identity.Service, which means one
    /// > transaction with the consumer's own hook inside it -- the audit entry, the
    /// > outbox row, the search stamp. Nothing on this service writes outside that
    /// > seam.
    /// > 
    /// > Nothing here decides who may call it. The permissions each method requires
    /// > are a default fragment the server package ships and the consumer composes
    /// > into its own policy, and who is calling is resolved from the context by the
    /// > consumer's authentication interceptor.
    internal protocol ClientProtocol: Sendable {
        /// Call the "Register" method.
        ///
        /// > Source IDL Documentation:
        /// >
        /// > The writes.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RegisterRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_RegisterRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_RegisterResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func register<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RegisterRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_RegisterRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_RegisterResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RegisterResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "UpdateProfile" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_UpdateProfileRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_UpdateProfileRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_UpdateProfileResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func updateProfile<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateProfileRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_UpdateProfileRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_UpdateProfileResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateProfileResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "UpdateAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_UpdateAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_UpdateAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_UpdateAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func updateAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_UpdateAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_UpdateAccountResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateAccountResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "RecordAgreement" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RecordAgreementRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_RecordAgreementRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_RecordAgreementResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func recordAgreement<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RecordAgreementRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_RecordAgreementRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_RecordAgreementResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RecordAgreementResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "Invite" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_InviteRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_InviteRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_InviteResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func invite<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_InviteRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_InviteRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_InviteResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_InviteResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "AcceptInvitation" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_AcceptInvitationRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_AcceptInvitationRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_AcceptInvitationResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func acceptInvitation<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_AcceptInvitationRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_AcceptInvitationRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_AcceptInvitationResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_AcceptInvitationResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "RejectInvitation" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RejectInvitationRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_RejectInvitationRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_RejectInvitationResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func rejectInvitation<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RejectInvitationRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_RejectInvitationRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_RejectInvitationResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RejectInvitationResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "CancelInvitation" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_CancelInvitationRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_CancelInvitationRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_CancelInvitationResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func cancelInvitation<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_CancelInvitationRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_CancelInvitationRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_CancelInvitationResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_CancelInvitationResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "CreateAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_CreateAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_CreateAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_CreateAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func createAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_CreateAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_CreateAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_CreateAccountResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_CreateAccountResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "TransferAccountOwnership" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func transferAccountOwnership<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "SetDefaultAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetDefaultAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SetDefaultAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SetDefaultAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func setDefaultAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetDefaultAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SetDefaultAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SetDefaultAccountResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetDefaultAccountResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "SetMembershipRoles" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetMembershipRolesRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SetMembershipRolesRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SetMembershipRolesResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func setMembershipRoles<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetMembershipRolesRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SetMembershipRolesRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SetMembershipRolesResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetMembershipRolesResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "RemoveMembership" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RemoveMembershipRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_RemoveMembershipRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_RemoveMembershipResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func removeMembership<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RemoveMembershipRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_RemoveMembershipRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_RemoveMembershipResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RemoveMembershipResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "ArchiveUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ArchiveUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ArchiveUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ArchiveUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func archiveUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ArchiveUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ArchiveUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ArchiveUserResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ArchiveUserResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "ArchiveAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ArchiveAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ArchiveAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ArchiveAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func archiveAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ArchiveAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ArchiveAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ArchiveAccountResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ArchiveAccountResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "UpdateUserAccountStatus" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func updateUserAccountStatus<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "SetUserServiceRoles" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func setUserServiceRoles<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "SetUserRequiresPasswordChange" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func setUserRequiresPasswordChange<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "GetPrincipal" method.
        ///
        /// > Source IDL Documentation:
        /// >
        /// > The reads.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetPrincipalRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetPrincipalRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetPrincipalResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func getPrincipal<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetPrincipalRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetPrincipalRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetPrincipalResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetPrincipalResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "GetUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func getUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetUserResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetUserResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "ListUsers" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListUsersRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListUsersRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListUsersResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func listUsers<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListUsersRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListUsersRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListUsersResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListUsersResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "SearchUsersByUsername" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func searchUsersByUsername<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "GetAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func getAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetAccountResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetAccountResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "ListAccounts" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListAccountsRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListAccountsRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListAccountsResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func listAccounts<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountsRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListAccountsRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListAccountsResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountsResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "ListAccountsForUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListAccountsForUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListAccountsForUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListAccountsForUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func listAccountsForUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountsForUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListAccountsForUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListAccountsForUserResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountsForUserResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "GetMembership" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetMembershipRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetMembershipRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetMembershipResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func getMembership<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetMembershipRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetMembershipRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetMembershipResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetMembershipResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "ListMembershipsForUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func listMembershipsForUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "ListAccountMembers" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListAccountMembersRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListAccountMembersRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListAccountMembersResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func listAccountMembers<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountMembersRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListAccountMembersRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListAccountMembersResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountMembersResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "GetInvitation" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetInvitationRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetInvitationRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetInvitationResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func getInvitation<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetInvitationRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetInvitationRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetInvitationResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetInvitationResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "ListInvitationsFromUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func listInvitationsFromUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable

        /// Call the "ListInvitationsForEmailAddress" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        func listInvitationsForEmailAddress<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse>,
            options: GRPCCore.CallOptions,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse>) async throws -> Result
        ) async throws -> Result where Result: Sendable
    }

    /// Generated client for the "primandproper.platform.identity.v1.IdentityService" service.
    ///
    /// The ``Client`` provides an implementation of ``ClientProtocol`` which wraps
    /// a `GRPCCore.GRPCCClient`. The underlying `GRPCClient` provides the long-lived
    /// means of communication with the remote peer.
    ///
    /// > Source IDL Documentation:
    /// >
    /// > IdentityService is the directory on the wire: the four nouns, their
    /// > lifecycle, and the reads a client needs to render them.
    /// > 
    /// > Every write here is one call into identity.Service, which means one
    /// > transaction with the consumer's own hook inside it -- the audit entry, the
    /// > outbox row, the search stamp. Nothing on this service writes outside that
    /// > seam.
    /// > 
    /// > Nothing here decides who may call it. The permissions each method requires
    /// > are a default fragment the server package ships and the consumer composes
    /// > into its own policy, and who is calling is resolved from the context by the
    /// > consumer's authentication interceptor.
    internal struct Client<Transport>: ClientProtocol where Transport: GRPCCore.ClientTransport {
        private let client: GRPCCore.GRPCClient<Transport>

        /// Creates a new client wrapping the provided `GRPCCore.GRPCClient`.
        ///
        /// - Parameters:
        ///   - client: A `GRPCCore.GRPCClient` providing a communication channel to the service.
        internal init(wrapping client: GRPCCore.GRPCClient<Transport>) {
            self.client = client
        }

        /// Call the "Register" method.
        ///
        /// > Source IDL Documentation:
        /// >
        /// > The writes.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RegisterRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_RegisterRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_RegisterResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func register<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RegisterRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_RegisterRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_RegisterResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RegisterResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.Register.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "UpdateProfile" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_UpdateProfileRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_UpdateProfileRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_UpdateProfileResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func updateProfile<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateProfileRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_UpdateProfileRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_UpdateProfileResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateProfileResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.UpdateProfile.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "UpdateAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_UpdateAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_UpdateAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_UpdateAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func updateAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_UpdateAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_UpdateAccountResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateAccountResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.UpdateAccount.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "RecordAgreement" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RecordAgreementRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_RecordAgreementRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_RecordAgreementResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func recordAgreement<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RecordAgreementRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_RecordAgreementRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_RecordAgreementResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RecordAgreementResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.RecordAgreement.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "Invite" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_InviteRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_InviteRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_InviteResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func invite<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_InviteRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_InviteRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_InviteResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_InviteResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.Invite.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "AcceptInvitation" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_AcceptInvitationRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_AcceptInvitationRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_AcceptInvitationResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func acceptInvitation<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_AcceptInvitationRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_AcceptInvitationRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_AcceptInvitationResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_AcceptInvitationResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.AcceptInvitation.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "RejectInvitation" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RejectInvitationRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_RejectInvitationRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_RejectInvitationResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func rejectInvitation<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RejectInvitationRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_RejectInvitationRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_RejectInvitationResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RejectInvitationResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.RejectInvitation.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "CancelInvitation" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_CancelInvitationRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_CancelInvitationRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_CancelInvitationResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func cancelInvitation<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_CancelInvitationRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_CancelInvitationRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_CancelInvitationResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_CancelInvitationResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.CancelInvitation.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "CreateAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_CreateAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_CreateAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_CreateAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func createAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_CreateAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_CreateAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_CreateAccountResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_CreateAccountResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.CreateAccount.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "TransferAccountOwnership" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func transferAccountOwnership<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.TransferAccountOwnership.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "SetDefaultAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetDefaultAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SetDefaultAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SetDefaultAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func setDefaultAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetDefaultAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SetDefaultAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SetDefaultAccountResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetDefaultAccountResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.SetDefaultAccount.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "SetMembershipRoles" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetMembershipRolesRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SetMembershipRolesRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SetMembershipRolesResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func setMembershipRoles<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetMembershipRolesRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SetMembershipRolesRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SetMembershipRolesResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetMembershipRolesResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.SetMembershipRoles.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "RemoveMembership" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RemoveMembershipRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_RemoveMembershipRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_RemoveMembershipResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func removeMembership<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RemoveMembershipRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_RemoveMembershipRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_RemoveMembershipResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RemoveMembershipResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.RemoveMembership.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "ArchiveUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ArchiveUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ArchiveUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ArchiveUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func archiveUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ArchiveUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ArchiveUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ArchiveUserResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ArchiveUserResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.ArchiveUser.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "ArchiveAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ArchiveAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ArchiveAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ArchiveAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func archiveAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ArchiveAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ArchiveAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ArchiveAccountResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ArchiveAccountResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.ArchiveAccount.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "UpdateUserAccountStatus" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func updateUserAccountStatus<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.UpdateUserAccountStatus.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "SetUserServiceRoles" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func setUserServiceRoles<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.SetUserServiceRoles.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "SetUserRequiresPasswordChange" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func setUserRequiresPasswordChange<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.SetUserRequiresPasswordChange.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "GetPrincipal" method.
        ///
        /// > Source IDL Documentation:
        /// >
        /// > The reads.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetPrincipalRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetPrincipalRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetPrincipalResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func getPrincipal<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetPrincipalRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetPrincipalRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetPrincipalResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetPrincipalResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.GetPrincipal.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "GetUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func getUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetUserResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetUserResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.GetUser.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "ListUsers" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListUsersRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListUsersRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListUsersResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func listUsers<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListUsersRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListUsersRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListUsersResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListUsersResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.ListUsers.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "SearchUsersByUsername" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func searchUsersByUsername<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.SearchUsersByUsername.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "GetAccount" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetAccountRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetAccountRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetAccountResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func getAccount<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetAccountRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetAccountRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetAccountResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetAccountResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.GetAccount.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "ListAccounts" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListAccountsRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListAccountsRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListAccountsResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func listAccounts<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountsRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListAccountsRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListAccountsResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountsResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.ListAccounts.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "ListAccountsForUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListAccountsForUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListAccountsForUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListAccountsForUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func listAccountsForUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountsForUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListAccountsForUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListAccountsForUserResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountsForUserResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.ListAccountsForUser.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "GetMembership" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetMembershipRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetMembershipRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetMembershipResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func getMembership<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetMembershipRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetMembershipRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetMembershipResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetMembershipResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.GetMembership.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "ListMembershipsForUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func listMembershipsForUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.ListMembershipsForUser.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "ListAccountMembers" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListAccountMembersRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListAccountMembersRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListAccountMembersResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func listAccountMembers<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountMembersRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListAccountMembersRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListAccountMembersResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountMembersResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.ListAccountMembers.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "GetInvitation" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetInvitationRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_GetInvitationRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_GetInvitationResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func getInvitation<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetInvitationRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_GetInvitationRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_GetInvitationResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetInvitationResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.GetInvitation.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "ListInvitationsFromUser" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func listInvitationsFromUser<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.ListInvitationsFromUser.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }

        /// Call the "ListInvitationsForEmailAddress" method.
        ///
        /// - Parameters:
        ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest` message.
        ///   - serializer: A serializer for `Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest` messages.
        ///   - deserializer: A deserializer for `Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse` messages.
        ///   - options: Options to apply to this RPC.
        ///   - handleResponse: A closure which handles the response, the result of which is
        ///       returned to the caller. Returning from the closure will cancel the RPC if it
        ///       hasn't already finished.
        /// - Returns: The result of `handleResponse`.
        internal func listInvitationsForEmailAddress<Result>(
            request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest>,
            serializer: some GRPCCore.MessageSerializer<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest>,
            deserializer: some GRPCCore.MessageDeserializer<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse>,
            options: GRPCCore.CallOptions = .defaults,
            onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse>) async throws -> Result = { response in
                try response.message
            }
        ) async throws -> Result where Result: Sendable {
            try await self.client.unary(
                request: request,
                descriptor: Primandproper_Platform_Identity_V1_IdentityService.Method.ListInvitationsForEmailAddress.descriptor,
                serializer: serializer,
                deserializer: deserializer,
                options: options,
                onResponse: handleResponse
            )
        }
    }
}

// Helpers providing default arguments to 'ClientProtocol' methods.
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
extension Primandproper_Platform_Identity_V1_IdentityService.ClientProtocol {
    /// Call the "Register" method.
    ///
    /// > Source IDL Documentation:
    /// >
    /// > The writes.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RegisterRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func register<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RegisterRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RegisterResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.register(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_RegisterRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_RegisterResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "UpdateProfile" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_UpdateProfileRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func updateProfile<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateProfileRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateProfileResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.updateProfile(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_UpdateProfileRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_UpdateProfileResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "UpdateAccount" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_UpdateAccountRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func updateAccount<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateAccountRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.updateAccount(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_UpdateAccountRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_UpdateAccountResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "RecordAgreement" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RecordAgreementRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func recordAgreement<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RecordAgreementRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RecordAgreementResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.recordAgreement(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_RecordAgreementRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_RecordAgreementResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "Invite" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_InviteRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func invite<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_InviteRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_InviteResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.invite(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_InviteRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_InviteResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "AcceptInvitation" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_AcceptInvitationRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func acceptInvitation<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_AcceptInvitationRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_AcceptInvitationResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.acceptInvitation(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_AcceptInvitationRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_AcceptInvitationResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "RejectInvitation" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RejectInvitationRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func rejectInvitation<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RejectInvitationRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RejectInvitationResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.rejectInvitation(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_RejectInvitationRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_RejectInvitationResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "CancelInvitation" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_CancelInvitationRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func cancelInvitation<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_CancelInvitationRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_CancelInvitationResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.cancelInvitation(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_CancelInvitationRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_CancelInvitationResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "CreateAccount" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_CreateAccountRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func createAccount<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_CreateAccountRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_CreateAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.createAccount(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_CreateAccountRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_CreateAccountResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "TransferAccountOwnership" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func transferAccountOwnership<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.transferAccountOwnership(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SetDefaultAccount" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetDefaultAccountRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func setDefaultAccount<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetDefaultAccountRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetDefaultAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.setDefaultAccount(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_SetDefaultAccountRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_SetDefaultAccountResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SetMembershipRoles" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetMembershipRolesRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func setMembershipRoles<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetMembershipRolesRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetMembershipRolesResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.setMembershipRoles(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_SetMembershipRolesRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_SetMembershipRolesResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "RemoveMembership" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_RemoveMembershipRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func removeMembership<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RemoveMembershipRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RemoveMembershipResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.removeMembership(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_RemoveMembershipRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_RemoveMembershipResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ArchiveUser" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ArchiveUserRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func archiveUser<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ArchiveUserRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ArchiveUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.archiveUser(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_ArchiveUserRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_ArchiveUserResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ArchiveAccount" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ArchiveAccountRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func archiveAccount<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ArchiveAccountRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ArchiveAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.archiveAccount(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_ArchiveAccountRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_ArchiveAccountResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "UpdateUserAccountStatus" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func updateUserAccountStatus<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.updateUserAccountStatus(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SetUserServiceRoles" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func setUserServiceRoles<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.setUserServiceRoles(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SetUserRequiresPasswordChange" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func setUserRequiresPasswordChange<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.setUserRequiresPasswordChange(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetPrincipal" method.
    ///
    /// > Source IDL Documentation:
    /// >
    /// > The reads.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetPrincipalRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getPrincipal<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetPrincipalRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetPrincipalResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.getPrincipal(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_GetPrincipalRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_GetPrincipalResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetUser" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetUserRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getUser<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetUserRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.getUser(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_GetUserRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_GetUserResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListUsers" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListUsersRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listUsers<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListUsersRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListUsersResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.listUsers(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_ListUsersRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_ListUsersResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SearchUsersByUsername" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func searchUsersByUsername<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.searchUsersByUsername(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetAccount" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetAccountRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getAccount<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetAccountRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.getAccount(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_GetAccountRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_GetAccountResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListAccounts" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListAccountsRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listAccounts<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountsRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountsResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.listAccounts(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_ListAccountsRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_ListAccountsResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListAccountsForUser" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListAccountsForUserRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listAccountsForUser<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountsForUserRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountsForUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.listAccountsForUser(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_ListAccountsForUserRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_ListAccountsForUserResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetMembership" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetMembershipRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getMembership<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetMembershipRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetMembershipResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.getMembership(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_GetMembershipRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_GetMembershipResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListMembershipsForUser" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listMembershipsForUser<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.listMembershipsForUser(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListAccountMembers" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListAccountMembersRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listAccountMembers<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountMembersRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountMembersResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.listAccountMembers(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_ListAccountMembersRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_ListAccountMembersResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetInvitation" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_GetInvitationRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getInvitation<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetInvitationRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetInvitationResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.getInvitation(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_GetInvitationRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_GetInvitationResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListInvitationsFromUser" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listInvitationsFromUser<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.listInvitationsFromUser(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListInvitationsForEmailAddress" method.
    ///
    /// - Parameters:
    ///   - request: A request containing a single `Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest` message.
    ///   - options: Options to apply to this RPC.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listInvitationsForEmailAddress<Result>(
        request: GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest>,
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        try await self.listInvitationsForEmailAddress(
            request: request,
            serializer: GRPCProtobuf.ProtobufSerializer<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest>(),
            deserializer: GRPCProtobuf.ProtobufDeserializer<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse>(),
            options: options,
            onResponse: handleResponse
        )
    }
}

// Helpers providing sugared APIs for 'ClientProtocol' methods.
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
extension Primandproper_Platform_Identity_V1_IdentityService.ClientProtocol {
    /// Call the "Register" method.
    ///
    /// > Source IDL Documentation:
    /// >
    /// > The writes.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func register<Result>(
        _ message: Primandproper_Platform_Identity_V1_RegisterRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RegisterResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RegisterRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.register(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "UpdateProfile" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func updateProfile<Result>(
        _ message: Primandproper_Platform_Identity_V1_UpdateProfileRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateProfileResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateProfileRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.updateProfile(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "UpdateAccount" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func updateAccount<Result>(
        _ message: Primandproper_Platform_Identity_V1_UpdateAccountRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateAccountRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.updateAccount(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "RecordAgreement" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func recordAgreement<Result>(
        _ message: Primandproper_Platform_Identity_V1_RecordAgreementRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RecordAgreementResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RecordAgreementRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.recordAgreement(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "Invite" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func invite<Result>(
        _ message: Primandproper_Platform_Identity_V1_InviteRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_InviteResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_InviteRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.invite(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "AcceptInvitation" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func acceptInvitation<Result>(
        _ message: Primandproper_Platform_Identity_V1_AcceptInvitationRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_AcceptInvitationResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_AcceptInvitationRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.acceptInvitation(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "RejectInvitation" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func rejectInvitation<Result>(
        _ message: Primandproper_Platform_Identity_V1_RejectInvitationRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RejectInvitationResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RejectInvitationRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.rejectInvitation(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "CancelInvitation" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func cancelInvitation<Result>(
        _ message: Primandproper_Platform_Identity_V1_CancelInvitationRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_CancelInvitationResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_CancelInvitationRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.cancelInvitation(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "CreateAccount" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func createAccount<Result>(
        _ message: Primandproper_Platform_Identity_V1_CreateAccountRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_CreateAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_CreateAccountRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.createAccount(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "TransferAccountOwnership" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func transferAccountOwnership<Result>(
        _ message: Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_TransferAccountOwnershipResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_TransferAccountOwnershipRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.transferAccountOwnership(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SetDefaultAccount" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func setDefaultAccount<Result>(
        _ message: Primandproper_Platform_Identity_V1_SetDefaultAccountRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetDefaultAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetDefaultAccountRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.setDefaultAccount(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SetMembershipRoles" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func setMembershipRoles<Result>(
        _ message: Primandproper_Platform_Identity_V1_SetMembershipRolesRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetMembershipRolesResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetMembershipRolesRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.setMembershipRoles(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "RemoveMembership" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func removeMembership<Result>(
        _ message: Primandproper_Platform_Identity_V1_RemoveMembershipRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_RemoveMembershipResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_RemoveMembershipRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.removeMembership(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ArchiveUser" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func archiveUser<Result>(
        _ message: Primandproper_Platform_Identity_V1_ArchiveUserRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ArchiveUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ArchiveUserRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.archiveUser(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ArchiveAccount" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func archiveAccount<Result>(
        _ message: Primandproper_Platform_Identity_V1_ArchiveAccountRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ArchiveAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ArchiveAccountRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.archiveAccount(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "UpdateUserAccountStatus" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func updateUserAccountStatus<Result>(
        _ message: Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_UpdateUserAccountStatusRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.updateUserAccountStatus(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SetUserServiceRoles" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func setUserServiceRoles<Result>(
        _ message: Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetUserServiceRolesResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetUserServiceRolesRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.setUserServiceRoles(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SetUserRequiresPasswordChange" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func setUserRequiresPasswordChange<Result>(
        _ message: Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SetUserRequiresPasswordChangeRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.setUserRequiresPasswordChange(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetPrincipal" method.
    ///
    /// > Source IDL Documentation:
    /// >
    /// > The reads.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getPrincipal<Result>(
        _ message: Primandproper_Platform_Identity_V1_GetPrincipalRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetPrincipalResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetPrincipalRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.getPrincipal(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetUser" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getUser<Result>(
        _ message: Primandproper_Platform_Identity_V1_GetUserRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetUserRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.getUser(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListUsers" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listUsers<Result>(
        _ message: Primandproper_Platform_Identity_V1_ListUsersRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListUsersResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListUsersRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.listUsers(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "SearchUsersByUsername" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func searchUsersByUsername<Result>(
        _ message: Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_SearchUsersByUsernameResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_SearchUsersByUsernameRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.searchUsersByUsername(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetAccount" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getAccount<Result>(
        _ message: Primandproper_Platform_Identity_V1_GetAccountRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetAccountResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetAccountRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.getAccount(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListAccounts" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listAccounts<Result>(
        _ message: Primandproper_Platform_Identity_V1_ListAccountsRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountsResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountsRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.listAccounts(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListAccountsForUser" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listAccountsForUser<Result>(
        _ message: Primandproper_Platform_Identity_V1_ListAccountsForUserRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountsForUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountsForUserRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.listAccountsForUser(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetMembership" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getMembership<Result>(
        _ message: Primandproper_Platform_Identity_V1_GetMembershipRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetMembershipResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetMembershipRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.getMembership(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListMembershipsForUser" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listMembershipsForUser<Result>(
        _ message: Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListMembershipsForUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListMembershipsForUserRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.listMembershipsForUser(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListAccountMembers" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listAccountMembers<Result>(
        _ message: Primandproper_Platform_Identity_V1_ListAccountMembersRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListAccountMembersResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListAccountMembersRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.listAccountMembers(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "GetInvitation" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func getInvitation<Result>(
        _ message: Primandproper_Platform_Identity_V1_GetInvitationRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_GetInvitationResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_GetInvitationRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.getInvitation(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListInvitationsFromUser" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listInvitationsFromUser<Result>(
        _ message: Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListInvitationsFromUserResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.listInvitationsFromUser(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }

    /// Call the "ListInvitationsForEmailAddress" method.
    ///
    /// - Parameters:
    ///   - message: request message to send.
    ///   - metadata: Additional metadata to send, defaults to empty.
    ///   - options: Options to apply to this RPC, defaults to `.defaults`.
    ///   - handleResponse: A closure which handles the response, the result of which is
    ///       returned to the caller. Returning from the closure will cancel the RPC if it
    ///       hasn't already finished.
    /// - Returns: The result of `handleResponse`.
    internal func listInvitationsForEmailAddress<Result>(
        _ message: Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest,
        metadata: GRPCCore.Metadata = [:],
        options: GRPCCore.CallOptions = .defaults,
        onResponse handleResponse: @Sendable @escaping (GRPCCore.ClientResponse<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressResponse>) async throws -> Result = { response in
            try response.message
        }
    ) async throws -> Result where Result: Sendable {
        let request = GRPCCore.ClientRequest<Primandproper_Platform_Identity_V1_ListInvitationsForEmailAddressRequest>(
            message: message,
            metadata: metadata
        )
        return try await self.listInvitationsForEmailAddress(
            request: request,
            options: options,
            onResponse: handleResponse
        )
    }
}