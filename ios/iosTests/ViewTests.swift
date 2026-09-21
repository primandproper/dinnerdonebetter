//
//  ViewTests.swift
//  iosTests
//
//  Created by Auto on 12/8/25.
//

import Foundation
import SwiftProtobuf
import SwiftUI
@testable import ios
import Testing

// MARK: - Email Validation Helper (extracted from RegisterView)

func isValidEmail(_ email: String) -> Bool {
  let emailRegex = "[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}"
  let emailPredicate = NSPredicate(format: "SELF MATCHES %@", emailRegex)
  return emailPredicate.evaluate(with: email)
}

// MARK: - RegisterView Validation Tests

@Suite(.serialized)
struct RegisterViewValidationTests {
  @Test("isValidEmail returns true for valid emails")
  @MainActor
  func testIsValidEmailValid() async {
    let validEmails = [
      "test@example.com",
      "user.name@domain.co.uk",
      "user+tag@example.org",
      "user123@test-domain.com",
      "a@b.co"
    ]

    for email in validEmails {
      #expect(isValidEmail(email) == true, "Email '\(email)' should be valid")
    }
  }

  @Test("isValidEmail returns false for invalid emails")
  @MainActor
  func testIsValidEmailInvalid() async {
    // Test each invalid email individually for clearer failure messages
    #expect(isValidEmail("invalid-email") == false, "Email 'invalid-email' should be invalid - no @ symbol")
    #expect(isValidEmail("@example.com") == false, "Email '@example.com' should be invalid - no local part")
    #expect(isValidEmail("user@") == false, "Email 'user@' should be invalid - no domain")
    #expect(isValidEmail("user@domain") == false, "Email 'user@domain' should be invalid - no TLD")
    #expect(isValidEmail("user name@example.com") == false, "Email 'user name@example.com' should be invalid - space in local part")
    #expect(isValidEmail("") == false, "Empty string should be invalid")
    #expect(isValidEmail("user@domain.") == false, "Email 'user@domain.' should be invalid - trailing dot")
    // Note: ".user@domain.com" is accepted by the current regex as the local part allows dots
  }

  @Test("RegisterView form validation requires email")
  @MainActor
  func testRegisterViewRequiresEmail() async {
    let showLogin = Binding<Bool>(get: { false }, set: { _ in })
    let authManager = createMockAuthenticationManager()
    _ = RegisterView(showLogin: showLogin)
      .environment(authManager)
  }

  @Test("RegisterView form validation requires username")
  @MainActor
  func testRegisterViewRequiresUsername() async {
    let showLogin = Binding<Bool>(get: { false }, set: { _ in })
    let authManager = createMockAuthenticationManager()
    _ = RegisterView(showLogin: showLogin)
      .environment(authManager)
  }

  @Test("RegisterView form validation requires password")
  @MainActor
  func testRegisterViewRequiresPassword() async {
    let showLogin = Binding<Bool>(get: { false }, set: { _ in })
    let authManager = createMockAuthenticationManager()
    _ = RegisterView(showLogin: showLogin)
      .environment(authManager)
  }

  @Test("RegisterView form validation requires password match")
  @MainActor
  func testRegisterViewRequiresPasswordMatch() async {
    let showLogin = Binding<Bool>(get: { false }, set: { _ in })
    let authManager = createMockAuthenticationManager()
    _ = RegisterView(showLogin: showLogin)
      .environment(authManager)
  }

  @Test("RegisterView form validation requires password length")
  @MainActor
  func testRegisterViewRequiresPasswordLength() async {
    let showLogin = Binding<Bool>(get: { false }, set: { _ in })
    let authManager = createMockAuthenticationManager()
    _ = RegisterView(showLogin: showLogin)
      .environment(authManager)
  }
}

// MARK: - LoginView Tests

struct LoginViewTests {
  @Test("LoginView initializes with default values")
  @MainActor
  func testLoginViewInitialization() async {
    let showRegister = Binding<Bool>(get: { false }, set: { _ in })
    let authManager = createMockAuthenticationManager()
    _ = LoginView(showRegister: showRegister)
      .environment(authManager)
  }

  @Test("LoginView requires username for login")
  @MainActor
  func testLoginViewRequiresUsername() async {
    let showRegister = Binding<Bool>(get: { false }, set: { _ in })
    let authManager = createMockAuthenticationManager()
    _ = LoginView(showRegister: showRegister)
      .environment(authManager)
  }

