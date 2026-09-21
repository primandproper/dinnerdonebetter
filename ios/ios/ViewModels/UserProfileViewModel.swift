//
//  UserProfileViewModel.swift
//  ios
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import SwiftProtobuf
import SwiftUI

@Observable
@MainActor
class UserProfileViewModel {
  var user: Primandproper_Platform_Identity_V1_User?
  var isLoading = false
  var errorMessage: String?
  var didSucceed = false

  var username: String = ""
  var firstName: String = ""
  var lastName: String = ""

  var hasTwoFactor: Bool {
    user?.hasTwoFactorSecretVerifiedAt == true
  }

  var usernameHasChanged: Bool {
    guard let user = user else { return false }
    return user.username != username
  }

  var detailsHasChanged: Bool {
    guard let user = user else { return false }
    return user.firstName != firstName || user.lastName != lastName
  }

  private let authManager: AuthenticationManager

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  func loadUser() async {
    isLoading = true
    errorMessage = nil

    do {
      let response = try await fetchUser()
      if response.hasResult {
        self.user = response.result
        initializeFormFields(from: response.result)
      }
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      errorMessage = ErrorDisplayFormatter.format(error, context: "load user").message
    }

    isLoading = false
  }

  // Changing the handle is on the auth service rather than in the directory's profile
  // update, and it asks for the password and a second factor: a handle is what somebody
  // signs in with, so moving it is a credential change.
  func updateUsername(currentPassword: String, totpToken: String = "") async -> Bool {
    guard usernameHasChanged, !username.isEmpty else { return false }
    guard !currentPassword.isEmpty else {
      errorMessage = "Password is required"
      return false
    }

    return await performUpdate {
      let (clientManager, metadata) = try await getClientManagerAndMetadata()
      var request = Auth_UpdateUserUsernameRequest()
      request.newUsername = username
      request.currentPassword = currentPassword
      request.totpToken = totpToken

      _ = try await clientManager.client.auth.updateUserUsername(
        request,
        metadata: metadata,
        options: clientManager.defaultCallOptions
      )
      await loadUser()
    }
  }

  // A name is not a credential, so nothing is asked for. Each field is absent unless it is
  // being set, which is what keeps sending a first name from blanking a last one.
  //
  // The birthday is gone. platform's user carries no such column, nothing on this backend
  // ever read one, and storing it means a table of this application's own — which is its
  // own piece of work rather than part of this port.
  func updateUserDetails() async -> Bool {
    guard detailsHasChanged else { return false }

    return await performUpdate {
      let (clientManager, metadata) = try await getClientManagerAndMetadata()
      var input = Primandproper_Platform_Identity_V1_ProfileUpdateInput()
      input.firstName = firstName
      input.lastName = lastName

      var request = Primandproper_Platform_Identity_V1_UpdateProfileRequest()
      request.input = input

      _ = try await clientManager.client.identity.updateProfile(
        request,
        metadata: metadata,
        options: clientManager.defaultCallOptions
      )
      await loadUser()
    }
  }

  private func fetchUser() async throws -> Auth_GetSelfResponse {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()
    return try await clientManager.client.auth.getSelf(
      Auth_GetSelfRequest(),
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )
  }

  private func initializeFormFields(from user: Primandproper_Platform_Identity_V1_User) {
    username = user.username
    firstName = user.firstName
    lastName = user.lastName
  }

  private func performUpdate(operation: () async throws -> Void) async -> Bool {
    isLoading = true
    errorMessage = nil

    do {
      try await operation()
      didSucceed = true
      isLoading = false
      return true
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      errorMessage = ErrorDisplayFormatter.format(error, context: "update").message
      isLoading = false
      return false
    }
  }

  private func getClientManagerAndMetadata() async throws -> (
    ClientManager<HTTP2ClientTransport.TransportServices>, GRPCCore.Metadata
  ) {
    guard let clientManager = try? authManager.getClientManager() else {
      throw NSError(
        domain: "UserProfileViewModel", code: 1,
        userInfo: [NSLocalizedDescriptionKey: "Failed to get client manager"])
    }

    guard let oauth2Token = await authManager.getOAuth2AccessToken() else {
      throw NSError(
        domain: "UserProfileViewModel", code: 2,
        userInfo: [NSLocalizedDescriptionKey: "Failed to get OAuth2 access token"])
    }

    let metadata = clientManager.authenticatedMetadata(accessToken: oauth2Token)
    return (clientManager, metadata)
  }
}
