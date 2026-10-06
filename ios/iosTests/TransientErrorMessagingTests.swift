//
//  TransientErrorMessagingTests.swift
//  iosTests
//
//  Whether a failure means "the server could not answer right now" is platform-client's
//  `isTransient`, not a list of codes kept in the app. These pin what each surface tells a
//  person on either side of that line.
//

import Foundation
import GRPCCore
import PlatformClient
import Testing

@testable import ios

private func refusal(
  _ code: RPCError.Code, _ message: String = "refused", reason: Reason? = nil
) -> PlatformError {
  PlatformError(code: code, serverMessage: message, reason: reason)
}

struct TransientErrorMessagingTests {
  @Test(
    "a transient sign-in failure says to try again",
    arguments: [RPCError.Code.unavailable, .deadlineExceeded, .resourceExhausted])
  func transientSignIn(code: RPCError.Code) {
    #expect(
      AuthenticationManager.signInRefusalMessage(refusal(code))
        == AuthenticationManager.transientRefusalMessage)
  }

  @Test(
    "a failure that is the server's bug, or the caller backing out, is not called an outage",
    arguments: [RPCError.Code.internalError, .unknown, .cancelled])
  func notTransientSignIn(code: RPCError.Code) {
    let message = AuthenticationManager.signInRefusalMessage(refusal(code, "boom"))
    #expect(message != AuthenticationManager.transientRefusalMessage)
    #expect(message.contains("boom"))
  }

  @Test("a wrong password says so, by its reason")
  func invalidCredentials() {
    let error = refusal(.unauthenticated, reason: .signIn(.invalidCredentials))
    #expect(
      AuthenticationManager.signInRefusalMessage(error) == "Invalid username or password.")
  }

  @Test(
    "a transient upload failure says to try again",
    arguments: [RPCError.Code.unavailable, .deadlineExceeded, .resourceExhausted])
  func transientUpload(code: RPCError.Code) {
    #expect(UploadErrorFormatter.formatRPCError(refusal(code)).contains("Try again"))
  }

  @Test("an upload refused for a reason the person can act on names that reason")
  func nonTransientUpload() {
    #expect(
      UploadErrorFormatter.formatRPCError(refusal(.invalidArgument, "too big"))
        == "Invalid request: too big")
    #expect(UploadErrorFormatter.formatRPCError(refusal(.cancelled)) == "The upload was cancelled.")
  }

  @Test("the server-down screen follows isTransient")
  func serverDown() {
    #expect(ErrorDisplayFormatter.isServerDown(refusal(.unavailable)))
    #expect(ErrorDisplayFormatter.isServerDown(refusal(.resourceExhausted)))
    #expect(!ErrorDisplayFormatter.isServerDown(refusal(.internalError)))
    #expect(!ErrorDisplayFormatter.isServerDown(refusal(.cancelled)))
  }

  @Test("a transient password-change failure says to try again")
  @MainActor
  func transientPasswordChange() {
    #expect(
      ChangePasswordViewModel.message(for: refusal(.unavailable))
        == AuthenticationManager.transientRefusalMessage)
  }
}