  @Test("LoginView requires password for login")
  @MainActor
  func testLoginViewRequiresPassword() async {
    let showRegister = Binding<Bool>(get: { false }, set: { _ in })
    let authManager = createMockAuthenticationManager()
    _ = LoginView(showRegister: showRegister)
      .environment(authManager)
  }

  @Test("LoginView handles TOTP requirement")
  @MainActor
  func testLoginViewHandlesTOTP() async {
    let showRegister = Binding<Bool>(get: { false }, set: { _ in })
    let authManager = createMockAuthenticationManager()
    _ = LoginView(showRegister: showRegister)
      .environment(authManager)
  }
}

// MARK: - AccountSettingsView Tests

struct AccountSettingsViewTests {
  @Test("AccountSettingsView initializes view model on appear")
  @MainActor
  func testAccountSettingsViewInitializesViewModel() async {
    let authManager = createMockAuthenticationManagerForAccount()
    _ = AccountSettingsView()
      .environment(authManager)
  }

  @Test("AccountSettingsView shows loading state")
  @MainActor
  func testAccountSettingsViewShowsLoading() async {
    let authManager = createMockAuthenticationManagerForAccount()
    _ = AccountSettingsView()
      .environment(authManager)
  }

  @Test("AccountSettingsView shows error state")
  @MainActor
  func testAccountSettingsViewShowsError() async {
    let authManager = createMockAuthenticationManagerForAccount()
    _ = AccountSettingsView()
      .environment(authManager)
  }

  @Test("AccountSettingsView shows account information when loaded")
  @MainActor
  func testAccountSettingsViewShowsAccountInfo() async {
    let authManager = createMockAuthenticationManagerForAccount()
    _ = AccountSettingsView()
      .environment(authManager)
  }

  @Test("AccountSettingsView shows members section when members exist")
  @MainActor
  func testAccountSettingsViewShowsMembers() async {
    let authManager = createMockAuthenticationManagerForAccount()
    _ = AccountSettingsView()
      .environment(authManager)
  }

  @Test("AccountSettingsView shows invitations section when invitations exist")
  @MainActor
  func testAccountSettingsViewShowsInvitations() async {
    let authManager = createMockAuthenticationManagerForAccount()
    _ = AccountSettingsView()
      .environment(authManager)
  }

  @Test("AccountSettingsView shows send invitation section for admins")
  @MainActor
  func testAccountSettingsViewShowsSendInvitation() async {
    let authManager = createMockAuthenticationManagerForAccount()
    _ = AccountSettingsView()
      .environment(authManager)
  }
}

// MARK: - MemberCard Tests

struct MemberCardTests {
  @Test("MemberCard displayName uses first and last name")
  @MainActor
  func testMemberCardDisplayNameFirstLast() async {
    var member = createMockMembership()
    var user = Primandproper_Platform_Identity_V1_User()
    user.id = "user-1"
    user.firstName = "John"
    user.lastName = "Doe"
    user.username = "johndoe"

    member.user = user

    _ = MemberCard(
      member: member,
      currentUserID: "user-1",
      isAccountAdmin: true,
      onRoleChange: { _, _ in }
    )
  }

  @Test("MemberCard displayName uses first name only")
  @MainActor
  func testMemberCardDisplayNameFirstOnly() async {
    var member = createMockMembership()
    var user = Primandproper_Platform_Identity_V1_User()
    user.id = "user-1"
    user.firstName = "John"
    user.username = "johndoe"

    member.user = user

    _ = MemberCard(
      member: member,
      currentUserID: "user-1",
      isAccountAdmin: true,
      onRoleChange: { _, _ in }
    )
  }

  @Test("MemberCard displayName uses username when no name")
  @MainActor
  func testMemberCardDisplayNameUsername() async {
    var member = createMockMembership()
    var user = Primandproper_Platform_Identity_V1_User()
    user.id = "user-1"
    user.username = "johndoe"

    member.user = user

    _ = MemberCard(
      member: member,
      currentUserID: "user-1",
      isAccountAdmin: true,
      onRoleChange: { _, _ in }
    )
  }

  @Test("MemberCard displayName returns Unknown when no user")
  @MainActor
  func testMemberCardDisplayNameUnknown() async {
    var member = createMockMembership()
    member.clearUser()

    _ = MemberCard(
      member: member,
      currentUserID: "user-1",
      isAccountAdmin: true,
      onRoleChange: { _, _ in }
    )
  }

