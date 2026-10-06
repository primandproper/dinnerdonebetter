//
//  ChangePasswordViewModel.swift
//  ios
//
//  Changing the signed-in person's password: from account settings by choice, or because an
//  operator forced it and the server is refusing everything else with PASSWORD_CHANGE_REQUIRED.
//

import Foundation
import PlatformClient
import SwiftUI

@Observable
@MainActor
final class ChangePasswordViewModel {
  var currentPassword: String = ""
  var newPassword: String = ""
  var confirmPassword: String = ""
  /// Sent whenever filled. The server requires it from someone with a proven second factor
  /// and ignores it for everybody else.
  var totpCode: String = ""
  /// Shown once the server has asked for a code, so most people never see the field.
  var requiresTOTP = false

  var isSaving = false
  var errorMessage: String?
  var didChange = false

  private let authManager: AuthenticationManager

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  var canSubmit: Bool {
    !isSaving && !currentPassword.isEmpty && !newPassword.isEmpty && !confirmPassword.isEmpty
      && (!requiresTOTP || !totpCode.isEmpty)
  }

  /// What is wrong with the form before anything is sent, or nil. How long or unusual a
  /// password must be is the server's policy, not repeated here.
  var validationError: String? {
    if newPassword != confirmPassword {
      return "The new passwords don't match."
    }
    if newPassword == currentPassword {
      return "Choose a password different from your current one."
    }
    return nil
  }

  func submit() async -> Bool {
    if let problem = validationError {
      errorMessage = problem
      return false
    }

    isSaving = true
    errorMessage = nil
    defer { isSaving = false }

    var request = Primandproper_Platform_Signin_V1_UpdatePasswordRequest()
    request.currentPassword = currentPassword
    request.newPassword = newPassword
    request.totpCode = totpCode

    do {
      _ = try await authManager.authenticatedCall("updatePassword") {
        [request] client, metadata, options in
        try await client.signIn.updatePassword(request, metadata: metadata, options: options)
      }
      await authManager.passwordWasChanged()
      currentPassword = ""
      newPassword = ""
      confirmPassword = ""
      totpCode = ""
      requiresTOTP = false
      didChange = true
      return true
    } catch {
      if error.platformError?.is(SignInReason.secondFactorRequired) == true {
        requiresTOTP = true
      }
      errorMessage = Self.message(for: error)
      return false
    }
  }

  /// What a person is told when the change was refused, by reason rather than by code.
  static func message(for error: any Error) -> String {
    guard let refusal = error.platformError else {
      return "Couldn't change your password: \(error.localizedDescription)"
    }
    if isTransient(refusal) {
      return AuthenticationManager.transientRefusalMessage
    }
    if refusal.is(SignInReason.secondFactorRequired) {
      return "Enter the code from your authenticator app."
    }
    if refusal.is(SignInReason.invalidCredentials) {
      return "Your current password or code is incorrect."
    }
    if refusal.is(SignInReason.passwordRefused) {
      return "That password can't be used: \(refusal.serverMessage)"
    }
    return "Couldn't change your password: \(refusal.serverMessage)"
  }
}
