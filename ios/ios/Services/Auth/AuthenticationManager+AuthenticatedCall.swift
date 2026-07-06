//
//  AuthenticationManager+AuthenticatedCall.swift
//  ios
//
//  A single, instrumented entry point for authenticated gRPC calls. Collapses the
//  per-view-model boilerplate (get client → get token → build metadata → invalidate on
//  session error → log) into one call, and layers on the platform primitives:
//    • an observability `operation` span (correlated logs + Instruments signposts),
//    • W3C `traceparent` propagation so the mobile trace links to the Go backend's trace,
//    • retry (idempotent calls only) + circuit breaking around the transport.
//

import CircuitBreaking
import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2TransportServices
import Observability
import Retry

/// Errors surfaced by `authenticatedCall` before the RPC is attempted.
enum AuthenticatedCallError: LocalizedError {
  case noAccessToken

  var errorDescription: String? {
    switch self {
    case .noAccessToken:
      return "Could not obtain an access token. Please sign in again."
    }
  }
}

extension AuthenticationManager {
  /// The concrete client type used throughout the app.
  typealias APIClient = Client<HTTP2ClientTransport.TransportServices>

  /// Run an authenticated gRPC call with observability, trace propagation, and resilience.
  ///
  /// - Parameters:
  ///   - name: Operation/span name (defaults to the calling function).
  ///   - idempotent: When `true`, the call is retried with exponential backoff. Leave `false`
  ///     (the default) for mutations so they are never silently re-sent.
  ///   - body: Performs the RPC using the resolved client, auth+trace metadata, and call options.
  /// - Returns: The RPC response.
  @discardableResult
  func authenticatedCall<Response: Sendable>(
    _ name: String = #function,
    idempotent: Bool = false,
    _ body:
      @Sendable @escaping (APIClient, GRPCCore.Metadata, GRPCCore.CallOptions) async throws ->
      Response
  ) async throws -> Response {
    try await observer.operation(name) { op in
      let manager = try getClientManager()

      guard let token = await getOAuth2AccessToken() else {
        throw op.error(AuthenticatedCallError.noAccessToken, "obtaining OAuth2 access token")
      }

      var metadata = manager.authenticatedMetadata(accessToken: token)
      // Link this call to the backend trace: the operation just installed its span as the
      // current task-local context, so this emits *this* call's traceparent.
      if let context = SpanContextStore.current {
        metadata.addString(
          W3CPropagation.traceparent(for: context), forKey: W3CPropagation.traceparentHeader)
      }
      op.set("rpc", name)
      op.set("rpc.idempotent", idempotent)

      // Capture only Sendable values in the @Sendable transport closure — never `self` or the
      // non-Sendable ClientManager.
      let client = manager.client
      let options = manager.defaultCallOptions
      let run: @Sendable () async throws -> Response = {
        try await body(client, metadata, options)
      }

      do {
        if idempotent {
          return try await retry.execute { try await breaker.execute(run) }
        } else {
          return try await breaker.execute(run)
        }
      } catch {
        await invalidateCredentialsIfSessionError(error)
        throw op.error(error, "gRPC \(name)")
      }
    }
  }
}
