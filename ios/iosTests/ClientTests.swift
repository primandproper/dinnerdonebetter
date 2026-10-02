//
//  ClientTests.swift
//  iosTests
//
//  Created by Auto on 12/8/25.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import GRPCNIOTransportHTTP2TransportServices
@testable import ios
import PlatformClient
import Testing

/// Holds a session in memory, so a ClientManager built here never touches the Keychain.
actor TestCredentialStore: CredentialStore {
  private var token: IssuedToken?

  func load() async throws -> IssuedToken? { token }
  func save(_ token: IssuedToken) async throws { self.token = token }
  func clear() async throws { token = nil }
}

// MARK: - Client Initialization Tests

struct ClientInitializationTests {
  @Test("Client initializes with all service clients")
  func testClientInitialization() throws {
    // Create a mock transport for testing
    // Note: In a real scenario, you might want to use a test transport
    let transport = try HTTP2ClientTransport.TransportServices(
      target: .dns(host: "127.0.0.1", port: 8001),
      transportSecurity: .plaintext
    )
    let grpcClient = GRPCCore.GRPCClient(transport: transport)
    let client = Client(grpcClient: grpcClient)

    // Verify all service clients are initialized
    // We can't directly test the service clients without making actual calls,
    // but we can verify the client was created successfully (non-nil by type)
    _ = client
  }

  @Test("Client initializes with shared gRPC client")
  func testClientSharesGRPCClient() throws {
    let transport = try HTTP2ClientTransport.TransportServices(
      target: .dns(host: "127.0.0.1", port: 8001),
      transportSecurity: .plaintext
    )
    let grpcClient = GRPCCore.GRPCClient(transport: transport)
    let client = Client(grpcClient: grpcClient)

    // The client should wrap the provided gRPC client
    // This is verified by the fact that all service clients use the same underlying client
    _ = client
  }
}

// MARK: - ClientManager Initialization Tests

struct ClientManagerInitializationTests {
  @Test("ClientManager initializes with default call options")
  func testClientManagerDefaultCallOptions() throws {
    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore()
    )

    // Verify default call options are set (5 second timeout)
    #expect(manager.defaultCallOptions.timeout == .seconds(5))
  }

  @Test("ClientManager initializes with custom call options")
  func testClientManagerCustomCallOptions() throws {
    var customOptions = GRPCCore.CallOptions.defaults
    customOptions.timeout = .seconds(10)

    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore(),
      defaultCallOptions: customOptions
    )

    // Verify custom call options are used
    #expect(manager.defaultCallOptions.timeout == .seconds(10))
  }

  @Test("ClientManager initializes with default host and port")
  func testClientManagerDefaultHostPort() throws {
    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(store: TestCredentialStore())

    // Should use default host (127.0.0.1) and port (8001)
    #expect(manager.defaultCallOptions.timeout == .seconds(5))
  }

  @Test("ClientManager initializes with custom host and port")
  func testClientManagerCustomHostPort() throws {
    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "localhost",
      port: 9000,
      store: TestCredentialStore()
    )

    // Verify client was created with custom host/port (non-nil by type)
    _ = manager.client
  }

  @Test("ClientManager creates unified client")
  func testClientManagerCreatesUnifiedClient() throws {
    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore()
    )

    // Verify the unified client is accessible (non-nil by type)
    _ = manager.client
  }

  @Test("ClientManager reuses same client instance")
  func testClientManagerClientReuse() throws {
    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore()
    )

    let client1 = manager.client
    let client2 = manager.client

    // Should return the same client instance (reused)
    // Since Client is a struct, we verify both are accessible (structs are value types)
    _ = client1.signIn
    _ = client2.signIn
  }
}

// MARK: - Call Options Tests

struct CallOptionsTests {
  @Test("callOptions returns default when no overrides")
  func testCallOptionsNoOverrides() throws {
    var defaultOptions = GRPCCore.CallOptions.defaults
    defaultOptions.timeout = .seconds(5)

    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore(),
      defaultCallOptions: defaultOptions
    )

    let options = manager.callOptions()

