//
//  DeviceNameInterceptorTests.swift
//  iosTests
//

import Foundation
import GRPCCore
@testable import ios
import Testing

struct DeviceNameInterceptorTests {
    @Test("This device is described by its model and its system")
    func describesThisDevice() {
        let interceptor = DeviceNameInterceptor.describingThisDevice()

        #expect(!interceptor.deviceName.isEmpty)
        #expect(interceptor.deviceName.contains("iOS"))
    }

    @Test("Every call names the device")
    func namesTheDevice() async throws {
        let name = "test device \(UUID().uuidString)"
        let interceptor = DeviceNameInterceptor(deviceName: name)

        let response: StreamingClientResponse<String> = try await interceptor.intercept(
            request: StreamingClientRequest<String>(of: String.self, producer: { _ in }),
            context: ClientContext(
                descriptor: MethodDescriptor(fullyQualifiedService: "test.Service", method: "Call"),
                remotePeer: "in-process",
                localPeer: "in-process"
            )
        ) { request, _ in
            #expect(Array(request.metadata[stringValues: deviceNameMetadataKey]) == [name])
            return StreamingClientResponse(of: String.self, error: RPCError(code: .unavailable, message: "not sent"))
        }

        #expect(response.accepted.isFailure)
    }
}

private extension Result {
    var isFailure: Bool {
        if case .failure = self { return true }
        return false
    }
}
