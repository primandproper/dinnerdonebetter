import CircuitBreaking
import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import GRPCNIOTransportHTTP2TransportServices
import Retry
import RevenueCat
import SwiftProtobuf
import SwiftUI
import UIKit

// NOTE: `import Observability` (broad) cannot appear in this file: Observability exports a
// `struct Observation`, which shadows Apple's `Observation` module that the `@Observable`
// macro expands against. Import only the protocols we reference here.
import protocol Observability.Logger
import protocol Observability.Observer
import protocol Secrets.SecretSource

/// URLSessionDelegate that prevents automatic redirect following
private class NoRedirectDelegate: NSObject, URLSessionTaskDelegate {
  func urlSession(
    _ session: URLSession, task: URLSessionTask,
    willPerformHTTPRedirection response: HTTPURLResponse, newRequest request: URLRequest,
    completionHandler: @escaping (URLRequest?) -> Void
  ) {
    // Return nil to prevent following redirects
    completionHandler(nil)
  }
}

@Observable
// swiftlint:disable:next type_body_length
class AuthenticationManager: AuthenticationManaging {
  var isAuthenticated: Bool = false
  var username: String = ""
  var accessToken: String = ""  // JWT token from login
  var refreshToken: String = ""  // JWT refresh token
  var oauth2AccessToken: String = ""  // OAuth2 access token for API calls
  var oauth2RefreshToken: String = ""  // OAuth2 refresh token
  var oauth2TokenExpiresAt: Date?  // OAuth2 token expiration
  var userID: String = ""
  var accountID: String = ""

  // Client manager following grpc-swift issue #2211 pattern
  // Reuses a single GRPCClient instance across all service clients
  private var clientManager: ClientManager<HTTP2ClientTransport.TransportServices>?

