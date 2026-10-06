//
//  Client.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import GRPCNIOTransportHTTP2TransportServices
import Observability
import PlatformClient

/// A unified gRPC client that provides access to all service clients.
/// This is the Swift analog of the Go client in backend/pkg/client/client.go
///
/// This implementation follows the pattern from grpc-swift issue #2211:
/// https://github.com/grpc/grpc-swift/issues/2211
///
/// A single GRPCClient instance is reused across all service clients for efficient
/// connection management.
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
internal struct Client<Transport> where Transport: GRPCCore.ClientTransport {
  /// Sign-in service client: who the caller is, their logins, and their credentials. Signing
  /// in and out goes through the Session, which owns the tokens; this is for everything else.
  internal let signIn: Primandproper_Platform_Signin_V1_SignInService.Client<Transport>

  /// Identity service client: the directory of users, accounts, memberships and invitations.
  internal let identity: Primandproper_Platform_Identity_V1_IdentityService.Client<Transport>

  /// Internal operations service client
  internal let internalOps: Internalops_InternalOperations.Client<Transport>

  /// Meal planning service client
  internal let mealPlanning: Mealplanning_MealPlanningService.Client<Transport>

  /// Notifications service client
  internal let notifications:
    Primandproper_Platform_Notifications_V1_NotificationsService.Client<Transport>

  /// Settings service client
  internal let settings: Primandproper_Platform_Settings_V1_SettingsService.Client<Transport>

  /// Media registry service client: uploading, and reading back, the caller's own objects.
  internal let mediaRegistry:
    Primandproper_Platform_Mediaregistry_V1_MediaRegistryService.Client<Transport>

  /// Internal gRPC client - shared across all service clients
  private let grpcClient: GRPCCore.GRPCClient<Transport>

  /// Initialize a new client with a gRPC client.
  ///
  /// - Parameter grpcClient: The underlying gRPC client to use for all service clients
  internal init(grpcClient: GRPCCore.GRPCClient<Transport>) {
    self.grpcClient = grpcClient

    // Initialize all service clients with the same underlying gRPC client
    // This follows the best practice of reusing a single GRPCClient instance
    // across multiple service clients (see grpc-swift issue #2211)
    self.signIn = Primandproper_Platform_Signin_V1_SignInService.Client(wrapping: grpcClient)
    self.identity = Primandproper_Platform_Identity_V1_IdentityService.Client(wrapping: grpcClient)
    self.internalOps = Internalops_InternalOperations.Client(wrapping: grpcClient)
    self.mealPlanning = Mealplanning_MealPlanningService.Client(wrapping: grpcClient)
    self.notifications = Primandproper_Platform_Notifications_V1_NotificationsService.Client(
      wrapping: grpcClient)
    self.settings = Primandproper_Platform_Settings_V1_SettingsService.Client(
      wrapping: grpcClient)
    self.mediaRegistry = Primandproper_Platform_Mediaregistry_V1_MediaRegistryService.Client(
      wrapping: grpcClient)
  }

  /// Start the connection for this client.
  /// This should be called after creating the client to establish the connection.
  ///
  /// - Note: This method starts the connection asynchronously. The connection
  ///   will be established in the background.
  internal func startConnections() {
    Task {
      do {
        try await grpcClient.runConnections()
      } catch {
        PlatformServices.shared.logger("gRPCClient").error("starting gRPC connections", error)
      }
    }
  }
}

// MARK: - Class-based Client Manager (following grpc-swift issue #2211 pattern)

/// A class-based client manager that follows the pattern from grpc-swift issue #2211.
/// This provides explicit lifecycle management for the gRPC client and connections.
///
/// Example usage:
/// ```swift
/// let clientManager = try ClientManager(host: "127.0.0.1", port: 8001, store: store)
/// // Connections are automatically started
/// let me = try await clientManager.session.call { metadata in
///   try await clientManager.client.signIn.getSelf(.init(), metadata: metadata)
/// }
/// ```
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
internal class ClientManager<Transport: GRPCCore.ClientTransport> {
  /// The underlying gRPC transport client
  private let grpcTransportClient: GRPCCore.GRPCClient<Transport>

  /// The unified client providing access to all service clients
  internal let client: Client<Transport>

  /// The one Session in the process: it holds the login, refreshes it, and is what every
  /// authenticated call goes through. Two would present the same refresh token twice, which
  /// the server treats as theft and answers by revoking the login.
  internal let session: Session

