//
//  DeviceTokenRegistrationService.swift
//  ios
//
//  Handles APNs device token registration with the backend and archiving on logout.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2TransportServices
import PlatformClient
import SwiftProtobuf

/// Service that registers APNs device tokens with the backend and archives them on logout.
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
actor DeviceTokenRegistrationService {
  static let shared = DeviceTokenRegistrationService()

  private var _storedDeviceTokenHex: String?
  private var _userDeviceTokenID: String?
  private weak var _authManager: (any AuthenticationManaging)?

  private init() {}

  /// Configure with the authentication manager. Call from a view that has access to it (e.g. ContentView.onAppear).
  nonisolated func configure(authManager: any AuthenticationManaging) {
    Task { await configureImpl(authManager: authManager) }
  }

  private func configureImpl(authManager: any AuthenticationManaging) async {
    _authManager = authManager
    if let token = _storedDeviceTokenHex {
      await tryReportToken(hexToken: token)
    }
  }

  /// Try to report a stored token. Call after successful login in case the token arrived before auth.
  nonisolated func tryReportStoredToken() {
    Task { await tryReportStoredTokenImpl() }
  }

  private func tryReportStoredTokenImpl() async {
    guard let token = _storedDeviceTokenHex else { return }
    await tryReportToken(hexToken: token, authManager: _authManager)
  }

  /// Report the APNs device token to the backend. Call from AppDelegate.didRegisterForRemoteNotificationsWithDeviceToken.
  nonisolated func reportDeviceToken(_ deviceToken: Data) {
    let hexToken = deviceToken.map { String(format: "%02.2hhx", $0) }.joined()
    Task { await reportDeviceTokenImpl(hexToken: hexToken) }
  }

  private func reportDeviceTokenImpl(hexToken: String) async {
    _storedDeviceTokenHex = hexToken
    await tryReportToken(hexToken: hexToken, authManager: _authManager)
  }

  /// Archive the current device token on the backend. Call from AuthenticationManager.logout before clearing credentials.
  func archiveCurrentDeviceToken(authManager: any AuthenticationManaging) async {
    guard let tokenID = _userDeviceTokenID, !tokenID.isEmpty else { return }

    do {
      let manager = try authManager.getClientManager()

      var request = Primandproper_Platform_Notifications_V1_RevokeDeviceRequest()
      request.deviceID = tokenID

      let client = manager.client
      let options = manager.defaultCallOptions
      _ = try await manager.session.call { metadata in
        try await client.notifications.revokeDevice(request, metadata: metadata, options: options)
      }

      _userDeviceTokenID = nil

      print("✅ Archived device token on logout")
    } catch {
      print("⚠️ Failed to archive device token on logout: \(error)")
    }
  }

  private func tryReportToken(hexToken: String, authManager: (any AuthenticationManaging)? = nil)
    async
  {
    let manager = authManager ?? _authManager

    guard let auth = manager, auth.isAuthenticated else {
      return
    }

    do {
      let clientManager = try auth.getClientManager()

      var request = Primandproper_Platform_Notifications_V1_RegisterDeviceRequest()
      request.input.token = hexToken
      request.input.platform = .ios

      let client = clientManager.client
      let options = clientManager.defaultCallOptions
      let response = try await clientManager.session.call { metadata in
        try await client.notifications.registerDevice(request, metadata: metadata, options: options)
      }

      _userDeviceTokenID = response.result.id

      print("✅ Registered device token with backend")
    } catch {
      print("⚠️ Failed to register device token: \(error)")
    }
  }
}
