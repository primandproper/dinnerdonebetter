//
//  AccountSettingsViewModel.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import SwiftProtobuf
import SwiftUI

@Observable
@MainActor
// swiftlint:disable:next type_body_length
class AccountSettingsViewModel {
  private struct FetchDataResult {
    let account: Primandproper_Platform_Identity_V1_Account
    let members: [Primandproper_Platform_Identity_V1_MembershipWithUser]
    let user: Auth_GetSelfResponse
    let invitations: [Primandproper_Platform_Identity_V1_Invitation]
    let instrumentOwnerships: [Mealplanning_AccountInstrumentOwnership]
    let validInstruments: [Mealplanning_ValidInstrument]
  }
  // Data
  var account: Primandproper_Platform_Identity_V1_Account?
  /// The household's roster.
  ///
  /// A read of its own rather than a field of the account: an account with thirty members
  /// would otherwise be thirty users on every read of it.
  var members: [Primandproper_Platform_Identity_V1_MembershipWithUser] = []
  var user: Auth_GetSelfResponse?
  var invitations: [Primandproper_Platform_Identity_V1_Invitation] = []
  var instrumentOwnerships: [Mealplanning_AccountInstrumentOwnership] = []
  var validInstruments: [Mealplanning_ValidInstrument] = []

  // Loading states
  var isLoading = false
  var errorMessage: String?
  var errorTitle: String = "Error"
  var errorIcon: String = "exclamationmark.triangle"
  var errorIconColor = DSTheme.Colors.warning
  var isServerDownError = false

  // Form state
  var accountName: String = ""
  var contactPhone: String = ""
  var addressLine1: String = ""
  var addressLine2: String = ""
  var city: String = ""
  var state: String = ""
  var zipCode: String = ""
  var country: String = "USA"

  // Invitation form state
  var invitationEmail: String = ""
  var invitationName: String = ""
  var invitationNote: String = ""

  // Instrument ownership form state
  var newInstrumentValidInstrumentID: String = ""
  var newInstrumentQuantity: UInt32 = 1
  var newInstrumentNotes: String = ""

  // Computed properties
  var isAccountAdmin: Bool {
    guard let membership = currentUserMembership else { return false }
    // A membership carries a set of roles rather than one, because a role is a grant and
    // somebody may hold several.
    return membership.membership.roles.contains("account_admin")
  }

  var currentUserMembership: Primandproper_Platform_Identity_V1_MembershipWithUser? {
    guard let userID = getCurrentUserID() else { return nil }
    return members.first { membership in
      membership.hasUser && membership.user.id == userID
    }
  }

  var currentUserID: String {
    return getCurrentUserID() ?? ""
  }

  private func getCurrentUserID() -> String? {
    guard let user = user, user.hasResult, !user.result.id.isEmpty else {
      return nil
    }
    return user.result.id
  }

  private let authManager: AuthenticationManager

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  func loadData() async {
    isLoading = true
    errorMessage = nil
    errorTitle = "Error"
    errorIcon = "exclamationmark.triangle"
    errorIconColor = DSTheme.Colors.warning
    isServerDownError = false

    do {
      let result = try await fetchAllData()
      self.account = result.account
      self.members = result.members
      self.user = result.user
      self.invitations = result.invitations
      self.instrumentOwnerships = result.instrumentOwnerships
      self.validInstruments = result.validInstruments
      initializeFormFields(from: result.account)
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      let display = ErrorDisplayFormatter.format(error, context: "load data")
      errorMessage = display.message
      errorTitle = display.title
      errorIcon = display.icon
      errorIconColor = display.iconColor
      isServerDownError = ErrorDisplayFormatter.isServerDown(error)
      print("❌ Error loading account settings: \(error)")
    }

    isLoading = false
  }

  private func fetchAllData() async throws -> FetchDataResult {
    async let accountTask = fetchActiveAccount()
    async let membersTask = fetchMembers()
    async let userTask = fetchUser()
    async let invitationsTask = fetchInvitations()
    async let instrumentOwnershipsTask = fetchInstrumentOwnerships()
    async let validInstrumentsTask = fetchValidInstruments()
    return FetchDataResult(
      account: try await accountTask,
      members: try await membersTask,
      user: try await userTask,
      invitations: try await invitationsTask,
      instrumentOwnerships: try await instrumentOwnershipsTask,
      validInstruments: try await validInstrumentsTask
    )
  }

  private func fetchActiveAccount() async throws -> Primandproper_Platform_Identity_V1_Account {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    let accountID = try await getActiveAccountID(clientManager: clientManager, metadata: metadata)
    return try await getAccountDetails(
      accountID: accountID, clientManager: clientManager, metadata: metadata)
  }

