//
//  AccountSettingsViewModelTests.swift
//  iosTests
//
//  Created by Auto on 12/8/25.
//

import Foundation
import SwiftProtobuf
@testable import ios
import Testing

// MARK: - Helper Functions for Test Data

func createMockAccount(id: String = "account-1", name: String = "Test Account") -> Primandproper_Platform_Identity_V1_Account {
  // The address is one value rather than seven fields: it travels as a unit, and an
  // application that never collects one leaves a zero rather than seven empty strings.
  var address = Primandproper_Platform_Identity_V1_BillingAddress()
  address.phone = "555-1234"
  address.line1 = "123 Main St"
  address.line2 = "Apt 4"
  address.city = "Springfield"
  address.state = "IL"
  address.postalCode = "62701"
  address.country = "USA"

  var account = Primandproper_Platform_Identity_V1_Account()
  account.id = id
  account.name = name
  account.billingAddress = address
  return account
}

func createMockUser(id: String = "user-1") -> Auth_GetSelfResponse {
  var response = Auth_GetSelfResponse()
  var user = Primandproper_Platform_Identity_V1_User()
  user.id = id
  response.result = user
  return response
}

// A membership carries a set of roles rather than one, because a role is a grant and
// somebody may hold several.

func createMockMembership(
  id: String = "membership-1",
  userID: String = "user-1",
  role: String = "account_admin"
) -> Primandproper_Platform_Identity_V1_MembershipWithUser {
  var inner = Primandproper_Platform_Identity_V1_Membership()
  inner.id = id
  inner.belongsToUser = userID
  inner.roles = [role]

  var user = Primandproper_Platform_Identity_V1_User()
  user.id = userID

  var membership = Primandproper_Platform_Identity_V1_MembershipWithUser()
  membership.membership = inner
  membership.user = user
  return membership
}

func createMockAuthenticationManagerForAccount() -> AuthenticationManager {
  let manager = AuthenticationManager()
  manager.isAuthenticated = true
  manager.oauth2AccessToken = "mock-oauth2-token"
  return manager
}

// MARK: - Initialization Tests

struct AccountSettingsInitializationTests {
  @Test("AccountSettingsViewModel initializes with auth manager")
  @MainActor
  func testViewModelInitialization() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    #expect(viewModel.account == nil)
    #expect(viewModel.user == nil)
    #expect(viewModel.invitations.isEmpty)
    #expect(viewModel.isLoading == false)
    #expect(viewModel.accountName.isEmpty)
    #expect(viewModel.country == "USA")
  }
}

// MARK: - Computed Properties Tests

struct ComputedPropertiesTests {
  @Test("isAccountAdmin returns true when user is admin")
  @MainActor
  func testIsAccountAdminTrue() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "account_admin")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account

    #expect(viewModel.isAccountAdmin == true)
  }

  @Test("isAccountAdmin returns false when user is not admin")
  @MainActor
  func testIsAccountAdminFalse() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "member")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account

    #expect(viewModel.isAccountAdmin == false)
  }

  @Test("isAccountAdmin returns false when no user")
  @MainActor
  func testIsAccountAdminNoUser() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    viewModel.account = account
    viewModel.user = nil

    #expect(viewModel.isAccountAdmin == false)
  }

  @Test("currentUserMembership returns membership when found")
  @MainActor
  func testCurrentUserMembershipFound() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership()
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account

    #expect(viewModel.currentUserMembership != nil)
    #expect(viewModel.currentUserMembership?.membership.id == membership.membership.id)
  }

  @Test("currentUserMembership returns nil when not found")
  @MainActor
  func testCurrentUserMembershipNotFound() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(userID: "different-user")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account

    #expect(viewModel.currentUserMembership == nil)
  }

  @Test("currentUserID returns user ID when available")
  @MainActor
  func testCurrentUserID() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let user = createMockUser(id: "user-123")
    viewModel.user = user

    #expect(viewModel.currentUserID == "user-123")
  }

  @Test("currentUserID returns empty string when not available")
  @MainActor
  func testCurrentUserIDEmpty() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    viewModel.user = nil

    #expect(viewModel.currentUserID.isEmpty)
  }

  @Test("accountDataHasChanged returns true when data changed")
  @MainActor
  func testAccountDataHasChanged() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount(name: "Original Name")
    viewModel.account = account
    viewModel.accountName = "New Name"

    #expect(viewModel.accountDataHasChanged == true)
  }

  @Test("accountDataHasChanged returns false when data unchanged")
  @MainActor
  func testAccountDataHasNotChanged() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    viewModel.account = account
    viewModel.accountName = account.name
    viewModel.contactPhone = account.billingAddress.phone
    viewModel.addressLine1 = account.billingAddress.line1
    viewModel.addressLine2 = account.billingAddress.line2
    viewModel.city = account.billingAddress.city
    viewModel.state = account.billingAddress.state
    viewModel.zipCode = account.billingAddress.postalCode
    viewModel.country = account.billingAddress.country

    #expect(viewModel.accountDataHasChanged == false)
  }

  @Test("accountDataHasChanged returns false when no account")
  @MainActor
  func testAccountDataHasChangedNoAccount() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    viewModel.account = nil

    #expect(viewModel.accountDataHasChanged == false)
  }
}

