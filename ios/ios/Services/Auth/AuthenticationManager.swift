import CircuitBreaking
import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import GRPCNIOTransportHTTP2TransportServices
import PlatformClient
import Retry
import RevenueCat
import SwiftProtobuf
import SwiftUI
import UIKit

// NOTE: `import Observability` (broad) cannot appear in this file: Observability exports a
// `struct Observation`, which shadows Apple's `Observation` module that the `@Observable`
// macro expands against. Import only the protocols we reference here.
import protocol Observability.Logger
import struct Observability.ObservabilityError
import protocol Observability.Observer

@Observable
// swiftlint:disable:next type_body_length
class AuthenticationManager: AuthenticationManaging {
  var isAuthenticated: Bool = false
  var username: String = ""
  var userID: String = ""
  var accountID: String = ""
  /// Whether an operator has forced this person to choose a new password. While it holds, the
  /// server refuses every other call with PASSWORD_CHANGE_REQUIRED, so the app shows the
  /// change-password form instead of anything that would make one.
  var passwordChangeRequired: Bool = false

  // Client manager following grpc-swift issue #2211 pattern
  // Reuses a single GRPCClient instance across all service clients, and holds the Session
  private var clientManager: ClientManager<HTTP2ClientTransport.TransportServices>?

  // Tracks which environment the current client was created for
  private var clientEnvironment: AppEnvironment?

  /// The APNs token this launch was handed, kept so that a ClientManager made after it, for
  /// another environment, is handed it too. Where it is registered, and under which login, is
  /// the ClientManager's `Devices`, which persists it.
  private var apnsToken: Data?

  // Mock support for UI tests
  private var mockManager: MockAuthenticationManager?
  private var isUsingMock: Bool {
    Features.useMockAuth
  }

  // Platform cross-cutting services, injected at the composition root. Used by
  // `authenticatedCall` (see AuthenticationManager+AuthenticatedCall.swift).
  let observer: any Observer
  let retry: any Retry.RetryPolicy
  let breaker: any CircuitBreaking.CircuitBreaker
  /// Component logger for lifecycle/session logging outside of an operation scope.
  let logger: any Logger

  init(
    observer: any Observer = PlatformServices.shared.observer("AuthenticationManager"),
    retry: any Retry.RetryPolicy = PlatformServices.shared.retry,
    breaker: any CircuitBreaking.CircuitBreaker = PlatformServices.shared.breaker
  ) {
    self.observer = observer
    self.retry = retry
    self.breaker = breaker
    self.logger = PlatformServices.shared.logger("AuthenticationManager")
    logger
      .withValue("environment", APIConfiguration.currentEnvironment.displayName)
      .withValue("grpc.host", APIConfiguration.grpcHost)
      .withValue("grpc.port", APIConfiguration.grpcPort)
      .info("initializing")

    // Check if we should use mock behavior (for UI tests)
    if isUsingMock {
      let mock = MockAuthenticationManager()
      let launchArguments = ProcessInfo.processInfo.arguments
      if launchArguments.contains("--mock-requires-totp") {
        mock.configure(behavior: .requiresTOTP)
      } else if launchArguments.contains("--mock-success") {
        mock.configure(behavior: .success)
      } else {
        mock.configure(behavior: .requiresTOTP)
      }
      self.mockManager = mock
      print("🎭 AuthenticationManager: Using mock mode")
    } else {
      restoreProfile()
      Task { await self.confirmRestoredSession() }
    }

    NotificationCenter.default.addObserver(
      forName: .environmentDidChange,
      object: nil,
      queue: .main
    ) { [weak self] _ in
      self?.handleEnvironmentChange()
    }
  }

  // MARK: - Restoring a login

  /// The Keychain item the Session keeps its login in. One per environment, so that switching
  /// to another server never presents it a refresh token a different one issued.
  private static func credentialStore(for environment: AppEnvironment) -> KeychainCredentialStore {
    KeychainCredentialStore(
      service: "\(Branding.keychainPrefix).session", account: environment.rawValue)
  }