  private func getActiveAccountID(
    clientManager: ClientManager<HTTP2ClientTransport.TransportServices>,
    metadata: GRPCCore.Metadata
  ) async throws -> String {
    let authResponse = try await clientManager.client.auth.getActiveAccount(
      Auth_GetActiveAccountRequest(),
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )

    guard authResponse.hasResult, !authResponse.result.id.isEmpty else {
      throw NSError(
        domain: "AccountSettingsViewModel", code: 3,
        userInfo: [NSLocalizedDescriptionKey: "No active account found"])
    }

    return authResponse.result.id
  }

  private func getAccountDetails(
    accountID: String,
    clientManager: ClientManager<HTTP2ClientTransport.TransportServices>,
    metadata: GRPCCore.Metadata
  ) async throws -> Primandproper_Platform_Identity_V1_Account {
    var request = Primandproper_Platform_Identity_V1_GetAccountRequest()
    request.accountID = accountID

    let identityResponse = try await clientManager.client.identity.getAccount(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )

    guard identityResponse.hasAccount else {
      throw NSError(
        domain: "AccountSettingsViewModel", code: 4,
        userInfo: [NSLocalizedDescriptionKey: "Account not found"])
    }

    return identityResponse.account
  }

  // The roster is a read of its own rather than a field of the account: an account with
  // thirty members would otherwise be thirty users on every read of it.
  private func fetchMembers() async throws
    -> [Primandproper_Platform_Identity_V1_MembershipWithUser]
  {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    let accountID = try await getActiveAccountID(clientManager: clientManager, metadata: metadata)

    var request = Primandproper_Platform_Identity_V1_ListAccountMembersRequest()
    request.accountID = accountID
    request.filter = QueryFilterMessage()

    let response = try await clientManager.client.identity.listAccountMembers(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )

    return response.results
  }