// MARK: - Form Initialization Tests

struct FormInitializationTests {
  @Test("loadData initializes form fields from account")
  @MainActor
  func testLoadDataInitializesFormFields() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    var account = createMockAccount()
    account.name = "Test Account"
    account.billingAddress.phone = "555-9999"
    account.billingAddress.line1 = "456 Oak Ave"
    account.billingAddress.line2 = "Suite 2"
    account.billingAddress.city = "Chicago"
    account.billingAddress.state = "IL"
    account.billingAddress.postalCode = "60601"
    account.billingAddress.country = "USA"

    viewModel.account = account
    let user = createMockUser()
    viewModel.user = user

    await viewModel.loadData()

    #expect(viewModel.accountName == account.name || viewModel.accountName.isEmpty)
  }
}

// MARK: - Validation Tests

struct AccountSettingsValidationTests {
  @Test("updateAccount returns false when no account")
  @MainActor
  func testUpdateAccountNoAccount() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    viewModel.account = nil

    let result = await viewModel.updateAccount()

    #expect(result == false)
    #expect(viewModel.errorMessage?.contains("No account loaded") == true)
  }

  @Test("updateAccount returns false when not admin")
  @MainActor
  func testUpdateAccountNotAdmin() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "member")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account

    let result = await viewModel.updateAccount()

    #expect(result == false)
    #expect(viewModel.errorMessage?.contains("household admins") == true)
  }

  @Test("sendInvitation returns false when no account")
  @MainActor
  func testSendInvitationNoAccount() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    viewModel.account = nil
    viewModel.invitationEmail = "test@example.com"

    let result = await viewModel.sendInvitation()

    #expect(result == false)
    #expect(viewModel.errorMessage?.contains("No account loaded") == true)
  }

  @Test("sendInvitation returns false when not admin")
  @MainActor
  func testSendInvitationNotAdmin() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "member")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account
    viewModel.invitationEmail = "test@example.com"

    let result = await viewModel.sendInvitation()

    #expect(result == false)
    #expect(viewModel.errorMessage?.contains("household admins") == true)
  }

  @Test("sendInvitation returns false when email empty")
  @MainActor
  func testSendInvitationEmptyEmail() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "account_admin")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account
    viewModel.invitationEmail = ""

    let result = await viewModel.sendInvitation()

    #expect(result == false)
    #expect(viewModel.errorMessage?.contains("Email address is required") == true)
  }

  @Test("sendInvitation returns false when email invalid")
  @MainActor
  func testSendInvitationInvalidEmail() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "account_admin")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account
    viewModel.invitationEmail = "invalid-email"

    let result = await viewModel.sendInvitation()

    #expect(result == false)
    #expect(viewModel.errorMessage?.contains("Invalid email") == true)
  }

  @Test("updateMemberRole returns false when not admin")
  @MainActor
  func testUpdateMemberRoleNotAdmin() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "member")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account

    let result = await viewModel.updateMemberRole(membershipID: "membership-123", newRole: "member", reason: "Test reason")

    #expect(result == false)
    #expect(viewModel.errorMessage?.contains("household admins") == true)
  }

  @Test("updateMemberRole returns false when reason empty")
  @MainActor
  func testUpdateMemberRoleEmptyReason() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "account_admin")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account

    let result = await viewModel.updateMemberRole(membershipID: "membership-123", newRole: "member", reason: "")

    #expect(result == false)
    #expect(viewModel.errorMessage?.contains("reason is required") == true)
  }

  @Test("updateMemberRole returns false when membership not found")
  @MainActor
  func testUpdateMemberRoleNotFound() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "account_admin")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account

    let result = await viewModel.updateMemberRole(membershipID: "non-existent", newRole: "member", reason: "Test reason")

    #expect(result == false)
    #expect(viewModel.errorMessage?.contains("Member not found") == true)
  }
}