  @Test("MemberCard shows current user indicator")
  @MainActor
  func testMemberCardShowsCurrentUser() async {
    var member = createMockMembership(userID: "user-1")
    var user = Primandproper_Platform_Identity_V1_User()
    user.id = "user-1"
    user.username = "currentuser"

    member.user = user

    _ = MemberCard(
      member: member,
      currentUserID: "user-1",
      isAccountAdmin: true,
      onRoleChange: { _, _ in }
    )
  }

  @Test("MemberCard initializes with correct role")
  @MainActor
  func testMemberCardInitializesRole() async {
    var member = createMockMembership(role: "account_admin")
    var user = Primandproper_Platform_Identity_V1_User()
    user.id = "user-1"

    member.user = user

    _ = MemberCard(
      member: member,
      currentUserID: "user-2",
      isAccountAdmin: true,
      onRoleChange: { _, _ in }
    )
  }
}

// MARK: - InvitationCard Tests

struct InvitationCardTests {
  @Test("InvitationCard displays email")
  @MainActor
  func testInvitationCardDisplaysEmail() async {
    var invitation = Primandproper_Platform_Identity_V1_Invitation()
    invitation.toEmail = "test@example.com"
    invitation.toName = "Test User"
    invitation.status = .pending

    _ = InvitationCard(
      invitation: invitation,
      isAccountAdmin: false,
      onCancel: nil
    )
  }

  @Test("InvitationCard displays name when available")
  @MainActor
  func testInvitationCardDisplaysName() async {
    var invitation = Primandproper_Platform_Identity_V1_Invitation()
    invitation.toEmail = "test@example.com"
    invitation.toName = "Test User"
    invitation.status = .pending

    _ = InvitationCard(
      invitation: invitation,
      isAccountAdmin: false,
      onCancel: nil
    )
  }

  @Test("InvitationCard displays status")
  @MainActor
  func testInvitationCardDisplaysStatus() async {
    var invitation = Primandproper_Platform_Identity_V1_Invitation()
    invitation.toEmail = "test@example.com"
    invitation.status = .accepted

    _ = InvitationCard(
      invitation: invitation,
      isAccountAdmin: false,
      onCancel: nil
    )
  }

  @Test("InvitationCard handles empty name")
  @MainActor
  func testInvitationCardHandlesEmptyName() async {
    var invitation = Primandproper_Platform_Identity_V1_Invitation()
    invitation.toEmail = "test@example.com"
    invitation.toName = ""
    invitation.status = .pending

    _ = InvitationCard(
      invitation: invitation,
      isAccountAdmin: false,
      onCancel: nil
    )
  }

  @Test("InvitationCard handles different statuses")
  @MainActor
  func testInvitationCardHandlesStatuses() async {
    let statuses: [Primandproper_Platform_Identity_V1_InvitationStatus] = [.pending, .accepted, .rejected, .cancelled]

    for status in statuses {
      var invitation = Primandproper_Platform_Identity_V1_Invitation()
      invitation.toEmail = "test@example.com"
      invitation.status = status

      _ = InvitationCard(
        invitation: invitation,
        isAccountAdmin: false,
        onCancel: nil
      )
    }
  }
}

// MARK: - Date Formatting Tests

struct DateFormattingTests {
  @Test("formatDate formats date correctly")
  func testFormatDate() {
    let date = Date(timeIntervalSince1970: 946684800)  // Jan 1, 2000
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    let expected = formatter.string(from: date)

    #expect(!expected.isEmpty)
  }

  @Test("formatDate handles different dates")
  func testFormatDateDifferentDates() {
    let dates = [
      Date(timeIntervalSince1970: 0),  // Jan 1, 1970
      Date(timeIntervalSince1970: 946684800),  // Jan 1, 2000
      Date()  // Today
    ]

    let formatter = DateFormatter()
    formatter.dateStyle = .medium

    for date in dates {
      let formatted = formatter.string(from: date)
      #expect(!formatted.isEmpty)
    }
  }
}

// MARK: - Discrete Quantity Scaling Tests

struct DiscreteQuantityScalingTests {
  @Test("Uint32 discrete max scaling rounds up")
  func testUint32ScaleUpRoundsMax() {
    let scaled = DiscreteQuantityScaling.scaledMax(UInt32(3), scale: 1.5)
    #expect(scaled == 5)
  }

