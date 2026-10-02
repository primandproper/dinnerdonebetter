//
//  AuthenticationManaging.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2TransportServices

/// Result of a login attempt
struct LoginResult {
  let success: Bool
  let error: String?
  let requiresTOTP: Bool
}

/// Result of a registration attempt
struct RegistrationResult {
  let success: Bool
  let error: String?
}

/// Input parameters for user registration
struct RegistrationInput {
  let emailAddress: String
  let username: String
  let password: String
  let accountName: String
  let firstName: String
  let lastName: String
  let invitationToken: String
  let invitationID: String
}

/// Protocol defining the authentication interface
/// Allows both AuthenticationManager and MockAuthenticationManager to be used interchangeably
///
/// It holds no tokens. The login lives in the ClientManager's Session, which refreshes it and
/// keeps it in the Keychain; what is here is who the login belongs to, for the UI.
protocol AuthenticationManaging: AnyObject {
  var isAuthenticated: Bool { get set }
  var username: String { get set }
  var userID: String { get set }
  var accountID: String { get set }

  func login(username: String, password: String, totpToken: String?) async -> LoginResult
  func register(input: RegistrationInput) async -> RegistrationResult
  func getClientManager() throws -> ClientManager<HTTP2ClientTransport.TransportServices>
  func invalidateCredentialsIfSessionError(_ error: Error) async
  func logout() async
}