  /// Default call options to use for all RPC calls.
  /// These can be overridden on a per-call basis.
  internal var defaultCallOptions: GRPCCore.CallOptions

  /// This install's push registration, made through `session` and over the same client. It is
  /// one per Session for the reason the Session is one per process: both act as the login the
  /// Session holds, and the registration is persisted under that login so a relaunch revokes
  /// what the last launch registered.
  internal let devices: Devices

  /// Initialize a new client manager over a gRPC client whose connections the caller runs.
  ///
  /// - Parameters:
  ///   - grpcClient: The client every service client and the Session share
  ///   - store: Where the Session keeps the login it holds
  ///   - deviceStore: Where `devices` keeps this install's push registration
  ///   - defaultCallOptions: Default call options to use for all RPC calls (default: 5 second timeout)
  internal init(
    client grpcClient: GRPCCore.GRPCClient<Transport>,
    store: any CredentialStore,
    deviceStore: any DeviceRegistrationStore = KeychainDeviceRegistrationStore(),
    defaultCallOptions: GRPCCore.CallOptions = {
      var options = GRPCCore.CallOptions.defaults
      options.timeout = .seconds(5)
      return options
    }()
  ) {
    self.grpcTransportClient = grpcClient

    // Create the unified client wrapper
    self.client = Client(grpcClient: grpcClient)
    self.session = Session(client: grpcClient, store: store)
    self.devices = Devices(session: session, client: grpcClient, store: deviceStore)

    // Store default call options
    self.defaultCallOptions = defaultCallOptions
  }

  /// Initialize a new client manager with a transport, and start its connections.
  ///
  /// - Parameters:
  ///   - transport: The transport to use for the gRPC client
  ///   - store: Where the Session keeps the login it holds
  ///   - deviceStore: Where `devices` keeps this install's push registration
  ///   - defaultCallOptions: Default call options to use for all RPC calls (default: 5 second timeout)
  /// - Throws: An error if the client cannot be created
  internal convenience init(
    transport: Transport,
    store: any CredentialStore,
    deviceStore: any DeviceRegistrationStore = KeychainDeviceRegistrationStore(),
    defaultCallOptions: GRPCCore.CallOptions = {
      var options = GRPCCore.CallOptions.defaults
      options.timeout = .seconds(5)
      return options
    }()
  ) throws {
    // Create a single GRPCClient instance
    let grpcClient = GRPCCore.GRPCClient(transport: transport)
    self.init(
      client: grpcClient, store: store, deviceStore: deviceStore,
      defaultCallOptions: defaultCallOptions)

    // Start the connection asynchronously (following issue #2211 pattern)
    Task {
      do {
        try await grpcClient.runConnections()
      } catch {
        PlatformServices.shared.logger("gRPCClient").error("starting gRPC connections", error)
      }
    }
  }

  /// Ends the login on the server: revokes this install's push registration while the login
  /// can still make the call, then signs out. Neither throws. A revocation that fails is
  /// logged and the sign-out goes ahead, since keeping someone signed in because their device
  /// could not be unregistered would be worse; the registration stays persisted, so the next
  /// sign-in on this device moves it to whoever that is.
  internal func signOut() async {
    do {
      try await devices.revoke()
    } catch {
      PlatformServices.shared.logger("ClientManager")
        .error("revoking the device registration", error)
    }
    await session.signOut()
  }

  /// Initialize a new client manager with HTTP2ClientTransport.
  ///
  /// - Parameters:
  ///   - host: The server host (e.g., "127.0.0.1" or "localhost")
  ///   - port: The server port (default: 8001)
  ///   - useTLS: Whether to use TLS for the connection (default: false for plaintext)
  ///   - store: Where the Session keeps the login it holds
  ///   - deviceStore: Where `devices` keeps this install's push registration
  ///   - defaultCallOptions: Default call options to use for all RPC calls (default: 5 second timeout)
  /// - Throws: An error if the transport cannot be created
  internal convenience init(
    host: String = "127.0.0.1",
    port: Int = 8001,
    useTLS: Bool = false,
    store: any CredentialStore,
    deviceStore: any DeviceRegistrationStore = KeychainDeviceRegistrationStore(),
    defaultCallOptions: GRPCCore.CallOptions = {
      var options = GRPCCore.CallOptions.defaults
      options.timeout = .seconds(5)
      return options
    }()
  ) throws where Transport == HTTP2ClientTransport.TransportServices {
    let transport: HTTP2ClientTransport.TransportServices
    if useTLS {
      transport = try HTTP2ClientTransport.TransportServices(
        target: .dns(host: host, port: port),
        transportSecurity: .tls
      )
    } else {
      transport = try HTTP2ClientTransport.TransportServices(
        target: .dns(host: host, port: port),
        transportSecurity: .plaintext
      )
    }
    try self.init(
      transport: transport, store: store, deviceStore: deviceStore,
      defaultCallOptions: defaultCallOptions)
  }