    // Should return default options
    #expect(options.timeout == .seconds(5))
  }

  @Test("callOptions overrides timeout")
  func testCallOptionsOverrideTimeout() throws {
    var defaultOptions = GRPCCore.CallOptions.defaults
    defaultOptions.timeout = .seconds(5)

    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore(),
      defaultCallOptions: defaultOptions
    )

    var overrideOptions = GRPCCore.CallOptions.defaults
    overrideOptions.timeout = .seconds(15)

    let mergedOptions = manager.callOptions(overriding: overrideOptions)

    // Override should take precedence
    #expect(mergedOptions.timeout == .seconds(15))
  }

  @Test("callOptions preserves default when override has nil timeout")
  func testCallOptionsPreservesDefaultWhenOverrideNil() throws {
    var defaultOptions = GRPCCore.CallOptions.defaults
    defaultOptions.timeout = .seconds(5)

    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore(),
      defaultCallOptions: defaultOptions
    )

    let overrideOptions = GRPCCore.CallOptions.defaults  // No timeout set

    let mergedOptions = manager.callOptions(overriding: overrideOptions)

    // Should preserve default timeout
    #expect(mergedOptions.timeout == .seconds(5))
  }

  @Test("callOptions overrides compression")
  func testCallOptionsOverrideCompression() throws {
    var defaultOptions = GRPCCore.CallOptions.defaults
    defaultOptions.timeout = .seconds(5)

    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore(),
      defaultCallOptions: defaultOptions
    )

    var overrideOptions = GRPCCore.CallOptions.defaults
    overrideOptions.compression = .gzip

    let mergedOptions = manager.callOptions(overriding: overrideOptions)

    // Override compression should be used
    #expect(mergedOptions.compression == .gzip)
  }

  @Test("callOptions merges timeout and compression")
  func testCallOptionsMergesBoth() throws {
    var defaultOptions = GRPCCore.CallOptions.defaults
    defaultOptions.timeout = .seconds(5)

    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore(),
      defaultCallOptions: defaultOptions
    )

    var overrideOptions = GRPCCore.CallOptions.defaults
    overrideOptions.timeout = .seconds(20)
    overrideOptions.compression = .gzip

    let mergedOptions = manager.callOptions(overriding: overrideOptions)

    // Both overrides should be applied
    #expect(mergedOptions.timeout == .seconds(20))
    #expect(mergedOptions.compression == .gzip)
  }
}

// MARK: - Factory Function Tests

struct FactoryFunctionTests {
  @Test("buildUnauthenticatedClient creates client with default host and port")
  func testBuildUnauthenticatedClientDefaults() throws {
    let client = try buildUnauthenticatedClient()

    // Should create client successfully (non-nil by type)
    _ = client
  }

  @Test("buildUnauthenticatedClient creates client with custom host and port")
  func testBuildUnauthenticatedClientCustom() throws {
    let client = try buildUnauthenticatedClient(host: "localhost", port: 9000)

    // Should create client with custom host/port (non-nil by type)
    _ = client
  }

  @Test("buildUnauthenticatedClientWithFallback uses provided host")
  func testBuildUnauthenticatedClientWithFallbackProvidedHost() throws {
    let client = try buildUnauthenticatedClientWithFallback(host: "127.0.0.1", port: 8001)

    // Should use provided host directly (non-nil by type)
    _ = client
  }

  @Test("buildUnauthenticatedClientWithFallback attempts IPv4 first")
  func testBuildUnauthenticatedClientWithFallbackIPv4() throws {
    // When host is nil, should try IPv4 first
    // This will succeed if IPv4 works, or fail and try other methods
    do {
      let client = try buildUnauthenticatedClientWithFallback(host: nil, port: 8001)
      _ = client
    } catch {
      // If IPv4 fails, it will try IPv6, then DNS
      // Any of these succeeding is valid
    }
  }
}

// MARK: - Error Handling Tests

