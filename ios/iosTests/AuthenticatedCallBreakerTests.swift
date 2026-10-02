//
//  AuthenticatedCallBreakerTests.swift
//  iosTests
//
//  The circuit breaker around authenticatedCall counts only what says the backend is down. A
//  call made while signed out, or one the server answered with a status, used to count, and a
//  burst of them tripped the breaker for the calls made right after sign-in.
//

import CircuitBreaking
import Foundation
import GRPCCore
import PlatformClient
import Testing

@testable import ios

struct AuthenticatedCallBreakerTests {
  /// The app's own breaker settings: half of at least five calls failing opens it.
  private func breaker() -> StandardCircuitBreaker<ContinuousClock> {
    StandardCircuitBreaker(name: "test", errorRatePercentage: 50, minimumSampleThreshold: 5)
  }

  private func burst(of error: any Error, on breaker: any CircuitBreaker) async {
    for _ in 0..<20 {
      _ = try? await AuthenticationManager.underBreaker(breaker) { () async throws -> Int in
        throw error
      }
    }
  }

  @Test("calls made while signed out never open the breaker")
  func signedOutBurst() async {
    let breaker = breaker()
    await burst(of: NotSignedInError(), on: breaker)
    #expect(await breaker.canProceed())
  }

  @Test("a server answering with a refusal is a server that is up")
  func refusalsBurst() async {
    let breaker = breaker()
    await burst(of: RPCError(code: .unauthenticated, message: "no"), on: breaker)
    await burst(of: RPCError(code: .notFound, message: "no"), on: breaker)
    #expect(await breaker.canProceed())
  }

  @Test("an unreachable server opens the breaker")
  func unavailableBurst() async {
    let breaker = breaker()
    await burst(of: RPCError(code: .unavailable, message: "down"), on: breaker)
    #expect(!(await breaker.canProceed()))
  }

  @Test("outcomes are read through PlatformError")
  func classification() {
    #expect(AuthenticationManager.breakerOutcome(of: NotSignedInError()) == .neither)
    #expect(AuthenticationManager.breakerOutcome(of: CancellationError()) == .neither)
    #expect(
      AuthenticationManager.breakerOutcome(of: PlatformError(RPCError(code: .deadlineExceeded, message: "")))
        == .failure)
    #expect(
      AuthenticationManager.breakerOutcome(of: PlatformError(RPCError(code: .alreadyExists, message: "")))
        == .success)
  }
}