  /// The Keychain item `Devices` keeps this install's push registration in. One per
  /// environment, like the login: a registration one server made means nothing to another.
  private static func deviceRegistrationStore(
    for environment: AppEnvironment
  ) -> KeychainDeviceRegistrationStore {
    KeychainDeviceRegistrationStore(
      service: "\(Branding.keychainPrefix).device", account: environment.rawValue)
  }

  /// Who the login belongs to, remembered so the app opens signed in without waiting on the
  /// Keychain or the network. None of it is a credential: the Session holds those.
  private enum ProfileKey: String, CaseIterable {
    case username, userID, accountID

    var defaultsKey: String { "\(Branding.keychainPrefix).profile.\(rawValue)" }
  }

  private func persistProfile() {
    let defaults = UserDefaults.standard
    defaults.set(username, forKey: ProfileKey.username.defaultsKey)
    defaults.set(userID, forKey: ProfileKey.userID.defaultsKey)
    defaults.set(accountID, forKey: ProfileKey.accountID.defaultsKey)
  }

  private func clearPersistedProfile() {
    for key in ProfileKey.allCases {
      UserDefaults.standard.removeObject(forKey: key.defaultsKey)
    }
  }

  /// Restore who was signed in. Called once during init, synchronously, so the first frame is
  /// the right one; `confirmRestoredSession` then checks the Session agrees.
  private func restoreProfile() {
    let defaults = UserDefaults.standard
    guard let savedUserID = defaults.string(forKey: ProfileKey.userID.defaultsKey),
      !savedUserID.isEmpty
    else {
      return
    }
    self.userID = savedUserID
    self.username = defaults.string(forKey: ProfileKey.username.defaultsKey) ?? ""
    self.accountID = defaults.string(forKey: ProfileKey.accountID.defaultsKey) ?? ""
    self.isAuthenticated = true
    // Do NOT call logInToRevenueCatIfNeeded here. AuthManager init runs during
    // @State setup, before IOSApp.init(), so any Task here runs before
    // Purchases.configure() and crashes. Sync is triggered from iosApp.onAppear.
  }

  /// Signs the app out locally if the Session holds no login after all: the Keychain item is
  /// gone, or belongs to a server this build no longer talks to.
  @MainActor
  private func confirmRestoredSession() async {
    guard isAuthenticated else { return }
    do {
      let manager = try getClientManager()
      if try await manager.session.held() == nil {
        logger.info("a profile was remembered but the Session holds no login; signing out")
        await endLocally()
        return
      }
      // A forced password change may have been imposed since the last launch.
      if case .authenticated(let signedIn) = try await manager.session.getAuthStatus(
        options: manager.defaultCallOptions)
      {
        passwordChangeRequired = signedIn.requiredActions.contains(.changePassword)
      }
    } catch {
      logger.error("loading the held session", error)
    }
  }

  /// Reads who the Session's login belongs to. Every sign-in calls it, since a token names an
  /// account but no user.
  @MainActor
  private func loadAuthStatus(fallbackUsername: String) async throws {
    let manager = try getClientManager()
    switch try await manager.session.getAuthStatus(options: manager.defaultCallOptions) {
    case .anonymous:
      throw NotSignedInError()
    case .authenticated(let signedIn):
      self.userID = signedIn.status.user.id
      self.username =
        signedIn.status.user.username.isEmpty ? fallbackUsername : signedIn.status.user.username
      self.accountID = signedIn.status.activeAccountID
      self.passwordChangeRequired = signedIn.requiredActions.contains(.changePassword)
      self.isAuthenticated = true
      persistProfile()
    }
  }

  /// Called once the password has been changed. The server lifts a forced change as part of
  /// the change, so the auth status is read again rather than assumed, and the profile with
  /// it; if it cannot be read, the form is let go anyway, and the next refusal would bring it
  /// back.
  @MainActor
  func passwordWasChanged() async {
    let wasRequired = passwordChangeRequired
    do {
      try await loadAuthStatus(fallbackUsername: username)
    } catch {
      logger.error("reading the auth status after a password change", error)
      passwordChangeRequired = false
    }
    // While the change was owed, the server refused the device registration along with every
    // other call, so the token is handed over again now that it is lifted.
    if wasRequired, !passwordChangeRequired, let apnsToken, let manager = try? getClientManager() {
      handOver(apnsToken, to: manager)
    }
  }