  private func fetchUser() async throws -> Auth_GetSelfResponse {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    return try await clientManager.client.auth.getSelf(
      Auth_GetSelfRequest(),
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  private func fetchInvitations() async throws -> [Primandproper_Platform_Identity_V1_Invitation] {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    let filter = QueryFilterMessage()
    var request = Primandproper_Platform_Identity_V1_ListInvitationsFromUserRequest()
    request.filter = filter

    let response = try await clientManager.client.identity.listInvitationsFromUser(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )

    return response.results
  }

  private func fetchInstrumentOwnerships() async throws -> [Mealplanning_AccountInstrumentOwnership]
  {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    var request = Mealplanning_GetAccountInstrumentOwnershipsRequest()
    request.filter = QueryFilterMessage()

    let response = try await clientManager.client.mealPlanning.getAccountInstrumentOwnerships(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )

    return response.results
  }

  private func fetchValidInstruments() async throws -> [Mealplanning_ValidInstrument] {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    var request = Mealplanning_SearchForValidInstrumentsNotOwnedByAccountRequest()
    request.query = ""
    request.filter = QueryFilterMessage()

    let response = try await clientManager.client.mealPlanning
      .searchForValidInstrumentsNotOwnedByAccount(
        request,
        metadata: metadata,
        options: clientManager.defaultCallOptions
      )

    return response.results
  }

  func createInstrumentOwnership() async -> Bool {
    guard validateInstrumentOwnershipInput() else {
      return false
    }

    return await performUpdate {
      try await executeInstrumentOwnershipCreation()
      newInstrumentValidInstrumentID = ""
      newInstrumentQuantity = 1
      newInstrumentNotes = ""
      await loadData()
    } errorMessage: {
      "Failed to add instrument: \($0.localizedDescription)"
    }
  }

  private func validateInstrumentOwnershipInput() -> Bool {
    guard !newInstrumentValidInstrumentID.isEmpty else {
      errorMessage = "Please select an instrument"
      return false
    }

    guard newInstrumentQuantity >= 1 else {
      errorMessage = "Quantity must be at least 1"
      return false
    }

    return true
  }

  private func executeInstrumentOwnershipCreation() async throws {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    var input = Mealplanning_AccountInstrumentOwnershipCreationRequestInput()
    input.validInstrumentID = newInstrumentValidInstrumentID
    input.quantity = newInstrumentQuantity
    input.notes = newInstrumentNotes

    var request = Mealplanning_CreateAccountInstrumentOwnershipRequest()
    request.input = input

    _ = try await clientManager.client.mealPlanning.createAccountInstrumentOwnership(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  func updateInstrumentOwnership(
    ownershipID: String,
    quantity: UInt32?,
    notes: String?
  ) async -> Bool {
    guard quantity ?? 1 >= 1 else {
      errorMessage = "Quantity must be at least 1"
      return false
    }

    return await performUpdate {
      try await executeInstrumentOwnershipUpdate(
        ownershipID: ownershipID,
        quantity: quantity,
        notes: notes
      )
      await loadData()
    } errorMessage: {
      "Failed to update instrument: \($0.localizedDescription)"
    }
  }

  private func executeInstrumentOwnershipUpdate(
    ownershipID: String,
    quantity: UInt32?,
    notes: String?
  ) async throws {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    var input = Mealplanning_AccountInstrumentOwnershipUpdateRequestInput()
    if let quantity = quantity {
      input.quantity = quantity
    }
    if let notes = notes {
      input.notes = notes
    }

    var request = Mealplanning_UpdateAccountInstrumentOwnershipRequest()
    request.accountInstrumentOwnershipID = ownershipID
    request.input = input

    _ = try await clientManager.client.mealPlanning.updateAccountInstrumentOwnership(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  func archiveInstrumentOwnership(ownershipID: String) async -> Bool {
    guard !ownershipID.isEmpty else {
      errorMessage = "Instrument ownership ID is required"
      return false
    }

    return await performUpdate {
      try await executeInstrumentOwnershipArchive(ownershipID: ownershipID)
      await loadData()
    } errorMessage: {
      "Failed to remove instrument: \($0.localizedDescription)"
    }
  }

  private func executeInstrumentOwnershipArchive(ownershipID: String) async throws {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    var request = Mealplanning_ArchiveAccountInstrumentOwnershipRequest()
    request.accountInstrumentOwnershipID = ownershipID

    _ = try await clientManager.client.mealPlanning.archiveAccountInstrumentOwnership(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  func updateAccount() async -> Bool {
    guard let accountID = validateAccountUpdate() else {
      return false
    }

    return await performUpdate {
      try await executeAccountUpdate(accountID: accountID)
      await loadData()
    } errorMessage: {
      "Failed to update account: \($0.localizedDescription)"
    }
  }

  private func validateAccountUpdate() -> String? {
    guard let account = account else {
      errorMessage = "No account loaded"
      return nil
    }

    guard isAccountAdmin else {
      errorMessage = "Only household admins can update household details"
      return nil
    }

    return account.id
  }

  private func executeAccountUpdate(accountID: String) async throws {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    let updateInput = createAccountUpdateInput()
    var request = Primandproper_Platform_Identity_V1_UpdateAccountRequest()
    request.accountID = accountID
    request.input = updateInput

    _ = try await clientManager.client.identity.updateAccount(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  // The address is one value rather than seven fields, because it travels as one: a caller
  // updating an address updates all of it, and an application that never collects one
  // leaves a zero rather than seven empty strings.
  private func createAccountUpdateInput() -> Primandproper_Platform_Identity_V1_AccountUpdateInput {
    var address = Primandproper_Platform_Identity_V1_BillingAddress()
    address.line1 = addressLine1
    address.line2 = addressLine2
    address.city = city
    address.state = state
    address.postalCode = zipCode
    address.country = country
    address.phone = contactPhone

    var updateInput = Primandproper_Platform_Identity_V1_AccountUpdateInput()
    updateInput.name = accountName
    updateInput.billingAddress = address
    return updateInput
  }

  func sendInvitation() async -> Bool {
    guard validateInvitationInput() else {
      return false
    }

    return await performUpdate {
      try await executeInvitationCreation()
      invitationEmail = ""
      invitationName = ""
      invitationNote = ""
      await loadData()
    } errorMessage: {
      "Failed to send invitation: \($0.localizedDescription)"
    }
  }

  private func validateInvitationInput() -> Bool {
    guard account != nil else {
      errorMessage = "No account loaded"
      return false
    }

    guard isAccountAdmin else {
      errorMessage = "Only household admins can send invitations"
      return false
    }

    guard !invitationEmail.isEmpty else {
      errorMessage = "Email address is required"
      return false
    }

    guard validateEmail(invitationEmail) else {
      errorMessage = "Invalid email address"
      return false
    }

    return true
  }

  private func validateEmail(_ email: String) -> Bool {
    let emailRegex = "[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}"
    let emailPredicate = NSPredicate(format: "SELF MATCHES %@", emailRegex)
    return emailPredicate.evaluate(with: email)
  }

  private func executeInvitationCreation() async throws {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    let accountID = try await getActiveAccountID(clientManager: clientManager, metadata: metadata)

    // The account is named on the request rather than taken from the session, and the roles
    // the invitation promises come from here: what somebody was invited to is what they
    // get, and an acceptance cannot ask for more.
    var request = Primandproper_Platform_Identity_V1_InviteRequest()
    request.accountID = accountID
    request.toEmail = invitationEmail
    request.toName = invitationName
    request.note = invitationNote
    request.roles = ["account_member"]

    _ = try await clientManager.client.identity.invite(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  func cancelInvitation(invitationID: String) async -> Bool {
    guard isAccountAdmin else {
      errorMessage = "Only household admins can cancel invitations"
      return false
    }
    guard !invitationID.isEmpty else {
      errorMessage = "Invitation ID is required"
      return false
    }

    return await performUpdate {
      let (clientManager, metadata) = try await getClientManagerAndMetadata()
      // No token: withdrawing is the sender's act, and the secret half of the link is the
      // recipient's. No read returns one.
      var request = Primandproper_Platform_Identity_V1_CancelInvitationRequest()
      request.invitationID = invitationID

      _ = try await clientManager.client.identity.cancelInvitation(
        request,
        metadata: metadata,
        options: clientManager.defaultCallOptions
      )
      await loadData()
    } errorMessage: {
      "Failed to cancel invitation: \($0.localizedDescription)"
    }
  }

  func updateMemberRole(membershipID: String, newRole: String, reason: String) async -> Bool {
    guard let membership = validateMemberRoleUpdate(membershipID: membershipID, reason: reason)
    else {
      return false
    }

    return await performUpdate {
      try await executeMemberRoleUpdate(
        userID: membership.user.id, newRole: newRole, reason: reason)
      await loadData()
    } errorMessage: {
      "Failed to update member role: \($0.localizedDescription)"
    }
  }

  private func validateMemberRoleUpdate(
    membershipID: String, reason: String
  ) -> Primandproper_Platform_Identity_V1_MembershipWithUser? {
    guard isAccountAdmin else {
      errorMessage = "Only household admins can change member roles"
      return nil
    }

    guard !reason.isEmpty else {
      errorMessage = "A reason is required for changing member roles"
      return nil
    }

    guard let membership = members.first(where: { $0.membership.id == membershipID }) else {
      errorMessage = "Member not found"
      return nil
    }

    guard membership.hasUser, !membership.user.id.isEmpty else {
      errorMessage = "User ID not found"
      return nil
    }

    return membership
  }

  private func performUpdate(
    operation: () async throws -> Void,
    errorMessage: (Error) -> String
  ) async -> Bool {
    isLoading = true
    self.errorMessage = nil

    do {
      try await operation()
      isLoading = false
      return true
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      self.errorMessage = errorMessage(error)
      print("❌ Error: \(error)")
      isLoading = false
      return false
    }
  }

  // Roles are replaced rather than merged, and the account is named. A caller adding a role
  // reads the membership and writes the union, which is visible at the call site; a merging
  // setter could not express a revocation at all.
  //
  // The reason is no longer sent. platform records who changed what through the hook this
  // application registers rather than through a field on the request, so a reason here
  // would be a value nothing stored — it is still required of the user, and still shown.
  private func executeMemberRoleUpdate(userID: String, newRole: String, reason: String) async throws
  {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    let accountID = try await getActiveAccountID(clientManager: clientManager, metadata: metadata)

    var request = Primandproper_Platform_Identity_V1_SetMembershipRolesRequest()
    request.accountID = accountID
    request.userID = userID
    request.roles = [newRole]

    _ = try await clientManager.client.identity.setMembershipRoles(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  private func initializeFormFields(from account: Primandproper_Platform_Identity_V1_Account) {
    accountName = account.name
    contactPhone = account.billingAddress.phone
    addressLine1 = account.billingAddress.line1
    addressLine2 = account.billingAddress.line2
    city = account.billingAddress.city
    state = account.billingAddress.state
    zipCode = account.billingAddress.postalCode
    country = account.billingAddress.country.isEmpty ? "USA" : account.billingAddress.country
  }

  private func getClientManagerAndMetadata() async throws -> (
    ClientManager<HTTP2ClientTransport.TransportServices>, GRPCCore.Metadata
  ) {
    guard let clientManager = try? authManager.getClientManager() else {
      throw NSError(
        domain: "AccountSettingsViewModel", code: 1,
        userInfo: [NSLocalizedDescriptionKey: "Failed to get client manager"])
    }

    guard let oauth2Token = await authManager.getOAuth2AccessToken() else {
      throw NSError(
        domain: "AccountSettingsViewModel", code: 2,
        userInfo: [NSLocalizedDescriptionKey: "Failed to get OAuth2 access token"])
    }

    let metadata = clientManager.authenticatedMetadata(accessToken: oauth2Token)
    return (clientManager, metadata)
  }

  // The address is one value rather than seven fields — see createAccountUpdateInput.
  var accountDataHasChanged: Bool {
    guard let account = account else { return false }
    let address = account.billingAddress

    return account.name != accountName
      || address.phone != contactPhone
      || address.line1 != addressLine1
      || address.line2 != addressLine2
      || address.city != city
      || address.state != state
      || address.postalCode != zipCode
      || address.country != country
  }
}
