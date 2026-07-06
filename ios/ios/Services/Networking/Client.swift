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
  /// Auth service client
  internal let auth: Auth_AuthService.Client<Transport>

  /// Identity service client
  internal let identity: Identity_IdentityService.Client<Transport>

  /// Audit service client
  internal let audit: Audit_AuditService.Client<Transport>

  /// Data privacy service client
  internal let dataPrivacy: Dataprivacy_DataPrivacyService.Client<Transport>

  /// Internal operations service client
  internal let internalOps: Internalops_InternalOperations.Client<Transport>

  /// Meal planning service client
  internal let mealPlanning: Mealplanning_MealPlanningService.Client<Transport>

  /// Notifications service client
  internal let notifications: Notifications_UserNotificationsService.Client<Transport>

  /// OAuth service client
  internal let oauth: Oauth_OAuthService.Client<Transport>

  /// Settings service client
  internal let settings: Settings_SettingsService.Client<Transport>

  /// Webhooks service client
  internal let webhooks: Webhooks_WebhooksService.Client<Transport>

  /// Uploaded media service client
  internal let uploadedMedia: UploadedMedia_UploadedMediaService.Client<Transport>

  /// Analytics service client
  internal let analytics: Analytics_AnalyticsService.Client<Transport>

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
    self.auth = Auth_AuthService.Client(wrapping: grpcClient)
    self.identity = Identity_IdentityService.Client(wrapping: grpcClient)
    self.audit = Audit_AuditService.Client(wrapping: grpcClient)
    self.dataPrivacy = Dataprivacy_DataPrivacyService.Client(wrapping: grpcClient)
    self.internalOps = Internalops_InternalOperations.Client(wrapping: grpcClient)
    self.mealPlanning = Mealplanning_MealPlanningService.Client(wrapping: grpcClient)
    self.notifications = Notifications_UserNotificationsService.Client(wrapping: grpcClient)
    self.oauth = Oauth_OAuthService.Client(wrapping: grpcClient)
    self.settings = Settings_SettingsService.Client(wrapping: grpcClient)
    self.webhooks = Webhooks_WebhooksService.Client(wrapping: grpcClient)
    self.uploadedMedia = UploadedMedia_UploadedMediaService.Client(wrapping: grpcClient)
    self.analytics = Analytics_AnalyticsService.Client(wrapping: grpcClient)
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
/// let clientManager = try ClientManager(host: "127.0.0.1", port: 8001)
/// // Connections are automatically started
/// let response = try await clientManager.client.auth.loginForToken(request)
/// ```
@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
internal class ClientManager<Transport: GRPCCore.ClientTransport> {
  /// The underlying gRPC transport client
  private let grpcTransportClient: GRPCCore.GRPCClient<Transport>

  /// The unified client providing access to all service clients
  internal let client: Client<Transport>

  /// Default call options to use for all RPC calls.
  /// These can be overridden on a per-call basis.
  internal var defaultCallOptions: GRPCCore.CallOptions

  /// Initialize a new client manager with a transport.
  ///
  /// - Parameters:
  ///   - transport: The transport to use for the gRPC client
  ///   - defaultCallOptions: Default call options to use for all RPC calls (default: 5 second timeout)
  /// - Throws: An error if the client cannot be created
  internal init(
    transport: Transport,
    defaultCallOptions: GRPCCore.CallOptions = {
      var options = GRPCCore.CallOptions.defaults
      options.timeout = .seconds(5)
      return options
    }()
  ) throws {
    // Create a single GRPCClient instance
    self.grpcTransportClient = GRPCCore.GRPCClient(transport: transport)

    // Create the unified client wrapper
    self.client = Client(grpcClient: grpcTransportClient)

    // Store default call options
    self.defaultCallOptions = defaultCallOptions

    // Start the connection asynchronously (following issue #2211 pattern)
    Task {
      do {
        try await grpcTransportClient.runConnections()
      } catch {
        PlatformServices.shared.logger("gRPCClient").error("starting gRPC connections", error)
      }
    }
  }

  /// Initialize a new client manager with HTTP2ClientTransport.
  ///
  /// - Parameters:
  ///   - host: The server host (e.g., "127.0.0.1" or "localhost")
  ///   - port: The server port (default: 8001)
  ///   - useTLS: Whether to use TLS for the connection (default: false for plaintext)
  ///   - defaultCallOptions: Default call options to use for all RPC calls (default: 5 second timeout)
  /// - Throws: An error if the transport cannot be created
  internal convenience init(
    host: String = "127.0.0.1",
    port: Int = 8001,
    useTLS: Bool = false,
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
    try self.init(transport: transport, defaultCallOptions: defaultCallOptions)
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

  /// Get authenticated metadata with authorization header.
  ///
  /// - Parameter accessToken: The access token to include in the authorization header
  /// - Returns: Metadata dictionary with authorization header
  internal func authenticatedMetadata(accessToken: String) -> GRPCCore.Metadata {
    return ["authorization": "Bearer \(accessToken)"]
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
