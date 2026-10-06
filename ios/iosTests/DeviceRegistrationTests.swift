//
//  DeviceRegistrationTests.swift
//  iosTests
//
//  Push registration goes through platform-client's `Devices`, which persists what it
//  registered. The local service it replaced kept the device ID in memory only, so a sign-out
//  after a relaunch had nothing to revoke and the last person's device kept receiving their
//  notifications.
//
//  These drive a real ClientManager — its Session and its Devices — against a gRPC server on a
//  loopback port. platform-client's FakeServer would be the natural fit, but linking
//  PlatformClientTesting into this hosted test bundle statically links a second copy of
//  GRPCCore beside the app's, and the first GRPCClient built here crashes.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2Posix
import GRPCNIOTransportHTTP2TransportServices
import GRPCProtobuf
import PlatformClient
import Synchronization
import SwiftProtobuf
import Testing

@testable import ios

private typealias NotificationsRPC = Primandproper_Platform_Notifications_V1_NotificationsService
private typealias SignInRPC = Primandproper_Platform_Signin_V1_SignInService
private typealias RegisterRequest = Primandproper_Platform_Notifications_V1_RegisterDeviceRequest
private typealias RegisterResponse = Primandproper_Platform_Notifications_V1_RegisterDeviceResponse
private typealias RevokeRequest = Primandproper_Platform_Notifications_V1_RevokeDeviceRequest
private typealias RevokeResponse = Primandproper_Platform_Notifications_V1_RevokeDeviceResponse
private typealias SignOutRequest = Primandproper_Platform_Signin_V1_SignOutRequest
private typealias SignOutResponse = Primandproper_Platform_Signin_V1_SignOutResponse

private let registerMethod = NotificationsRPC.Method.RegisterDevice.descriptor
private let revokeMethod = NotificationsRPC.Method.RevokeDevice.descriptor
private let signOutMethod = SignInRPC.Method.SignOut.descriptor

/// Holds a registration in memory, as the Keychain store holds one across launches.
private actor TestDeviceRegistrationStore: DeviceRegistrationStore {
  private var registration: DeviceRegistration?

  func load() async throws -> DeviceRegistration? { registration }
  func save(_ registration: DeviceRegistration) async throws { self.registration = registration }
  func clear() async throws { registration = nil }
}

/// A login good for the next hour, as the Session would have saved one.
private func heldLogin() -> IssuedToken {
  var token = IssuedToken()
  token.token = "access-\(UUID().uuidString)"
  token.tokenID = UUID().uuidString
  token.expiresAt = Google_Protobuf_Timestamp(date: Date().addingTimeInterval(60 * 60))
  token.refreshToken = "refresh-\(UUID().uuidString)"
  token.refreshTokenExpiresAt = Google_Protobuf_Timestamp(
    date: Date().addingTimeInterval(24 * 60 * 60))
  token.activeAccountID = UUID().uuidString
  token.familyID = UUID().uuidString
  return token
}

/// A gRPC server on a loopback port answering RegisterDevice, RevokeDevice and SignOut, and
/// recording each call in order. Every other method is unrouted, so it answers UNIMPLEMENTED.
private final class LoopbackBackend: Sendable {
  struct Call: Sendable {
    let method: MethodDescriptor
    let request: any SwiftProtobuf.Message
  }

  private let state = Mutex<(calls: [Call], revokeFails: Bool)>(([], false))

  var calls: [Call] { state.withLock { $0.calls } }

  func calls(to method: MethodDescriptor) -> [Call] {
    calls.filter { $0.method == method }
  }

  /// Makes RevokeDevice answer UNAVAILABLE, as a server that is down would.
  func failRevocations() {
    state.withLock { $0.revokeFails = true }
  }

  private func record(_ method: MethodDescriptor, _ request: any SwiftProtobuf.Message) -> Bool {
    state.withLock { state in
      state.calls.append(Call(method: method, request: request))
      return state.revokeFails
    }
  }

  /// Serves for as long as `body` runs, handing it the port.
  func serve<Result: Sendable>(_ body: @Sendable (Int) async throws -> Result) async throws
    -> Result
  {
    var router = RPCRouter<HTTP2ServerTransport.Posix>()
    route(&router, registerMethod) { (request: RegisterRequest, _) in
      var response = RegisterResponse()
      response.result.id = "device-\(request.input.token)"
      return response
    }
    route(&router, revokeMethod) { (_: RevokeRequest, revokeFails) in
      if revokeFails { throw RPCError(code: .unavailable, message: "down") }
      return RevokeResponse()
    }
    route(&router, signOutMethod) { (_: SignOutRequest, _) in SignOutResponse() }

    let transport = HTTP2ServerTransport.Posix(
      address: .ipv4(host: "127.0.0.1", port: 0), transportSecurity: .plaintext)
    let server = GRPCServer(transport: transport, router: router)
    return try await withThrowingTaskGroup(of: Void.self) { group in
      group.addTask { try await server.serve() }
      defer { server.beginGracefulShutdown() }
      guard let port = try await transport.listeningAddress.ipv4?.port else {
        throw RPCError(code: .internalError, message: "the loopback server bound no IPv4 port")
      }
      return try await body(port)
    }
  }

