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
    let account: Identity_Account
    let user: Auth_GetSelfResponse
    let invitations: [Identity_AccountInvitation]
    let instrumentOwnerships: [Mealplanning_AccountInstrumentOwnership]
    let validInstruments: [Mealplanning_ValidInstrument]
  }
  // Data
  var account: Identity_Account?
  var user: Auth_GetSelfResponse?
  var invitations: [Identity_AccountInvitation] = []
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
    guard let userID = getCurrentUserID() else { return false }
    return account?.members.first { membership in
      membership.hasBelongsToUser && membership.belongsToUser.id == userID
    }?.accountRole == "account_admin"
  }

  var currentUserMembership: Identity_AccountUserMembershipWithUser? {
    guard let userID = getCurrentUserID() else { return nil }
    return account?.members.first { membership in
      membership.hasBelongsToUser && membership.belongsToUser.id == userID
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
    async let userTask = fetchUser()
    async let invitationsTask = fetchInvitations()
    async let instrumentOwnershipsTask = fetchInstrumentOwnerships()
    async let validInstrumentsTask = fetchValidInstruments()
    return FetchDataResult(
      account: try await accountTask,
      user: try await userTask,
      invitations: try await invitationsTask,
      instrumentOwnerships: try await instrumentOwnershipsTask,
      validInstruments: try await validInstrumentsTask
    )
  }

  private func fetchActiveAccount() async throws -> Identity_Account {
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
  ) async throws -> Identity_Account {
    var request = Identity_GetAccountRequest()
    request.accountID = accountID

    let identityResponse = try await clientManager.client.identity.getAccount(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )

    guard identityResponse.hasResult else {
      throw NSError(
        domain: "AccountSettingsViewModel", code: 4,
        userInfo: [NSLocalizedDescriptionKey: "Account not found"])
    }

    return identityResponse.result
  }

  private func fetchUser() async throws -> Auth_GetSelfResponse {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    return try await clientManager.client.auth.getSelf(
      Auth_GetSelfRequest(),
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  private func fetchInvitations() async throws -> [Identity_AccountInvitation] {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    let filter = QueryFilterMessage()
    var request = Identity_GetSentAccountInvitationsRequest()
    request.filter = filter

    let response = try await clientManager.client.identity.getSentAccountInvitations(
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
    var request = Identity_UpdateAccountRequest()
    request.accountID = accountID
    request.input = updateInput

    _ = try await clientManager.client.identity.updateAccount(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  private func createAccountUpdateInput() -> Identity_AccountUpdateRequestInput {
    var updateInput = Identity_AccountUpdateRequestInput()
    updateInput.name = accountName
    updateInput.contactPhone = contactPhone
    updateInput.addressLine1 = addressLine1
    updateInput.addressLine2 = addressLine2
    updateInput.city = city
    updateInput.state = state
    updateInput.zipCode = zipCode
    updateInput.country = country
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

  private func createInvitationInput() -> Identity_AccountInvitationCreationRequestInput {
    var invitationInput = Identity_AccountInvitationCreationRequestInput()
    invitationInput.toEmail = invitationEmail
    invitationInput.toName = invitationName
    invitationInput.note = invitationNote
    return invitationInput
  }

  private func executeInvitationCreation() async throws {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    let invitationInput = createInvitationInput()
    var request = Identity_CreateAccountInvitationRequest()
    request.input = invitationInput

    _ = try await clientManager.client.identity.createAccountInvitation(
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
      var request = Identity_CancelAccountInvitationRequest()
      request.accountInvitationID = invitationID
      request.input = Identity_AccountInvitationUpdateRequestInput()

      _ = try await clientManager.client.identity.cancelAccountInvitation(
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
        userID: membership.belongsToUser.id, newRole: newRole, reason: reason)
      await loadData()
    } errorMessage: {
      "Failed to update member role: \($0.localizedDescription)"
    }
  }

  private func validateMemberRoleUpdate(
    membershipID: String, reason: String
  ) -> Identity_AccountUserMembershipWithUser? {
    guard isAccountAdmin else {
      errorMessage = "Only household admins can change member roles"
      return nil
    }

    guard !reason.isEmpty else {
      errorMessage = "A reason is required for changing member roles"
      return nil
    }

    guard let membership = account?.members.first(where: { $0.id == membershipID }) else {
      errorMessage = "Member not found"
      return nil
    }

    guard membership.hasBelongsToUser, !membership.belongsToUser.id.isEmpty else {
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

  private func createMemberRoleUpdateInput(newRole: String, reason: String)
    -> Identity_ModifyUserPermissionsInput
  {
    var updateInput = Identity_ModifyUserPermissionsInput()
    updateInput.newRole = newRole
    updateInput.reason = reason
    return updateInput
  }

  private func executeMemberRoleUpdate(userID: String, newRole: String, reason: String) async throws
  {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    let updateInput = createMemberRoleUpdateInput(newRole: newRole, reason: reason)

    var request = Identity_UpdateAccountMemberPermissionsRequest()
    request.userID = userID
    request.input = updateInput

    _ = try await clientManager.client.identity.updateAccountMemberPermissions(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  private func initializeFormFields(from account: Identity_Account) {
    accountName = account.name
    contactPhone = account.contactPhone
    addressLine1 = account.addressLine1
    addressLine2 = account.addressLine2
    city = account.city
    state = account.state
    zipCode = account.zipCode
    country = account.country.isEmpty ? "USA" : account.country
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

  var accountDataHasChanged: Bool {
    guard let account = account else { return false }
    return account.name != accountName
      || account.contactPhone != contactPhone
      || account.addressLine1 != addressLine1
      || account.addressLine2 != addressLine2
      || account.city != city
      || account.state != state
      || account.zipCode != zipCode
      || account.country != country
  }
}
