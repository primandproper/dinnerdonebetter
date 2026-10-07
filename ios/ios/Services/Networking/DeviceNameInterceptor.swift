//
//  DeviceNameInterceptor.swift
//  ios
//
//  Names this device on every call, so the server can record it beside each login it signs in
//  or renews. "Where you're signed in" then lists "iPhone, iOS 18.2" rather than a gRPC library's
//  user agent. It is shown to the person whose login it is, and decides nothing.
//

import Foundation
import GRPCCore

/// The metadata key the server reads a device's name from. See internal/authentication/devices
/// in the backend.
let deviceNameMetadataKey = "x-device-name"

@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
struct DeviceNameInterceptor: ClientInterceptor {
  let deviceName: String

  func intercept<Input: Sendable, Output: Sendable>(
    request: StreamingClientRequest<Input>,
    context: ClientContext,
    next: (
      _ request: StreamingClientRequest<Input>,
      _ context: ClientContext
    ) async throws -> StreamingClientResponse<Output>
  ) async throws -> StreamingClientResponse<Output> {
    var request = request
    if !deviceName.isEmpty {
      request.metadata.replaceOrAddString(deviceName, forKey: deviceNameMetadataKey)
    }
    return try await next(request, context)
  }

  /// This device: its model and its system, as "iPhone15,2, iOS 18.2". The device's own name
  /// ("Alex's iPhone") needs an entitlement and is not something to send a server anyway.
  ///
  /// Read from the kernel and the process rather than UIDevice, which is main-actor isolated
  /// and would make building a client a main-actor affair.
  static func describingThisDevice() -> DeviceNameInterceptor {
    var systemInfo = utsname()
    uname(&systemInfo)
    let machine =
      withUnsafeBytes(of: &systemInfo.machine) { bytes in
        String(bytes: bytes.prefix { $0 != 0 }, encoding: .utf8)
      } ?? "unknown"

    // The simulator reports the Mac's architecture; it says which device it is pretending to be.
    let model = ProcessInfo.processInfo.environment["SIMULATOR_MODEL_IDENTIFIER"] ?? machine
    let version = ProcessInfo.processInfo.operatingSystemVersion

    return DeviceNameInterceptor(
      deviceName: "\(model), iOS \(version.majorVersion).\(version.minorVersion)")
  }
}