  /// Log in to RevenueCat with the current account ID so purchases are tied to the user.
  /// Call from iosApp.onAppear (after Purchases.configure), not from restoreCredentials.
  func logInToRevenueCatIfNeeded() async {
    guard RevenueCatConfiguration.isConfigured else { return }
    guard !accountID.isEmpty else { return }
    do {
      _ = try await Purchases.shared.logIn(accountID)
      print("✅ RevenueCat: Logged in as account \(accountID)")
    } catch {
      print("⚠️ RevenueCat: Failed to log in: \(error)")
    }
  }

  /// Log out from RevenueCat when the user signs out.
  private func logOutFromRevenueCat() {
    guard RevenueCatConfiguration.isConfigured else { return }
    Task {
      do {
        _ = try await Purchases.shared.logOut()
        print("✅ RevenueCat: Logged out")
      } catch {
        print("⚠️ RevenueCat: Failed to log out: \(error)")
      }
    }
  }

  /// Tear down the existing gRPC client so the next request creates one
  /// pointing at the newly selected environment.
  private func handleEnvironmentChange() {
    print(
      "🔄 Environment changed to \(APIConfiguration.currentEnvironment.displayName), resetting connection"
    )
    // Sign out of the old server first: logout ends the login through the client it was
    // made with, and only then is that client dropped.
    Task {
      await logout()
      self.clientManager = nil
      self.clientEnvironment = nil
    }
  }