  // Tracks which environment the current client was created for
  private var clientEnvironment: AppEnvironment?

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
    }

    restoreCredentials()

    NotificationCenter.default.addObserver(
      forName: .environmentDidChange,
      object: nil,
      queue: .main
    ) { [weak self] _ in
      self?.handleEnvironmentChange()
    }
  }

  // MARK: - Keychain Persistence

  /// Save current credential state to the Keychain so it survives app restarts / redeploys.
  private func persistCredentials() {
    KeychainManager.save(accessToken, for: .accessToken)
    KeychainManager.save(refreshToken, for: .refreshToken)
    KeychainManager.save(oauth2AccessToken, for: .oauth2AccessToken)
    KeychainManager.save(oauth2RefreshToken, for: .oauth2RefreshToken)
    KeychainManager.save(username, for: .username)
    KeychainManager.save(userID, for: .userID)
    KeychainManager.save(accountID, for: .accountID)

    if let expiresAt = oauth2TokenExpiresAt {
      KeychainManager.save(
        String(expiresAt.timeIntervalSince1970), for: .oauth2TokenExpiresAt)
    }
  }

  /// Restore credentials from the Keychain. Called once during init.
  private func restoreCredentials() {
    guard let savedToken = KeychainManager.loadString(for: .oauth2AccessToken),
      !savedToken.isEmpty
    else {
      print("🔧 No saved credentials found in Keychain")
      return
    }

    self.accessToken = KeychainManager.loadString(for: .accessToken) ?? ""
    self.refreshToken = KeychainManager.loadString(for: .refreshToken) ?? ""
    self.oauth2AccessToken = savedToken
    self.oauth2RefreshToken = KeychainManager.loadString(for: .oauth2RefreshToken) ?? ""
    self.username = KeychainManager.loadString(for: .username) ?? ""
    self.userID = KeychainManager.loadString(for: .userID) ?? ""
    self.accountID = KeychainManager.loadString(for: .accountID) ?? ""

    if let expiresString = KeychainManager.loadString(for: .oauth2TokenExpiresAt),
      let interval = Double(expiresString)
    {
      self.oauth2TokenExpiresAt = Date(timeIntervalSince1970: interval)
    }

    self.isAuthenticated = true
    print("🔧 Restored credentials from Keychain for user: \(self.username)")
    // Do NOT call logInToRevenueCatIfNeeded here. AuthManager init runs during
    // @State setup, before IOSApp.init(), so any Task here runs before
    // Purchases.configure() and crashes. Sync is triggered from iosApp.onAppear.
  }

  /// Clear all saved credentials from the Keychain.
  private func clearPersistedCredentials() {
    KeychainManager.deleteAll()
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
    clientManager = nil
    clientEnvironment = nil
    Task { await logout() }
  }

  /// Get or create the client manager, following the grpc-swift issue #2211 pattern.
  /// This ensures we reuse a single GRPCClient instance across all requests.
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
      host: host, port: port, useTLS: useTLS
    )
    clientManager = manager
    clientEnvironment = env
    return manager
  }

  // swiftlint:disable:next cyclomatic_complexity
  func login(username: String, password: String, totpToken: String? = nil) async -> LoginResult {
    // If using mock, delegate to mock manager
    if let mock = mockManager {
      let result = await mock.login(username: username, password: password, totpToken: totpToken)
      // Sync state from mock to self
      await MainActor.run {
        self.isAuthenticated = mock.isAuthenticated
        self.username = mock.username
        self.accessToken = mock.accessToken
        self.refreshToken = mock.refreshToken
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

    // Create the login request message
    var loginInput = Auth_UserLoginInput()
    loginInput.username = username
    loginInput.password = password
    if let totpToken = totpToken, !totpToken.isEmpty {
      loginInput.totpToken = totpToken
    }

    var requestMessage = Auth_LoginForTokenRequest()
    requestMessage.input = loginInput

    print("📤 Creating gRPC client and sending login request...")

    do {
      // Check for cancellation before starting
      try Task.checkCancellation()

      // Get or create the client manager (follows grpc-swift issue #2211 pattern)
      // This reuses a single GRPCClient instance across all requests
      let manager = try getClientManager()

      // Use the auth service client from the unified Client
      // The ClientManager automatically manages connection lifecycle and provides default call options
      // Default timeout (5 seconds) is set at the ClientManager level
      let response = try await manager.client.auth.loginForToken(
        requestMessage,
        options: manager.defaultCallOptions
      )

      // Extract the token response
      if response.hasResult {
        let tokenResponse = response.result

        print("✅ Login successful, storing tokens")

        // Store JWT authentication data
        await MainActor.run {
          self.isAuthenticated = true
          self.username = username
          self.accessToken = tokenResponse.accessToken
          self.refreshToken = tokenResponse.refreshToken
          self.userID = tokenResponse.userID
          self.accountID = tokenResponse.accountID
        }

        // Exchange JWT token for OAuth2 token
        print("🔄 Exchanging JWT token for OAuth2 token...")
        let oauth2Result = await exchangeJWTForOAuth2Token(jwtToken: tokenResponse.accessToken)
        if !oauth2Result.success {
          print("⚠️ Failed to get OAuth2 token: \(oauth2Result.error ?? "Unknown error")")
          let err = "Failed to complete authentication: \(oauth2Result.error ?? "Unknown error")"
          AnalyticsConfiguration.provideEventReporter().track(
            event: "login_failed", properties: ["error": err])
          await MainActor.run {
            self.isAuthenticated = false
            self.accessToken = ""
            self.refreshToken = ""
            self.userID = ""
            self.accountID = ""
          }
          return LoginResult(success: false, error: err, requiresTOTP: false)
        } else {
          print("✅ OAuth2 token obtained successfully")
          await MainActor.run {
            self.persistCredentials()
            let reporter = AnalyticsConfiguration.provideEventReporter()
            reporter.identify(
              userID: tokenResponse.userID,
              properties: [
                "username": username,
                "accountID": tokenResponse.accountID,
              ]
            )
            reporter.track(event: "login_succeeded", properties: [:])
            DeviceTokenRegistrationService.shared.tryReportStoredToken()
          }
          await logInToRevenueCatIfNeeded()
          await MainActor.run {
            UIApplication.shared.registerForRemoteNotifications()
          }
        }

        return LoginResult(success: true, error: nil, requiresTOTP: false)
      } else {
        print("⚠️ Response received but no token result")
        AnalyticsConfiguration.provideEventReporter().track(
          event: "login_failed",
          properties: ["error": "No token received from server"])
        return LoginResult(
          success: false, error: "No token received from server", requiresTOTP: false)
      }
    } catch let error as GRPCCore.RPCError {
      print("❌ RPC error code: \(error.code)")
      print("❌ RPC error message: \(error.message)")

      // Check if TOTP is required
      let requiresTOTP = error.message.contains("TOTP code required")

      // Provide user-friendly error messages
      let reporter = AnalyticsConfiguration.provideEventReporter()
      switch error.code {
      case .deadlineExceeded:
        reporter.track(
          event: "login_failed",
          properties: ["error": "Request timed out. Please check your connection."])
        return LoginResult(
          success: false, error: "Request timed out. Please check your connection.",
          requiresTOTP: false)
      case .unavailable:
        reporter.track(
          event: "login_failed",
          properties: ["error": "Server is unavailable. Please try again later."])
        return LoginResult(
          success: false, error: "Server is unavailable. Please try again later.",
          requiresTOTP: false)
      case .unauthenticated:
        if requiresTOTP {
          reporter.track(event: "login_2fa_required", properties: [:])
          return LoginResult(
            success: false, error: "Please enter your 2FA code.", requiresTOTP: true)
        }
        reporter.track(
          event: "login_failed",
          properties: ["error": "Invalid username or password."])
        return LoginResult(
          success: false, error: "Invalid username or password.", requiresTOTP: false)
      default:
        if requiresTOTP {
          reporter.track(event: "login_2fa_required", properties: [:])
          return LoginResult(
            success: false, error: "Please enter your 2FA code.", requiresTOTP: true)
        }
        reporter.track(
          event: "login_failed",
          properties: ["error": "Login failed: \(error.message)"])
        return LoginResult(
          success: false, error: "Login failed: \(error.message)", requiresTOTP: false)
      }
    } catch let error as CancellationError {
      print("❌ CancellationError details: \(String(describing: error))")
      print("⏱️ Error occurred at: \(Date())")

      // Try to get underlying error information
      let nsError = error as NSError
      print("❌ NSError: \(nsError)")

      // Check for connection-related error codes
      if nsError.domain == NSPOSIXErrorDomain {
        switch nsError.code {
        case 61:  // ECONNREFUSED
          let err = "Connection refused. Is the server running on 127.0.0.1:8001?"
          AnalyticsConfiguration.provideEventReporter().track(
            event: "login_failed", properties: ["error": err])
          return LoginResult(success: false, error: err, requiresTOTP: false)
        case 64:  // EHOSTDOWN
          let err = "Host is down. Check that the server is running."
          AnalyticsConfiguration.provideEventReporter().track(
            event: "login_failed", properties: ["error": err])
          return LoginResult(success: false, error: err, requiresTOTP: false)
        default:
          break
        }
      }
    } catch {
      print("❌ Error details: \(String(describing: error))")
      let err = "Login failed: \(error.localizedDescription)"
      AnalyticsConfiguration.provideEventReporter().track(
        event: "login_failed", properties: ["error": err])
      return LoginResult(success: false, error: err, requiresTOTP: false)
    }

    AnalyticsConfiguration.provideEventReporter().track(
      event: "login_failed", properties: ["error": "Unknown login error"])
    return LoginResult(success: false, error: "Unknown login error", requiresTOTP: false)
  }

  /// Exchange JWT token for OAuth2 access token
  private func exchangeJWTForOAuth2Token(jwtToken: String) async -> (
    success: Bool, error: String?
  ) {
    // Generate state for OAuth2 flow
    let state = UUID().uuidString

    // Build authorization URL
    guard var components = URLComponents(string: APIConfiguration.oauth2AuthorizeURL) else {
      return (false, "Failed to build authorization URL")
    }
    components.queryItems = [
      URLQueryItem(name: "response_type", value: "code"),
      URLQueryItem(name: "client_id", value: await oauth2ClientID()),
      URLQueryItem(name: "redirect_uri", value: APIConfiguration.serverURL),
      URLQueryItem(name: "state", value: state),
      URLQueryItem(name: "code_challenge_method", value: "plain"),
    ]

    guard let authURL = components.url else {
      return (false, "Failed to build authorization URL")
    }

    // Step 1: Get authorization code
    var request = URLRequest(url: authURL)
    request.httpMethod = "GET"
    request.setValue("Bearer \(jwtToken)", forHTTPHeaderField: "Authorization")

    // Configure session to not follow redirects
    let delegate = NoRedirectDelegate()
    let session = URLSession(configuration: .default, delegate: delegate, delegateQueue: nil)

    do {
      let (_, response) = try await session.data(for: request)

      guard let httpResponse = response as? HTTPURLResponse else {
        return (false, "Invalid response type")
      }

      // Check if this is a redirect response (3xx status code)
      guard (300...399).contains(httpResponse.statusCode) else {
        return (false, "Expected redirect response, got status code: \(httpResponse.statusCode)")
      }

      // Get the authorization code from the Location header
      guard let locationHeader = httpResponse.value(forHTTPHeaderField: "Location"),
        let locationURL = URL(string: locationHeader),
        let codeComponents = URLComponents(url: locationURL, resolvingAgainstBaseURL: false),
        let code = codeComponents.queryItems?.first(where: { $0.name == "code" })?.value
      else {
        return (false, "No authorization code in response")
      }

      print("📝 Received OAuth2 authorization code")

      // Step 2: Exchange code for access token
      return await exchangeCodeForToken(code: code)
    } catch {
      print("❌ Error getting authorization code: \(error)")
      return (false, "Failed to get authorization code: \(error.localizedDescription)")
    }
  }

  /// Exchange OAuth2 authorization code for access token
  private func exchangeCodeForToken(code: String) async -> (success: Bool, error: String?) {
    guard let tokenURL = URL(string: APIConfiguration.oauth2TokenURL) else {
      return (false, "Invalid token URL")
    }

    var request = URLRequest(url: tokenURL)
    request.httpMethod = "POST"
    request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")

    // Build form data
    let formData = [
      "grant_type": "authorization_code",
      "code": code,
      "redirect_uri": APIConfiguration.serverURL,
      "client_id": await oauth2ClientID(),
      "client_secret": await oauth2ClientSecret(),
    ]

    // Encode form data
    let formString = formData.map {
      "\($0.key)=\($0.value.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? $0.value)"
    }
    .joined(separator: "&")
    request.httpBody = formString.data(using: .utf8)

    do {
      let (data, response) = try await URLSession.shared.data(for: request)

      guard let httpResponse = response as? HTTPURLResponse else {
        return (false, "Invalid response type")
      }

      guard (200...299).contains(httpResponse.statusCode) else {
        let errorMessage = String(data: data, encoding: .utf8) ?? "Unknown error"
        return (false, "Token exchange failed: \(errorMessage)")
      }

      // Parse JSON response
      guard let json = try JSONSerialization.jsonObject(with: data) as? [String: Any],
        let accessToken = json["access_token"] as? String
      else {
        return (false, "Invalid token response format")
      }

      let refreshToken = json["refresh_token"] as? String ?? ""
      let expiresIn = json["expires_in"] as? Int ?? 86400  // Default to 24 hours

      // Calculate expiration time
      let expiresAt = Date().addingTimeInterval(TimeInterval(expiresIn))

      await MainActor.run {
        self.oauth2AccessToken = accessToken
        self.oauth2RefreshToken = refreshToken
        self.oauth2TokenExpiresAt = expiresAt
      }

      return (true, nil)
    } catch {
      print("❌ Error exchanging code for token: \(error)")
      return (false, "Failed to exchange code: \(error.localizedDescription)")
    }
  }

  /// Refresh OAuth2 access token using refresh token
  func refreshOAuth2Token() async -> Bool {
    // If using mock, delegate to mock manager
    if let mock = mockManager {
      let result = await mock.refreshOAuth2Token()
      // Sync state
      await MainActor.run {
        self.oauth2AccessToken = mock.oauth2AccessToken
        self.oauth2RefreshToken = mock.oauth2RefreshToken
        self.oauth2TokenExpiresAt = mock.oauth2TokenExpiresAt
      }
      return result
    }

    guard !oauth2RefreshToken.isEmpty else {
      print("⚠️ No OAuth2 refresh token available")
      return false
    }

    guard let tokenURL = URL(string: APIConfiguration.oauth2TokenURL) else {
      print("❌ Invalid token URL")
      return false
    }

    var request = URLRequest(url: tokenURL)
    request.httpMethod = "POST"
    request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")

    // Build form data for refresh
    let formData = [
      "grant_type": "refresh_token",
      "refresh_token": oauth2RefreshToken,
      "client_id": await oauth2ClientID(),
      "client_secret": await oauth2ClientSecret(),
    ]

    let formString = formData.map {
      "\($0.key)=\($0.value.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? $0.value)"
    }
    .joined(separator: "&")
    request.httpBody = formString.data(using: .utf8)

    do {
      let (data, response) = try await URLSession.shared.data(for: request)

      guard let httpResponse = response as? HTTPURLResponse,
        (200...299).contains(httpResponse.statusCode)
      else {
        print("❌ Token refresh failed")
        return false
      }

      guard let json = try JSONSerialization.jsonObject(with: data) as? [String: Any],
        let accessToken = json["access_token"] as? String
      else {
        print("❌ Invalid token response format")
        return false
      }

      let refreshToken = json["refresh_token"] as? String ?? oauth2RefreshToken
      let expiresIn = json["expires_in"] as? Int ?? 86400
      let expiresAt = Date().addingTimeInterval(TimeInterval(expiresIn))

      await MainActor.run {
        self.oauth2AccessToken = accessToken
        self.oauth2RefreshToken = refreshToken
        self.oauth2TokenExpiresAt = expiresAt
        self.persistCredentials()
      }

      print("✅ OAuth2 token refreshed successfully")
      return true
    } catch {
      print("❌ Error refreshing token: \(error)")
      return false
    }
  }

  /// Get OAuth2 access token, refreshing if needed
  func getOAuth2AccessToken() async -> String? {
    // If using mock, delegate to mock manager
    if let mock = mockManager {
      return await mock.getOAuth2AccessToken()
    }

    // Check if token is expired or will expire soon (within 5 minutes)
    if let expiresAt = oauth2TokenExpiresAt,
      expiresAt.timeIntervalSinceNow < 300
    {
      print("🔄 OAuth2 token expired or expiring soon, refreshing...")
      let refreshed = await refreshOAuth2Token()
      if !refreshed {
        // If refresh fails, try to get a new token using JWT
        if !accessToken.isEmpty {
          print("🔄 Attempting to get new OAuth2 token using JWT...")
          let result = await exchangeJWTForOAuth2Token(jwtToken: accessToken)
          if !result.success {
            return nil
          }
        } else {
          return nil
        }
      }
    }

    return oauth2AccessToken.isEmpty ? nil : oauth2AccessToken
  }

  // MARK: - OAuth2 client credentials (via platform Secrets)

  /// Resolves the OAuth2 client id from the platform secret source (env + Info.plist), with the
  /// existing Info.plist accessor as a fallback. Reads are observed (never logging the value).
  private func oauth2ClientID() async -> String {
    (try? await PlatformServices.shared.secrets.getSecret(name: "OAuth2ClientID"))
      ?? APIConfiguration.oauth2ClientID
  }

  private func oauth2ClientSecret() async -> String {
    (try? await PlatformServices.shared.secrets.getSecret(name: "OAuth2ClientSecret"))
      ?? APIConfiguration.oauth2ClientSecret
  }

  func register(input: RegistrationInput) async -> RegistrationResult {
    // If using mock, delegate to mock manager
    if let mock = mockManager {
      return await mock.register(input: input)
    }

    print("📝 Registration attempt for user: \(input.username)")

    // Build the registration input
    var registrationInput = Identity_UserRegistrationInput()
    registrationInput.emailAddress = input.emailAddress.trimmingCharacters(in: .whitespaces)
    registrationInput.username = input.username.trimmingCharacters(in: .whitespaces)
    registrationInput.password = input.password
    registrationInput.accountName = input.accountName.trimmingCharacters(in: .whitespaces)
    registrationInput.firstName = input.firstName.trimmingCharacters(in: .whitespaces)
    registrationInput.lastName = input.lastName.trimmingCharacters(in: .whitespaces)
    registrationInput.invitationToken = input.invitationToken.trimmingCharacters(in: .whitespaces)
    registrationInput.invitationID = input.invitationID.trimmingCharacters(in: .whitespaces)

    // Convert birthday to protobuf timestamp if provided
    if let birthday = input.birthday {
      var timestamp = SwiftProtobuf.Google_Protobuf_Timestamp()
      timestamp.seconds = Int64(birthday.timeIntervalSince1970)
      timestamp.nanos = Int32(
        (birthday.timeIntervalSince1970 - Double(timestamp.seconds)) * 1_000_000_000)
      registrationInput.birthday = timestamp
    }

    var requestMessage = Identity_CreateUserRequest()
    requestMessage.input = registrationInput

    print("📤 Creating gRPC client and sending registration request...")

    do {
      // Check for cancellation before starting
      try Task.checkCancellation()

      // Get or create the client manager
      let manager = try getClientManager()

      // Use the identity service client from the unified Client
      let response = try await manager.client.identity.createUser(
        requestMessage,
        options: manager.defaultCallOptions
      )

      // Check if registration was successful
      if response.hasCreated {
        print("✅ Registration successful")
        return RegistrationResult(success: true, error: nil)
      } else {
        print("⚠️ Response received but no creation result")
        return RegistrationResult(
          success: false, error: "No creation result received from server")
      }
    } catch let error as GRPCCore.RPCError {
      print("❌ RPC error code: \(error.code)")
      print("❌ RPC error message: \(error.message)")

      // Provide user-friendly error messages
      switch error.code {
      case .deadlineExceeded:
        return RegistrationResult(
          success: false, error: "Request timed out. Please check your connection.")
      case .unavailable:
        return RegistrationResult(
          success: false, error: "Server is unavailable. Please try again later.")
      case .alreadyExists:
        return RegistrationResult(
          success: false, error: "Username or email address already exists.")
      case .invalidArgument:
        return RegistrationResult(
          success: false, error: "Invalid registration data. Please check your input.")
      default:
        return RegistrationResult(
          success: false, error: "Registration failed: \(error.message)")
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

  /// If the error indicates session/auth failure (e.g. invalid context, unauthenticated),
  /// clear credentials so the user can re-authenticate.
  func invalidateCredentialsIfSessionError(_ error: Error) async {
    guard let rpcError = error as? GRPCCore.RPCError else { return }
    let shouldInvalidate: Bool
    switch rpcError.code {
    case .unauthenticated:
      shouldInvalidate = true
    case .internalError:
      shouldInvalidate = rpcError.message.contains("building session context data for user")
    default:
      shouldInvalidate = false
    }
    if shouldInvalidate {
      print("🔐 Session error detected, invalidating credentials: \(rpcError.message)")
      await logout()
    }
  }

  func logout() async {
    // Revoke server session (best-effort, using JWT so session ID is available)
    if !accessToken.isEmpty {
      do {
        let manager = try getClientManager()
        let metadata = manager.authenticatedMetadata(accessToken: self.accessToken)
        _ = try await manager.client.auth.revokeCurrentSession(
          Auth_RevokeCurrentSessionRequest(),
          metadata: metadata,
          options: manager.defaultCallOptions
        )
      } catch {
        // Proceed with logout even if revocation fails
      }
    }

    await DeviceTokenRegistrationService.shared.archiveCurrentDeviceToken(authManager: self)
    await CurrentUserService.shared.clear()
    logOutFromRevenueCat()
    AnalyticsConfiguration.provideEventReporter().reset()
    self.isAuthenticated = false
    self.username = ""
    self.accessToken = ""
    self.refreshToken = ""
    self.oauth2AccessToken = ""
    self.oauth2RefreshToken = ""
    self.oauth2TokenExpiresAt = nil
    self.userID = ""
    self.accountID = ""
    clearPersistedCredentials()
  }
}
