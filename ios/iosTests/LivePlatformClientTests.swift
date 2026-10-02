//
//  LivePlatformClientTests.swift
//  iosTests
//
//  Drives the app's own sign-in and the screens ported onto platform-client-swift against a
//  running backend: the localdev server on :8001. They are skipped unless LIVE_BACKEND is set,
//  which xcodebuild passes through as TEST_RUNNER_LIVE_BACKEND=1.
//

import CryptoKit
import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2TransportServices
import PlatformClient
import Testing

@testable import ios

private let liveBackend = ProcessInfo.processInfo.environment["LIVE_BACKEND"] != nil

@Suite("Live backend", .serialized, .enabled(if: liveBackend))
@MainActor
struct LivePlatformClientTests {
  init() {
    APIConfiguration.currentEnvironment = .local
  }

  /// A fresh person, registered through the app's own sign-up, and signed in.
  private func signedInNewcomer() async throws -> (AuthenticationManager, String) {
    let suffix = UUID().uuidString.prefix(8).lowercased()
    let username = "ios_\(suffix)"
    let password = "Password-\(UUID().uuidString)"
    let manager = AuthenticationManager()
    await manager.logout()

    let registration = await manager.register(
      input: RegistrationInput(
        emailAddress: "\(username)@example.email",
        username: username,
        password: password,
        accountName: "\(username)'s household",
        firstName: "Test",
        lastName: "Person",
        invitationToken: "",
        invitationID: ""
      ))
    #expect(registration.success, "registration: \(registration.error ?? "")")

    let login = await manager.login(username: username, password: password, totpToken: nil)
    #expect(login.success, "login: \(login.error ?? "")")
    return (manager, username)
  }

  @Test("register, sign in, and every ported screen loads")
  func signedInScreens() async throws {
    let (manager, username) = try await signedInNewcomer()
    #expect(manager.isAuthenticated)
    #expect(manager.username == username)
    #expect(!manager.userID.isEmpty)
    #expect(!manager.accountID.isEmpty)
    #expect(try await manager.getClientManager().session.held() != nil)

    let me = try await CurrentUserService.shared.currentUser(using: manager, forceRefresh: true)
    #expect(me.username == username)

    let profile = UserProfileViewModel(authManager: manager)
    await profile.loadUser()
    #expect(profile.errorMessage == nil, "profile: \(profile.errorMessage ?? "")")
    #expect(profile.user?.id == manager.userID)

    let account = AccountSettingsViewModel(authManager: manager)
    await account.loadData()
    #expect(account.errorMessage == nil, "account: \(account.errorMessage ?? "")")
    #expect(account.account?.id == manager.accountID)
    #expect(account.currentUserID == manager.userID)

    let settings = ServiceSettingsViewModel(
      authManager: manager, userSettingsService: UserSettingsService())
    await settings.loadData()
    #expect(settings.errorMessage == nil, "settings: \(settings.errorMessage ?? "")")
    if let first = settings.configurableSettings.first(where: { !$0.setting.enumeration.isEmpty }) {
      let next = first.setting.enumeration.last ?? first.currentValue
      let saved = await settings.saveSetting(definition: first.setting, value: next)
      #expect(saved, "saving \(first.setting.name): \(settings.errorMessage ?? "")")
    }

    let sessions = SessionsViewModel(authManager: manager)
    await sessions.loadSessions()
    #expect(sessions.errorMessage == nil, "sessions: \(sessions.errorMessage ?? "")")
    #expect(sessions.sessions.contains { $0.current })

    await manager.logout()
    #expect(!manager.isAuthenticated)
    #expect(try await manager.getClientManager().session.held() == nil)
  }

  @Test("an access token inside the skew is refreshed before the call, and sign-out ends it")
  func refreshThenSignOut() async throws {
    let (manager, _) = try await signedInNewcomer()
    let session = try manager.getClientManager().session
    let issued = try #require(try await session.held())

    // A second Session over the same Keychain item, holding a token that claims to be about
    // to expire: its first call has to exchange the refresh token first.
    let store = KeychainCredentialStore(
      service: "\(Branding.keychainPrefix).session", account: AppEnvironment.local.rawValue)
    var aging = issued
    aging.expiresAt = .init(date: Date().addingTimeInterval(5))
    try await store.save(aging)
    let other = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: APIConfiguration.grpcHost, port: APIConfiguration.grpcPort, store: store)
    let client = other.client
    let me = try await other.session.call { metadata in
      try await client.signIn.getSelf(.init(), metadata: metadata)
    }
    #expect(me.user.id == manager.userID)
    let refreshed = try #require(try await other.session.held())
    #expect(refreshed.token != issued.token)
    #expect(refreshed.refreshToken != issued.refreshToken)

    // Ending that login on the server means its refresh token is dead everywhere.
    await other.session.signOut()
    var exchange = Primandproper_Platform_Signin_V1_ExchangeRefreshTokenRequest()
    exchange.refreshToken = refreshed.refreshToken
    await #expect(throws: (any Error).self) {
      _ = try await client.signIn.exchangeRefreshToken(exchange)
    }
    await manager.logout()
  }

  @Test("the seeded admin is asked for a second factor, and gets in with it")
  func secondFactor() async throws {
    let manager = AuthenticationManager()
    await manager.logout()

    let first = await manager.login(username: "admin_user", password: "admin_pass", totpToken: nil)
    #expect(!first.success)
    #expect(first.requiresTOTP, "error: \(first.error ?? "")")

    let second = await manager.login(
      username: "admin_user", password: "admin_pass", totpToken: Self.seededAdminTOTP())
    #expect(second.success, "error: \(second.error ?? "")")
    #expect(manager.username == "admin_user")
    await manager.logout()
  }

  /// The localdev admin's secret is base32 of all zero bytes, and HMAC pads a short key with
  /// zeros, so any all-zero key yields the same code.
  private static func seededAdminTOTP(at date: Date = Date()) -> String {
    var counter = UInt64(date.timeIntervalSince1970 / 30).bigEndian
    let message = Data(bytes: &counter, count: 8)
    let mac = Array(
      HMAC<Insecure.SHA1>.authenticationCode(
        for: message, using: SymmetricKey(data: Data(count: 64))))
    let offset = Int(mac[19] & 0x0f)
    let value =
      (UInt32(mac[offset] & 0x7f) << 24) | (UInt32(mac[offset + 1]) << 16)
      | (UInt32(mac[offset + 2]) << 8) | UInt32(mac[offset + 3])
    return String(format: "%06d", value % 1_000_000)
  }
}