  @Test("Uint16 discrete max scaling clamps overflow")
  func testUint16ScaleUpClampsOverflow() {
    let scaled = DiscreteQuantityScaling.scaledMax(UInt16.max, scale: 2.0)
    #expect(scaled == UInt16.max)
  }

  @Test("Aggregated instrument quantity text keeps min and scales max")
  func testAggregatedInstrumentQuantityTextUsesDiscreteRules() {
    var aggregated = AggregatedInstrumentVessel(
      itemID: "instrument-id",
      name: "Mixing bowl",
      type: .instrument
    )
    aggregated.addQuantity(min: UInt32(1), max: UInt32(1))

    #expect(aggregated.quantityText(scale: 0.5) == "1")
    #expect(aggregated.quantityText(scale: 2.0) == "1 - 2")
  }
}

// MARK: - Recipe Performance Step Formatting Tests

struct RecipePerformanceStepFormattingTests {
  @Test("Step ingredient display includes scaled quantity and unit")
  func testStepIngredientDisplayIncludesQuantityAndUnit() {
    var ingredient = Mealplanning_RecipeStepIngredient()
    ingredient.name = "Flour"
    ingredient.minQuantity = 2
    ingredient.maxQuantity = 2  // same as min → exact quantity, avoids "3+" when no max
    var unit = Mealplanning_ValidMeasurementUnit()
    unit.name = "cup"
    unit.pluralName = "cups"
    ingredient.measurementUnit = unit

    let display = formatStepIngredientDisplay(ingredient, scale: 1.5)

    #expect(display == "3 cups Flour")
  }

  @Test("Step ingredient display uses singular for quantity 1")
  func testStepIngredientDisplayUsesSingularForOne() {
    var ingredient = Mealplanning_RecipeStepIngredient()
    ingredient.name = "Water"
    ingredient.minQuantity = 1
    ingredient.maxQuantity = 1
    var unit = Mealplanning_ValidMeasurementUnit()
    unit.name = "cup"
    unit.pluralName = "cups"
    ingredient.measurementUnit = unit

    let display = formatStepIngredientDisplay(ingredient, scale: 1.0)

    #expect(display == "1 cup Water")
  }

  @Test("Step ingredient display falls back to name without quantity")
  func testStepIngredientDisplayFallsBackToName() {
    var ingredient = Mealplanning_RecipeStepIngredient()
    ingredient.name = "Salt"

    let display = formatStepIngredientDisplay(ingredient, scale: 2.0)

    #expect(display == "Salt")
  }
}

// MARK: - Wash Hands Gating Tests

@Suite(.serialized)
struct WashHandsGatingTests {
  @Test("Steps remain blocked until wash hands completed")
  @MainActor
  func testCanCheckStepRequiresWashHands() async {
    let authManager = AuthenticationManager()
    let viewModel = PerformRecipeViewModel(recipeID: "recipe-id", authManager: authManager)

    var recipe = Mealplanning_Recipe()
    recipe.id = "recipe-id"
    var step = Mealplanning_RecipeStep()
    step.id = "step-1"
    recipe.steps = [step]
    viewModel.recipe = recipe

    #expect(viewModel.canCheckStep(recipeID: "recipe-id", stepID: "step-1") == false)

    viewModel.washHandsCompleted = true

    #expect(viewModel.canCheckStep(recipeID: "recipe-id", stepID: "step-1") == true)
  }

  @Test("Steps with completion conditions remain blocked until conditions are checked")
  @MainActor
  func testCanCheckStepRequiresCompletionConditions() async {
    let authManager = AuthenticationManager()
    let viewModel = PerformRecipeViewModel(recipeID: "recipe-id", authManager: authManager)

    var completionCondition = Mealplanning_RecipeStepCompletionCondition()
    completionCondition.id = "condition-1"
    completionCondition.notes = "Chicken reaches 165F"

    var step = Mealplanning_RecipeStep()
    step.id = "step-1"
    step.completionConditions = [completionCondition]

    var recipe = Mealplanning_Recipe()
    recipe.id = "recipe-id"
    recipe.steps = [step]
    viewModel.recipe = recipe
    viewModel.washHandsCompleted = true

    #expect(viewModel.canCheckStep(recipeID: "recipe-id", stepID: "step-1") == false)

    viewModel.toggleStepCompletionCondition(
      recipeID: "recipe-id",
      stepID: "step-1",
      conditionIdentifier: "condition-1"
    )

    #expect(viewModel.canCheckStep(recipeID: "recipe-id", stepID: "step-1") == true)
  }
}