  /// Get call options by merging default options with any overrides.
  /// Properties in `overrides` will take precedence over defaults.
  ///
  /// - Parameter overrides: Call options that override the defaults
  /// - Returns: Merged call options
  internal func callOptions(overriding overrides: GRPCCore.CallOptions = .defaults)
    -> GRPCCore.CallOptions
  {
    var merged = defaultCallOptions

    // Override timeout if specified in overrides
    if overrides.timeout != nil {
      merged.timeout = overrides.timeout
    }

    // Override compression if specified
    if overrides.compression != nil {
      merged.compression = overrides.compression
    }

    return merged
  }
}

// MARK: - Factory Methods

@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
internal func buildClient<Transport: GRPCCore.ClientTransport>(transport: Transport) -> Client<
  Transport
> {
  let grpcClient = GRPCCore.GRPCClient(transport: transport)
  let client = Client(grpcClient: grpcClient)
  // Start connections following the issue #2211 pattern
  client.startConnections()
  return client
}

/// Build an unauthenticated gRPC client using TransportServices with plaintext security.
/// Connections are automatically started following the pattern from grpc-swift issue #2211.
///
/// - Parameters:
///   - host: The server host (e.g., "127.0.0.1" or "localhost")
///   - port: The server port (default: 8001)
/// - Returns: A new client instance with connections started
/// - Throws: An error if the transport cannot be created
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
internal func buildUnauthenticatedClient(
  host: String = "127.0.0.1",
  port: Int = 8001
) throws -> Client<HTTP2ClientTransport.TransportServices> {
  let transport = try HTTP2ClientTransport.TransportServices(
    target: .dns(host: host, port: port),
    transportSecurity: .plaintext
  )
  return buildClient(transport: transport)
}

/// Build an unauthenticated gRPC client using TransportServices with plaintext security.
/// Attempts multiple connection strategies (IPv4, IPv6, DNS) for better compatibility.
///
/// - Parameters:
///   - host: The server host (e.g., "127.0.0.1", "localhost", or nil for auto-detection)
///   - port: The server port (default: 8001)
/// - Returns: A new client instance
/// - Throws: An error if all transport creation attempts fail
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
internal func buildUnauthenticatedClientWithFallback(
  host: String? = nil,
  port: Int = 8001
) throws -> Client<HTTP2ClientTransport.TransportServices> {
  // If a specific host is provided, use it directly
  if let host = host {
    return try buildUnauthenticatedClient(host: host, port: port)
  }

  // Try IPv4 first (most reliable on iOS Simulator)
  do {
    return try buildUnauthenticatedClient(host: "127.0.0.1", port: port)
  } catch {
    // Try IPv6 next
    do {
      let transport = try HTTP2ClientTransport.TransportServices(
        target: .ipv6(address: "::1", port: port),
        transportSecurity: .plaintext
      )
      return buildClient(transport: transport)
    } catch {
      // Fallback to DNS resolution
      let transport = try HTTP2ClientTransport.TransportServices(
        target: .dns(host: "localhost", port: port),
        transportSecurity: .plaintext
      )
      return buildClient(transport: transport)
    }
  }
}

/// Execute a closure with a gRPC client, ensuring proper lifecycle management.
/// This is similar to `withGRPCClient` but uses the unified Client wrapper.
///
/// - Parameters:
///   - host: The server host (e.g., "127.0.0.1" or "localhost")
///   - port: The server port (default: 8001)
///   - body: A closure that receives the Client and returns a result
/// - Returns: The result from the closure
/// - Throws: Any error thrown by the transport creation or the closure
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
internal func withClient<Result>(
  host: String = "127.0.0.1",
  port: Int = 8001,
  body: @Sendable @escaping (Client<HTTP2ClientTransport.TransportServices>) async throws -> Result
) async throws -> Result {
  let transport = try HTTP2ClientTransport.TransportServices(
    target: .dns(host: host, port: port),
    transportSecurity: .plaintext
  )
  let grpcClient = GRPCCore.GRPCClient(transport: transport)
  let client = Client(grpcClient: grpcClient)
  client.startConnections()
  return try await body(client)
}
