//
//  AuthenticationManager+AuthenticatedCall.swift
//  ios
//
//  A single, instrumented entry point for authenticated gRPC calls. Collapses the
//  per-view-model boilerplate (get client → get token → build metadata → invalidate on
//  session error → log) into one call. The token is the Session's: it refreshes ahead of
//  expiry, and refreshes and retries once when a call is refused UNAUTHENTICATED. On top of
//  that this layers the platform primitives:
//    • an observability `operation` span (correlated logs + Instruments signposts),
//    • W3C `traceparent` propagation so the mobile trace links to the Go backend's trace,
//    • retry (idempotent calls only) + circuit breaking around the transport.
//

import CircuitBreaking
import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2TransportServices
import Observability
import PlatformClient
import Retry

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

      // Link this call to the backend trace: the operation just installed its span as the
      // current task-local context, so this emits *this* call's traceparent.
      var trace = GRPCCore.Metadata()
      if let context = SpanContextStore.current {
        trace.addString(
          W3CPropagation.traceparent(for: context), forKey: W3CPropagation.traceparentHeader)
      }
      op.set("rpc", name)
      op.set("rpc.idempotent", idempotent)

      // Capture only Sendable values in the @Sendable transport closure — never `self` or the
      // non-Sendable ClientManager.
      let client = manager.client
      let session = manager.session
      let options = manager.defaultCallOptions
      let run: @Sendable () async throws -> Response = { [trace] in
        try await session.call { credentials in
          var metadata = trace
          metadata.add(contentsOf: credentials)
          return try await body(client, metadata, options)
        }
      }

      do {
        let breaker = self.breaker
        let guarded: @Sendable () async throws -> Response = {
          try await Self.underBreaker(breaker, run)
        }
        if idempotent {
          return try await retry.execute(guarded)
        } else {
          return try await guarded()
        }
      } catch {
        await invalidateCredentialsIfSessionError(error)
        throw op.error(error, "gRPC \(name)")
      }
    }
  }

  /// The account the Session's login is acting in. It is read off the token rather than asked
  /// of the server: the token is what every call is authorized as, so it is the answer by
  /// definition.
  func activeAccountID() async throws -> String {
    guard let held = try await getClientManager().session.held(), !held.activeAccountID.isEmpty
    else {
      throw NotSignedInError()
    }
    return held.activeAccountID
  }

  /// Runs `operation` under the breaker, counting only what says the backend is unhealthy.
  ///
  /// The breaker's own `execute` counts every error as a failure, which is wrong twice over for a
  /// call made through the Session: a call made while signed out throws before anything leaves
  /// the device, and a server that answered NOT_FOUND or UNAUTHENTICATED is a server that is up.
  /// Counting either let a signed-out burst trip the breaker and refuse the calls right after
  /// sign-in.
  static func underBreaker<Response: Sendable>(
    _ breaker: any CircuitBreaker,
    _ operation: @Sendable () async throws -> Response
  ) async throws -> Response {
    guard await breaker.canProceed() else { throw CircuitOpenError() }
    do {
      let response = try await operation()
      await breaker.recordSuccess()
      return response
    } catch {
      switch breakerOutcome(of: error) {
      case .failure: await breaker.recordFailure()
      case .success: await breaker.recordSuccess()
      case .neither: break
      }
      throw error
    }
  }

  enum BreakerOutcome { case success, failure, neither }

  /// What `error` says about the backend's health.
  static func breakerOutcome(of error: any Error) -> BreakerOutcome {
    if error is NotSignedInError || error is CancellationError {
      return .neither
    }
    guard let status = error.platformError else {
      return .failure
    }
    switch status.code {
    case .unavailable, .deadlineExceeded, .internalError, .unknown, .resourceExhausted:
      return .failure
    case .cancelled:
      return .neither
    default:
      return .success
    }
  }
}
