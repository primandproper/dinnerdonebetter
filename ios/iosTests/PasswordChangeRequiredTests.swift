//
//  PasswordChangeRequiredTests.swift
//  iosTests
//
//  An operator can force a password change; the server then refuses every other call with
//  PASSWORD_CHANGE_REQUIRED. Like SECOND_FACTOR_REQUIRED, that is a refusal the person acts on,
//  so the app sends them to the change-password form rather than showing generic failures.
//

import Foundation
import GRPCCore
import Observability
import PlatformClient
import Testing

@testable import ios

private let passwordChangeRefusal = PlatformError(
  code: .failedPrecondition,
  serverMessage: "a password change is required",
  reason: .signIn(.passwordChangeRequired)
)

@Suite(.serialized)
struct PasswordChangeRequiredTests {
  @Test("a call refused with PASSWORD_CHANGE_REQUIRED routes to the form and keeps the login")
  @MainActor
  func refusalRoutesToTheForm() async {
    let manager = AuthenticationManager()
    manager.isAuthenticated = true
    manager.userID = "user-1"

    await manager.invalidateCredentialsIfSessionError(passwordChangeRefusal)

    #expect(manager.passwordChangeRequired)
    #expect(manager.isAuthenticated)
    #expect(manager.userID == "user-1")
  }

  @Test("the reason is read off a refusal authenticatedCall wrapped, too")
  @MainActor
  func wrappedRefusalRoutesToTheForm() async {
    let manager = AuthenticationManager()
    manager.isAuthenticated = true

    await manager.invalidateCredentialsIfSessionError(
      ObservabilityError("gRPC updateAccount", passwordChangeRefusal))

    #expect(manager.passwordChangeRequired)
    #expect(manager.isAuthenticated)
  }

  @Test("the same code without the reason is not a forced password change")
  @MainActor
  func failedPreconditionAloneIsNot() async {
    let manager = AuthenticationManager()
    manager.isAuthenticated = true

    await manager.invalidateCredentialsIfSessionError(
      PlatformError(code: .failedPrecondition, serverMessage: "something else"))

    #expect(!manager.passwordChangeRequired)
    #expect(manager.isAuthenticated)
  }

  @Test("signing out lets go of the form")
  @MainActor
  func logoutClearsTheRequirement() async {
    let manager = AuthenticationManager()
    manager.isAuthenticated = true
    manager.passwordChangeRequired = true

    await manager.logout()

    #expect(!manager.passwordChangeRequired)
    #expect(!manager.isAuthenticated)
  }
}

struct ChangePasswordViewModelTests {
  @MainActor
  private func filledIn(
    current: String = "old-\(UUID().uuidString)",
    new: String = "new-\(UUID().uuidString)",
    confirm: String? = nil
  ) -> ChangePasswordViewModel {
    let viewModel = ChangePasswordViewModel(authManager: AuthenticationManager())
    viewModel.currentPassword = current
    viewModel.newPassword = new
    viewModel.confirmPassword = confirm ?? new
    return viewModel
  }

  @Test("a filled-in form can be submitted and has nothing wrong with it")
  @MainActor
  func validForm() {
    let viewModel = filledIn()
    #expect(viewModel.canSubmit)
    #expect(viewModel.validationError == nil)
  }

  @Test("mismatched new passwords are refused before anything is sent")
  @MainActor
  func mismatch() async {
    let viewModel = filledIn(confirm: "different-\(UUID().uuidString)")

    let changed = await viewModel.submit()

    #expect(!changed)
    #expect(viewModel.errorMessage == "The new passwords don't match.")
  }

  @Test("the same password again is refused before anything is sent")
  @MainActor
  func unchanged() async {
    let password = "same-\(UUID().uuidString)"
    let viewModel = filledIn(current: password, new: password)

    let changed = await viewModel.submit()

    #expect(!changed)
    #expect(viewModel.errorMessage == "Choose a password different from your current one.")
  }

  @Test("an empty field keeps the button disabled")
  @MainActor
  func emptyField() {
    let viewModel = filledIn()
    viewModel.currentPassword = ""
    #expect(!viewModel.canSubmit)
  }

  @Test("refusals are told apart by reason")
  @MainActor
  func refusalMessages() {
    #expect(
      ChangePasswordViewModel.message(
        for: PlatformError(
          code: .permissionDenied, serverMessage: "", reason: .signIn(.invalidCredentials)))
        == "Your current password or code is incorrect.")
    #expect(
      ChangePasswordViewModel.message(
        for: PlatformError(
          code: .permissionDenied, serverMessage: "", reason: .signIn(.secondFactorRequired)))
        == "Enter the code from your authenticator app.")
    #expect(
      ChangePasswordViewModel.message(
        for: PlatformError(
          code: .invalidArgument, serverMessage: "too short", reason: .signIn(.passwordRefused)))
        == "That password can't be used: too short")
  }
}