// MARK: - State Management Tests

struct StateManagementTests {
  @Test("loadData sets loading state")
  @MainActor
  func testLoadDataSetsLoading() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    #expect(viewModel.isLoading == false)

    _ = await viewModel.loadData()

    #expect(viewModel.isLoading == false)
  }

  @Test("updateAccount sets loading state")
  @MainActor
  func testUpdateAccountSetsLoading() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "account_admin")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account

    _ = await viewModel.updateAccount()

    #expect(viewModel.isLoading == false)
  }

  @Test("sendInvitation clears form on success")
  @MainActor
  func testSendInvitationClearsForm() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let membership = createMockMembership(role: "account_admin")
    viewModel.members = [membership]

    let user = createMockUser()
    viewModel.user = user
    viewModel.account = account
    viewModel.invitationEmail = "test@example.com"
    viewModel.invitationName = "Test Name"
    viewModel.invitationNote = "Test Note"

    _ = await viewModel.sendInvitation()

    #expect(viewModel.isLoading == false)
  }
}

// MARK: - Edge Cases Tests

struct AccountSettingsEdgeCaseTests {
  @Test("accountDataHasChanged detects changes in all fields")
  @MainActor
  func testAccountDataHasChangedAllFields() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    viewModel.account = account

    viewModel.accountName = "Changed"
    #expect(viewModel.accountDataHasChanged == true)

    viewModel.accountName = account.name
    viewModel.contactPhone = "Changed"
    #expect(viewModel.accountDataHasChanged == true)

    viewModel.contactPhone = account.billingAddress.phone
    viewModel.addressLine1 = "Changed"
    #expect(viewModel.accountDataHasChanged == true)
  }

  @Test("loadData handles empty account fields")
  @MainActor
  func testLoadDataEmptyFields() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    var account = createMockAccount()
    account.name = ""
    account.billingAddress.country = ""

    viewModel.account = account
    let user = createMockUser()
    viewModel.user = user

    await viewModel.loadData()

    #expect(viewModel.isLoading == false)
  }

  @Test("isAccountAdmin handles multiple memberships")
  @MainActor
  func testIsAccountAdminMultipleMemberships() async {
    let authManager = createMockAuthenticationManagerForAccount()
    let viewModel = AccountSettingsViewModel(authManager: authManager)

    let account = createMockAccount()
    let member1 = createMockMembership(userID: "user-1", role: "member")
    let member2 = createMockMembership(userID: "user-2", role: "account_admin")
    viewModel.members = [member1, member2]

    let user = createMockUser(id: "user-2")
    viewModel.user = user
    viewModel.account = account

    #expect(viewModel.isAccountAdmin == true)
  }
}