  private func route<Input: SwiftProtobuf.Message, Output: SwiftProtobuf.Message>(
    _ router: inout RPCRouter<HTTP2ServerTransport.Posix>,
    _ method: MethodDescriptor,
    _ answer: @escaping @Sendable (Input, _ revokeFails: Bool) async throws -> Output
  ) {
    router.registerHandler(
      forMethod: method,
      deserializer: ProtobufDeserializer<Input>(),
      serializer: ProtobufSerializer<Output>()
    ) { request, _ in
      let single = try await ServerRequest(stream: request)
      let revokeFails = self.record(method, single.message)
      let response = try await answer(single.message, revokeFails)
      return StreamingServerResponse(single: ServerResponse(message: response))
    }
  }
}

/// What a relaunch keeps: the login and the device registration outlive any one ClientManager.
private struct Install {
  let backend = LoopbackBackend()
  let credentials = TestCredentialStore()
  let registrations = TestDeviceRegistrationStore()
  let apnsToken = Data((0..<8).map { _ in UInt8.random(in: 0...255) })

  init() async throws {
    try await credentials.save(heldLogin())
  }

  /// Runs `body` with the backend up, handing it a way to launch the app: each call is a
  /// fresh ClientManager over the same stores.
  func run(
    _ body: @Sendable (_ launch: () throws -> ClientManager<HTTP2ClientTransport.TransportServices>)
      async throws -> Void
  ) async throws {
    let credentials = self.credentials
    let registrations = self.registrations
    try await backend.serve { port in
      try await body {
        try ClientManager<HTTP2ClientTransport.TransportServices>(
          host: "127.0.0.1", port: port, store: credentials, deviceStore: registrations)
      }
    }
  }
}

@Suite("Device registration")
struct DeviceRegistrationTests {
  @Test("a sign-out after a relaunch revokes what the earlier launch registered")
  func revokesAcrossARelaunch() async throws {
    let install = try await Install()
    let token = install.apnsToken

    try await install.run { launch in
      let firstLaunch = try launch()
      let id = try await firstLaunch.devices.register(apnsToken: token, platform: .ios)
      #expect(try await install.registrations.load()?.id == id)

      // The relaunch is handed the same token, which needs no call: it is already registered
      // under this login. Then the person signs out.
      let secondLaunch = try launch()
      #expect(try await secondLaunch.devices.register(apnsToken: token, platform: .ios) == id)
      await secondLaunch.signOut()

      #expect(install.backend.calls(to: registerMethod).count == 1)
      let revoked = install.backend.calls(to: revokeMethod)
      #expect(revoked.count == 1)
      #expect((revoked.first?.request as? RevokeRequest)?.deviceID == id)
    }
    #expect(try await install.registrations.load() == nil)
  }

  @Test("the device is revoked before the login is signed out")
  func revokesBeforeSigningOut() async throws {
    let install = try await Install()
    let token = install.apnsToken

    try await install.run { launch in
      let manager = try launch()
      _ = try await manager.devices.register(apnsToken: token, platform: .ios)
      await manager.signOut()
    }

    let order = install.backend.calls.map(\.method).filter {
      $0 == revokeMethod || $0 == signOutMethod
    }
    #expect(order == [revokeMethod, signOutMethod])
    #expect(try await install.credentials.load() == nil)
  }

  @Test("a sign-out with nothing registered revokes nothing and still signs out")
  func signsOutWithoutARegistration() async throws {
    let install = try await Install()

    try await install.run { launch in
      let manager = try launch()
      await manager.signOut()
    }

    #expect(install.backend.calls(to: revokeMethod).isEmpty)
    #expect(install.backend.calls(to: signOutMethod).count == 1)
    #expect(try await install.credentials.load() == nil)
  }

  @Test("a revocation the server could not take still signs out, and keeps the registration")
  func revocationFailureStillSignsOut() async throws {
    let install = try await Install()
    let token = install.apnsToken
    install.backend.failRevocations()

    try await install.run { launch in
      let manager = try launch()
      _ = try await manager.devices.register(apnsToken: token, platform: .ios)
      await manager.signOut()
    }

    #expect(install.backend.calls(to: signOutMethod).count == 1)
    #expect(try await install.credentials.load() == nil)
    // Still persisted, so the next sign-in on this device moves it to whoever that is.
    #expect(try await install.registrations.load() != nil)
  }
}
