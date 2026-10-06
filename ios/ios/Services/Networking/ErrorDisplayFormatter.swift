//
//  ErrorDisplayFormatter.swift
//  ios
//
//  Formats errors for user display. When the server is unavailable, shows a
//  friendly "cooks in the kitchen" message instead of technical error details.
//

import Darwin
import Foundation
import GRPCCore
import Observability
import PlatformClient
import SwiftUI

/// Display info for showing an error to the user.
struct ErrorDisplay {
  let message: String
  let title: String
  let icon: String
  let iconColor: Color
}

enum ErrorDisplayFormatter {

  /// Server-down display: friendly cooking-themed message for when the backend is unreachable.
  private static let serverDownDisplay = ErrorDisplay(
    message:
      "Our cooks are out of the kitchen, but they'll be back shortly. Please try again in a moment!",
    title: "The kitchen is closed",
    icon: "fork.knife",
    iconColor: DSTheme.Colors.tertiary  // Warm, friendly cooking vibe
  )

  /// Generic error display for unexpected or non-connectivity errors.
  private static func genericDisplay(for error: Error, context: String) -> ErrorDisplay {
    ErrorDisplay(
      message: "Failed to \(context): \(error.localizedDescription)",
      title: "Error",
      icon: "exclamationmark.triangle",
      iconColor: DSTheme.Colors.warning
    )
  }

  /// Returns true if the error indicates the server is down or unreachable.
  static func isServerDown(_ error: Error) -> Bool {
    // Whether a status means the server could not answer right now is platform-client's
    // `isTransient`, not a list of codes kept here.
    if let refusal = error.platformError {
      return isTransient(refusal)
    }

    // Check for connection refused, timeout, or other network errors in NSError
    let nsError = error as NSError
    if nsError.domain == NSPOSIXErrorDomain {
      // Connection refused, connection reset, host unreachable, etc.
      let connectionErrors: [Int32] = [
        ECONNREFUSED,  // Connection refused
        ECONNRESET,  // Connection reset
        ETIMEDOUT,  // Operation timed out
        ENETUNREACH,  // Network is unreachable
        EHOSTUNREACH,  // No route to host
      ]
      return connectionErrors.contains(Int32(nsError.code))
    }

    // Check underlying error (e.g. URLSession wraps NSError)
    if let underlying = nsError.userInfo[NSUnderlyingErrorKey] as? Error {
      return isServerDown(underlying)
    }

    return false
  }

  /// Formats an error for display. Uses friendly server-down messaging when appropriate.
  static func format(_ error: Error, context: String = "load data") -> ErrorDisplay {
    if isServerDown(error) {
      return serverDownDisplay
    }
    return genericDisplay(for: error, context: context)
  }
}

extension Error {
  /// The server's answer behind this error, when there is one. `authenticatedCall` wraps what
  /// it throws in an `ObservabilityError`, and the Session hands statuses back as
  /// PlatformErrors rather than RPCErrors, so neither `as? RPCError` nor `as? PlatformError`
  /// sees through to it on its own.
  var platformError: PlatformError? {
    if let wrapped = self as? ObservabilityError {
      return wrapped.underlying.platformError
    }
    return toPlatformError(self) as? PlatformError
  }
}
