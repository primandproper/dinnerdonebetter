import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import PlatformClient

/// One login the signed-in user holds: this device's, or another's.
typealias SignIn = Primandproper_Platform_Signin_V1_ActiveSignIn

@Observable
@MainActor
class SessionsViewModel {
  var sessions: [SignIn] = []
  var isLoading = false
  var errorMessage: String?

  private let authManager: AuthenticationManager

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  func loadSessions() async {
    isLoading = true
    errorMessage = nil

    do {
      let response = try await authManager.authenticatedCall("listSignIns", idempotent: true) {
        client, metadata, options in
        try await client.signIn.listSignIns(.init(), metadata: metadata, options: options)
      }
      sessions = response.signIns
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      errorMessage = "Failed to load sessions"
      print("❌ Error loading sessions: \(error)")
    }

    isLoading = false
  }

  /// Ends another login. A login is a refresh-token family, so this ends it on whichever
  /// device holds it at its next refresh.
  func revokeSession(familyID: String) async -> Bool {
    isLoading = true
    errorMessage = nil

    do {
      var request = Primandproper_Platform_Signin_V1_EndSignInRequest()
      request.familyID = familyID
      _ = try await authManager.authenticatedCall("endSignIn") { client, metadata, options in
        try await client.signIn.endSignIn(request, metadata: metadata, options: options)
      }
      sessions.removeAll { $0.familyID == familyID }
      isLoading = false
      return true
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      errorMessage = "Failed to revoke session"
      print("❌ Error revoking session: \(error)")
      isLoading = false
      return false
    }
  }

  func revokeAllOtherSessions() async -> Bool {
    isLoading = true
    errorMessage = nil

    do {
      _ = try await authManager.authenticatedCall("endOtherSignIns") { client, metadata, options in
        try await client.signIn.endOtherSignIns(.init(), metadata: metadata, options: options)
      }
      sessions.removeAll { !$0.current }
      isLoading = false
      return true
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      errorMessage = "Failed to revoke sessions"
      print("❌ Error revoking sessions: \(error)")
      isLoading = false
      return false
    }
  }
}
