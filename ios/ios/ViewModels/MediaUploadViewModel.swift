//
//  MediaUploadViewModel.swift
//  ios
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import PlatformClient
import SwiftUI

private let uploadChunkSize = 64 * 1024  // 64 KB

/// The subject type the media registry lets a caller attach an upload to themselves with. It is
/// the one attachment the registry authorizes on its own, and it is what an avatar is.
let mediaRegistryUserSubjectType = "user"

/// The messages of one media registry upload: its header, then its bytes in chunks.
///
/// Platform's upload stream is the same for the registry's own UploadObject and for the meal
/// planning upload RPCs, which carry it inside their requests, so every upload in the app builds
/// its stream here.
enum MediaUploadParts {
  static func make(
    name: String,
    contentType: String,
    data: Data,
    belongsTo: Primandproper_Platform_Mediaregistry_V1_Subject? = nil
  ) -> [Primandproper_Platform_Mediaregistry_V1_UploadObjectRequest] {
    var header = Primandproper_Platform_Mediaregistry_V1_UploadObjectHeader()
    header.name = name
    header.contentType = contentType
    if let belongsTo {
      header.belongsTo = belongsTo
    }

    var first = Primandproper_Platform_Mediaregistry_V1_UploadObjectRequest()
    first.part = .header(header)

    var parts = [first]

    var offset = 0
    while offset < data.count {
      let end = min(offset + uploadChunkSize, data.count)

      var chunk = Primandproper_Platform_Mediaregistry_V1_UploadObjectRequest()
      chunk.part = .chunk(data.subdata(in: offset..<end))
      parts.append(chunk)

      offset = end
    }

    return parts
  }
}

/// Generic media upload view model: uploads an image to the media registry as the caller's own
/// object. Where in the bucket it lands is the server's layout, not the client's choice.
@Observable
@MainActor
public class MediaUploadViewModel {
  public var isUploading = false
  public var errorMessage: String?
  public var lastUploadedStoragePath: String?

  private let authManager: AuthenticationManager

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  /// Uploads media data as the caller's own object.
  /// - Parameters:
  ///   - imageData: Raw file data (e.g. from PhotosPickerItem or camera)
  ///   - contentType: MIME type (e.g. "image/jpeg", "image/png")
  ///   - objectName: The object's name, one path segment (e.g. "uuid.jpg")
  public func upload(
    imageData: Data,
    contentType: String,
    objectName: String
  ) async {
    isUploading = true
    errorMessage = nil
    lastUploadedStoragePath = nil

    do {

      var uploadOptions = GRPCCore.CallOptions.defaults
      uploadOptions.timeout = .seconds(60)

      let parts = MediaUploadParts.make(name: objectName, contentType: contentType, data: imageData)

      let response = try await authManager.authenticatedCall("uploadObject") {
        client, metadata, _ in
        try await client.mediaRegistry.uploadObject(
          metadata: metadata,
          options: uploadOptions,
          requestProducer: { writer in
            for part in parts {
              try await writer.write(part)
            }
          }
        )
      }

      lastUploadedStoragePath = response.result.key
    } catch {
      if let error = error.platformError {
        let statusMessage = UploadErrorFormatter.formatRPCError(error)
        errorMessage = "Upload failed: \(statusMessage)"
        print("❌ Media upload RPC error: \(error.code), \(error.serverMessage)")
      } else {
        errorMessage = "Upload failed: \(error.localizedDescription)"
      }
    }

    isUploading = false
  }

}

/// Shared RPC error formatting for upload flows.
enum UploadErrorFormatter {
  static let notSignedIn = "Session expired. Please sign in again."

  static func formatRPCError(_ error: PlatformError) -> String {
    switch error.code {
    case .cancelled:
      return
        "Request was cancelled. This can happen if the connection was interrupted or the request took too long."
    case .deadlineExceeded:
      return "Request timed out. Try a smaller image."
    case .unauthenticated:
      return "Session expired. Please sign in again."
    case .unavailable:
      return
        "Server unavailable. Is the backend running at \(APIConfiguration.grpcHost):\(APIConfiguration.grpcPort)?"
    case .permissionDenied:
      return "Permission denied."
    case .invalidArgument:
      return "Invalid request: \(error.serverMessage)"
    default:
      return "\(error.code): \(error.serverMessage)"
    }
  }
}
