//
//  CallerPermissions.swift
//  ios
//
//  What the signed-in person may do in their active account, as the server says: the effective
//  permissions IdentityService/GetPrincipal answers. Controls for the caller's own actions are
//  gated on these rather than on role names, so the client never carries its own copy of the
//  server's role-to-permission table.
//

import Foundation
import PlatformClient

/// The effective permissions of the caller in their active account.
///
/// A permission is a hint for a control, never the check: the server still refuses a call the
/// caller may not make. Role names are still right for labeling or assigning *another*
/// member's role (HouseholdMembersView's badges and picker), which GetPrincipal cannot answer
/// since it only describes the caller.
struct CallerPermissions: Equatable, Sendable {
  /// Permission names, as the backend's policy spells them. Each is the grant on the method
  /// behind the control it gates.
  enum Name {
    /// Editing the account's name and address: IdentityService/UpdateAccount. platform-go
    /// `identity/grpc.PermissionUpdateAccounts`.
    static let updateAccount = "identity.accounts.update"
    /// Sending and cancelling invitations: IdentityService/Invite and CancelInvitation.
    /// platform-go `identity/grpc.PermissionInviteMembers`.
    static let inviteMembers = "identity.invitations.send"
    /// Setting a member's roles and removing them: IdentityService/SetMembershipRoles and
    /// RemoveMembership. platform-go `identity/grpc.PermissionManageMembers`.
    static let manageMembers = "identity.members.manage"
    /// Reviewing submitted recipes: MealPlanningService/UpdateRecipeStatus. The backend's
    /// `authorization.UpdateRecipesStatusPermission`, which only the service-admin role holds.
    static let updateRecipeStatus = "update.recipe_status"
  }

  /// The names the server granted, or nil when it did not say.
  private let granted: Set<String>?

  /// Nothing is known yet: every gate is closed until the server has answered, so no control
  /// appears and then vanishes.
  static let unloaded = CallerPermissions(granted: [])

  init(granted: Set<String>?) {
    self.granted = granted
  }

  /// Reads the permissions off a GetPrincipal response.
  ///
  /// An absent field is a deployment that serves no permissions, which the proto says a client
  /// must not read as "holds nothing". Gates then stay open and the server's refusal is the
  /// answer, as it is for any control.
  init(_ response: Primandproper_Platform_Identity_V1_GetPrincipalResponse) {
    self.granted = response.hasPermissions ? Set(response.permissions.permissions) : nil
  }

  /// Whether the caller may do what `permission` grants.
  func allows(_ permission: String) -> Bool {
    granted?.contains(permission) ?? true
  }
}

extension AuthenticationManager {
  /// Reads the caller's effective permissions in their active account. Read fresh rather than
  /// cached, like the auth status: a role revoked a moment ago should close its controls now.
  func callerPermissions() async throws -> CallerPermissions {
    var request = Primandproper_Platform_Identity_V1_GetPrincipalRequest()
    request.activeAccountID = try await activeAccountID()
    let response = try await authenticatedCall("getPrincipal", idempotent: true) {
      [request] client, metadata, options in
      try await client.identity.getPrincipal(request, metadata: metadata, options: options)
    }
    return CallerPermissions(response)
  }
}