  /// Get or create the client manager, following the grpc-swift issue #2211 pattern.
  /// This ensures we reuse a single GRPCClient instance, and so a single Session, across all
  /// requests.
  func getClientManager() throws -> ClientManager<HTTP2ClientTransport.TransportServices> {
    // If using mock, delegate to mock manager
    if let mock = mockManager {
      return try mock.getClientManager()
    }

    let env = APIConfiguration.currentEnvironment
    if let existing = clientManager, clientEnvironment == env {
      return existing
    }

    let host = APIConfiguration.grpcHost
    let port = APIConfiguration.grpcPort
    let useTLS = APIConfiguration.grpcUsesTLS
    print(
      "🔧 Creating ClientManager for \(env.displayName): \(host):\(port) (TLS: \(useTLS))"
    )
    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: host, port: port, useTLS: useTLS, store: Self.credentialStore(for: env),
      deviceStore: Self.deviceRegistrationStore(for: env)
    )
    clientManager = manager
    clientEnvironment = env
    if let apnsToken {
      handOver(apnsToken, to: manager)
    }
    return manager
  }

  // MARK: - Push registration

  /// Hands the APNs token iOS issued to the current ClientManager's `Devices`. Call from
  /// `application(_:didRegisterForRemoteNotificationsWithDeviceToken:)`. Before sign-in the
  /// token is held there and registered once the Session is authenticated, so nothing here
  /// waits for a login.
  func registerDeviceToken(_ token: Data) {
    apnsToken = token
    guard let manager = try? getClientManager() else { return }
    handOver(token, to: manager)
  }

  private func handOver(
    _ token: Data, to manager: ClientManager<HTTP2ClientTransport.TransportServices>
  ) {
    let devices = manager.devices
    let logger = self.logger
    Task {
      do {
        _ = try await devices.register(apnsToken: token, platform: .ios)
      } catch is NotSignedInError {
        // Held, and registered when somebody signs in.
      } catch {
        logger.error("registering the device token", error)
      }
    }
  }

  func login(username: String, password: String, totpToken: String? = nil) async -> LoginResult {
    // If using mock, delegate to mock manager
    if let mock = mockManager {
      let result = await mock.login(username: username, password: password, totpToken: totpToken)
      // Sync state from mock to self
      await MainActor.run {
        self.isAuthenticated = mock.isAuthenticated
        self.username = mock.username
        self.userID = mock.userID
        self.accountID = mock.accountID
        if result.success {
          let reporter = AnalyticsConfiguration.provideEventReporter()
          reporter.identify(
            userID: mock.userID,
            properties: [
              "username": mock.username,
              "accountID": mock.accountID,
            ]
          )
          reporter.track(event: "login_succeeded", properties: [:])
        } else if result.requiresTOTP {
          AnalyticsConfiguration.provideEventReporter().track(
            event: "login_2fa_required", properties: [:])
        } else {
          AnalyticsConfiguration.provideEventReporter().track(
            event: "login_failed",
            properties: ["error": result.error ?? "Unknown error"])
        }
      }
      if result.success {
        await logInToRevenueCatIfNeeded()
      }
      return result
    }

    print("🔐 Login attempt for user: \(username)")
    let reporter = AnalyticsConfiguration.provideEventReporter()
    let failed = { (message: String) -> LoginResult in
      reporter.track(event: "login_failed", properties: ["error": message])
      return LoginResult(success: false, error: message, requiresTOTP: false)
    }

    let request = PasswordSignIn(
      handle: .username(username),
      password: password,
      totpCode: totpToken.flatMap { $0.isEmpty ? nil : $0 }
    )

    do {
      try Task.checkCancellation()
      let manager = try getClientManager()

      switch try await manager.session.signIn(request) {
      case .secondFactorRequired:
        reporter.track(event: "login_2fa_required", properties: [:])
        return LoginResult(success: false, error: "Please enter your 2FA code.", requiresTOTP: true)
      case .signedIn:
        print("✅ Login successful")
      }

      try await loadAuthStatus(fallbackUsername: username)
      await MainActor.run {
        reporter.identify(
          userID: self.userID,
          properties: [
            "username": self.username,
            "accountID": self.accountID,
          ]
        )
        reporter.track(event: "login_succeeded", properties: [:])
      }
      await logInToRevenueCatIfNeeded()
      await MainActor.run {
        UIApplication.shared.registerForRemoteNotifications()
      }
      return LoginResult(success: true, error: nil, requiresTOTP: false)
    } catch let error as PlatformError where error.is(SignInReason.passwordChangeRequired) {
      // Like a second factor, this is a refusal the person can act on rather than a failure:
      // the login was made, and the server is holding it at the change-password form until
      // a new password is chosen. The app routes there on `passwordChangeRequired`.
      reporter.track(event: "login_password_change_required", properties: [:])
      // GetAuthStatus is one of the calls the server still answers, so who the login belongs
      // to is read and remembered here, as on any sign-in. If it can't be, the form still
      // shows, and `passwordWasChanged` reads it again.
      do {
        try await loadAuthStatus(fallbackUsername: username)
      } catch {
        logger.error("reading the auth status of a login held for a password change", error)
      }
      await MainActor.run {
        self.passwordChangeRequired = true
        self.isAuthenticated = true
        if self.username.isEmpty { self.username = username }
      }
      return LoginResult(success: true, error: nil, requiresTOTP: false)
    } catch let error as PlatformError {
      print("❌ Sign-in refused: \(error)")
      return failed(Self.signInRefusalMessage(error))
    } catch is CancellationError {
      return failed("Login was cancelled")
    } catch {
      print("❌ Error details: \(String(describing: error))")
      return failed("Login failed: \(error.localizedDescription)")
    }
  }

  func register(input: RegistrationInput) async -> RegistrationResult {
    // If using mock, delegate to mock manager
    if let mock = mockManager {
      return await mock.register(input: input)
    }

    print("📝 Registration attempt for user: \(input.username)")

    // The birthday is gone with the column behind it: platform's user has none, nothing on
    // the backend ever read one, and storing it means a table of this application's own.
    var request = Primandproper_Platform_Signin_V1_RegisterRequest()
    request.user.emailAddress = input.emailAddress.trimmingCharacters(in: .whitespaces)
    request.user.username = input.username.trimmingCharacters(in: .whitespaces)
    request.user.firstName = input.firstName.trimmingCharacters(in: .whitespaces)
    request.user.lastName = input.lastName.trimmingCharacters(in: .whitespaces)
    request.account.name = input.accountName.trimmingCharacters(in: .whitespaces)
    request.password = input.password

    let invitationToken = input.invitationToken.trimmingCharacters(in: .whitespaces)
    let invitationID = input.invitationID.trimmingCharacters(in: .whitespaces)
    if !invitationToken.isEmpty || !invitationID.isEmpty {
      request.invitation.token = invitationToken
      request.invitation.invitationID = invitationID
    }

    // Both are required: the registration policy refuses a sign-up without them. The form
    // gates submission on the person accepting them.
    request.agreements = [.termsOfService, .privacyPolicy]

    do {
      try Task.checkCancellation()
      let manager = try getClientManager()
      let response = try await TokenCaller().callAnonymous {
        try await manager.client.signIn.register(request, options: manager.defaultCallOptions)
      }
      guard response.hasRegistration else {
        return RegistrationResult(
          success: false, error: "No creation result received from server")
      }
      print("✅ Registration successful")
      return RegistrationResult(success: true, error: nil)
    } catch let error as PlatformError {
      print("❌ Registration refused: \(error)")
      if isTransient(error) {
        return RegistrationResult(success: false, error: Self.transientRefusalMessage)
      }
      switch error.code {
      case .alreadyExists:
        return RegistrationResult(
          success: false, error: "Username or email address already exists.")
      case .invalidArgument:
        return RegistrationResult(
          success: false, error: "Invalid registration data. Please check your input.")
      default:
        return RegistrationResult(
          success: false, error: "Registration failed: \(error.serverMessage)")
      }
    } catch is CancellationError {
      print("❌ Registration cancelled")
      return RegistrationResult(success: false, error: "Registration was cancelled")
    } catch {
      print("❌ Error details: \(String(describing: error))")
      return RegistrationResult(
        success: false, error: "Registration failed: \(error.localizedDescription)")
    }
  }

  /// What a person is told when the server could not answer right now. Which code said so
  /// (UNAVAILABLE, DEADLINE_EXCEEDED, RESOURCE_EXHAUSTED) is platform-client's `isTransient`,
  /// so this never grows its own list.
  static let transientRefusalMessage =
    "We couldn't reach the server. Please check your connection and try again."

  /// What a person is told when a password sign-in was refused. A second factor never reaches
  /// here: `Session.signIn` answers it as `.secondFactorRequired`.
  static func signInRefusalMessage(_ error: PlatformError) -> String {
    if isTransient(error) {
      return transientRefusalMessage
    }
    if error.is(SignInReason.invalidCredentials) || error.code == .unauthenticated {
      return "Invalid username or password."
    }
    return "Login failed: \(error.serverMessage)"
  }

  /// Signs the app out locally when a call failed because the login is over. The Session has
  /// already decided that by the time a call throws: it refreshes and retries an
  /// UNAUTHENTICATED once, and ends the login itself if the retry is refused too. What is left
  /// here is the UI's half.
  ///
  /// It also notices a call refused with PASSWORD_CHANGE_REQUIRED, which is not the login
  /// ending but the server holding it at the form: the app sends the person there.
  func invalidateCredentialsIfSessionError(_ error: Error) async {
    if error.platformError?.is(SignInReason.passwordChangeRequired) == true {
      await MainActor.run { self.passwordChangeRequired = true }
      return
    }
    let underlying = (error as? ObservabilityError)?.underlying ?? error
    let ended = underlying is NotSignedInError || error.platformError?.code == .unauthenticated
    guard ended, isAuthenticated else { return }
    print("🔐 The login has ended: \(error)")
    await endLocally()
  }

  /// Ends the login on the server, then here: the device's push registration is revoked first,
  /// while the login can still make that call, then the Session signs out. Neither throws: a
  /// sign-out that could not be delivered still clears the login on this device.
  func logout() async {
    if let manager = try? getClientManager() {
      await manager.signOut()
    }
    await endLocally()
  }

  /// The app's half of signing out: everything that knew who was signed in forgets it.
  @MainActor
  private func endLocally() async {
    if let manager = clientManager {
      try? await manager.session.clear()
    }
    await CurrentUserService.shared.clear()
    logOutFromRevenueCat()
    AnalyticsConfiguration.provideEventReporter().reset()
    self.isAuthenticated = false
    self.username = ""
    self.userID = ""
    self.accountID = ""
    self.passwordChangeRequired = false
    clearPersistedProfile()
  }
}