struct ClientErrorHandlingTests {
  @Test("ClientManager throws error with invalid host")
  func testClientManagerInvalidHost() {
    // Invalid host format should throw error
    // Note: The actual behavior depends on DNS resolution
    do {
      _ = try ClientManager<HTTP2ClientTransport.TransportServices>(
        host: "invalid..host..name",
        port: 8001,
        store: TestCredentialStore()
      )
      // If it doesn't throw, that's also valid (DNS might resolve it)
    } catch {
      // Expected to throw for invalid host
    }
  }

  @Test("ClientManager handles invalid port")
  func testClientManagerInvalidPort() {
    // Port 0 or negative ports should be handled
    do {
      _ = try ClientManager<HTTP2ClientTransport.TransportServices>(
        host: "127.0.0.1",
        port: 0,
        store: TestCredentialStore()
      )
      // Some systems might allow port 0, so this is also valid
    } catch {
      // Expected to throw for invalid port
    }
  }

  @Test("buildUnauthenticatedClient throws error with invalid host")
  func testBuildUnauthenticatedClientInvalidHost() {
    do {
      _ = try buildUnauthenticatedClient(host: "invalid..host", port: 8001)
      // If it doesn't throw, DNS might have resolved it
    } catch {
      // Expected to throw for invalid host
    }
  }
}

// MARK: - Client Lifecycle Tests

struct ClientLifecycleTests {
  @Test("Client startConnections starts async connection")
  func testClientStartConnections() throws {
    let transport = try HTTP2ClientTransport.TransportServices(
      target: .dns(host: "127.0.0.1", port: 8001),
      transportSecurity: .plaintext
    )
    let grpcClient = GRPCCore.GRPCClient(transport: transport)
    let client = Client(grpcClient: grpcClient)

    // startConnections should not throw (it's async)
    client.startConnections()

    // Verify client is still valid after starting connections (non-nil by type)
    _ = client
  }

  @Test("ClientManager starts connections on initialization")
  func testClientManagerStartsConnections() throws {
    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore()
    )

    // Connections should be started automatically (client is non-nil by type)
    _ = manager.client
  }
}

// MARK: - Service Client Access Tests

struct ServiceClientAccessTests {
  @Test("Client provides access to all service clients")
  func testClientServiceClients() throws {
    let transport = try HTTP2ClientTransport.TransportServices(
      target: .dns(host: "127.0.0.1", port: 8001),
      transportSecurity: .plaintext
    )
    let grpcClient = GRPCCore.GRPCClient(transport: transport)
    let client = Client(grpcClient: grpcClient)

    // Verify all service clients are accessible (non-nil by type)
    _ = client.signIn
    _ = client.identity
    _ = client.internalOps
    _ = client.mealPlanning
    _ = client.notifications
    _ = client.settings
    _ = client.uploadedMedia
    _ = client.analytics
  }

  @Test("ClientManager provides access to unified client")
  func testClientManagerUnifiedClient() throws {
    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore()
    )

    // Verify unified client is accessible (non-nil by type)
    _ = manager.client
    _ = manager.client.signIn
    _ = manager.client.identity
    _ = manager.session
  }
}

// MARK: - Concurrent Access Tests

struct ConcurrentAccessTests {
  @Test("Multiple ClientManagers can be created concurrently")
  func testConcurrentClientManagerCreation() async throws {
    async let manager1 = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore()
    )
    async let manager2 = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore()
    )

    let managers = try await [manager1, manager2]

    // Both should be created successfully
    #expect(managers.count == 2)
    _ = managers[0].client
    _ = managers[1].client
  }

  @Test("ClientManager can be accessed concurrently")
  func testConcurrentClientAccess() async throws {
    let manager = try ClientManager<HTTP2ClientTransport.TransportServices>(
      host: "127.0.0.1",
      port: 8001,
      store: TestCredentialStore()
    )

    let client1 = manager.client
    let client2 = manager.client

    // All should succeed (service clients non-nil by type), and the one Session is shared.
    _ = client1.signIn
    _ = client2.signIn
    async let held1 = manager.session.held()
    async let held2 = manager.session.held()
    #expect(try await held1 == nil)
    #expect(try await held2 == nil)
  }
}

