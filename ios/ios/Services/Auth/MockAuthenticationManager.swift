//
//  MockAuthenticationManager.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2TransportServices
import SwiftUI

/// Mock AuthenticationManager for testing purposes
/// Can simulate various login scenarios including TOTP requirements
/// This is a standalone class that mimics AuthenticationManager's interface
@Observable
class MockAuthenticationManager: AuthenticationManaging {
  // Configuration for mock behavior
  enum MockBehavior {
    case success
    case requiresTOTP
    case requiresTOTPThenSuccess
    case invalidCredentials
    case serverError
  }

  // Public properties matching AuthenticationManager
  var isAuthenticated: Bool = false
  var username: String = ""
  var userID: String = ""
  var accountID: String = ""

  // Mock configuration
  private var mockBehavior: MockBehavior = .requiresTOTP
  private var loginAttemptCount: Int = 0
  private var totpCodeProvided: Bool = false

  init() {
    print("🎭 MockAuthenticationManager: Initialized")
  }

  /// Configure the mock behavior
  func configure(behavior: MockBehavior) {
    self.mockBehavior = behavior
    self.loginAttemptCount = 0
    self.totpCodeProvided = false
  }

  func login(username: String, password: String, totpToken: String? = nil) async -> LoginResult {
    print("🎭 MockAuthenticationManager: Login attempt for user: \(username)")

    loginAttemptCount += 1

    // Simulate network delay
    try? await Task.sleep(nanoseconds: 500_000_000)  // 0.5 seconds

    switch mockBehavior {
    case .success:
      // Simulate successful login
      await MainActor.run {
        self.isAuthenticated = true
        self.username = username
        self.userID = "mock-user-123"
        self.accountID = "mock-account-456"
      }
      return LoginResult(success: true, error: nil, requiresTOTP: false)

    case .requiresTOTP:
      // First attempt: require TOTP
      // Second attempt with TOTP: success
      if let totpToken = totpToken, !totpToken.isEmpty {
        // TOTP provided, simulate success
        totpCodeProvided = true
        await MainActor.run {
          self.isAuthenticated = true
          self.username = username
          self.userID = "mock-user-123"
          self.accountID = "mock-account-456"
        }
        return LoginResult(success: true, error: nil, requiresTOTP: false)
      } else {
        // No TOTP provided, require it
        return LoginResult(success: false, error: "Please enter your 2FA code.", requiresTOTP: true)
      }

    case .requiresTOTPThenSuccess:
      // Similar to requiresTOTP but explicitly tracks the flow
      if loginAttemptCount == 1 {
        // First attempt without TOTP
        return LoginResult(success: false, error: "Please enter your 2FA code.", requiresTOTP: true)
      } else {
        // Second attempt with TOTP
        if let totpToken = totpToken, !totpToken.isEmpty {
          totpCodeProvided = true
          await MainActor.run {
            self.isAuthenticated = true
            self.username = username
            self.userID = "mock-user-123"
            self.accountID = "mock-account-456"
          }
          return LoginResult(success: true, error: nil, requiresTOTP: false)
        } else {
          return LoginResult(success: false, error: "TOTP code is required.", requiresTOTP: true)
        }
      }

    case .invalidCredentials:
      return LoginResult(
        success: false, error: "Invalid username or password.", requiresTOTP: false)

    case .serverError:
      return LoginResult(
        success: false, error: "Server is unavailable. Please try again later.", requiresTOTP: false
      )
    }
  }

  func getClientManager() throws -> ClientManager<HTTP2ClientTransport.TransportServices> {
    // For mock, we throw an error since UI tests focused on login shouldn't need the real client
    // If you need getClientManager in UI tests, consider creating a full mock client manager
    throw NSError(
      domain: "MockAuthenticationManager",
      code: 1,
      userInfo: [
        NSLocalizedDescriptionKey: "MockAuthenticationManager does not support getClientManager. "
          + "This is expected for UI tests that only test authentication flow. "
          + "If you need client functionality, extend MockAuthenticationManager to provide a mock client."
      ]
    )
  }

  func register(input: RegistrationInput) async -> RegistrationResult {
    print("🎭 MockAuthenticationManager: Registration attempt for user: \(input.username)")

    // Simulate network delay
    try? await Task.sleep(nanoseconds: 500_000_000)  // 0.5 seconds

    // Mock successful registration
    return RegistrationResult(success: true, error: nil)
  }

  func invalidateCredentialsIfSessionError(_ error: Error) async {
    // No-op for mock; UI tests don't need credential invalidation
  }

  func logout() async {
    self.isAuthenticated = false
    self.username = ""
    self.userID = ""
    self.accountID = ""
  }
}
