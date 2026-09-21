//
//  UploadAvatarViewModel.swift
//  ios
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import SwiftUI

private let uploadChunkSize = 64 * 1024  // 64 KB
private let uploadBucketName = "avatars"

/// Uploading a user's avatar.
///
/// It goes to the media service now, where every other image already went. The RPC it used
/// to call was on the identity service, and platform's directory has no method for one —
/// correctly, since an avatar is a row in the upload registry and a reference to it rather
/// than a field of a user.
///
/// What is missing is the other half: nothing yet records that this object is *this user's*
/// avatar, so the image lands in the bucket and no read hands it back. Reading one is a
/// media-surface RPC that does not exist yet, which is why the avatar comes back as
/// initials everywhere it is rendered.
@Observable
@MainActor
class UploadAvatarViewModel {
  var isUploading = false
  var errorMessage: String?
  var didSucceed = false

  private let authManager: AuthenticationManager

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  func uploadAvatar(imageData: Data, contentType: String, objectName: String) async {
    isUploading = true
    errorMessage = nil
    didSucceed = false

    do {
      let (clientManager, metadata) = try await getClientManagerAndMetadata()

      var uploadOptions = GRPCCore.CallOptions.defaults
      uploadOptions.timeout = .seconds(60)

      _ = try await clientManager.client.uploadedMedia.upload(
        metadata: metadata,
        options: uploadOptions,
        requestProducer: { writer in
          // 1. Send metadata
          var meta = UploadedMedia_UploadMetadata()
          meta.bucket = uploadBucketName
          meta.objectName = objectName
          meta.contentType = contentType

          var metadataReq = UploadedMedia_UploadRequest()
          metadataReq.payload = .metadata(meta)
          try await writer.write(metadataReq)

          // 2. Send chunks
          var offset = 0
          while offset < imageData.count {
            let end = min(offset + uploadChunkSize, imageData.count)
            let chunk = imageData.subdata(in: offset..<end)
            offset = end

            var chunkReq = UploadedMedia_UploadRequest()
            chunkReq.payload = .chunk(chunk)
            try await writer.write(chunkReq)
          }
        }
      )

      didSucceed = true
    } catch let error as GRPCCore.RPCError {
      let statusMessage = formatRPCError(error)
      errorMessage = "Upload failed: \(statusMessage)"
      print("❌ Avatar upload RPC error: \(error.code), \(error.message)")
    } catch {
      errorMessage = "Upload failed: \(error.localizedDescription)"
    }

    isUploading = false
  }

  private func getClientManagerAndMetadata() async throws -> (
    ClientManager<HTTP2ClientTransport.TransportServices>, GRPCCore.Metadata
  ) {
    guard let clientManager = try? authManager.getClientManager() else {
      throw NSError(
        domain: "UploadAvatarViewModel", code: 1,
        userInfo: [NSLocalizedDescriptionKey: "Failed to get client manager"])
    }

    guard let oauth2Token = await authManager.getOAuth2AccessToken() else {
      throw NSError(
        domain: "UploadAvatarViewModel", code: 2,
        userInfo: [NSLocalizedDescriptionKey: "Failed to get OAuth2 access token"])
    }

    let metadata = clientManager.authenticatedMetadata(accessToken: oauth2Token)
    return (clientManager, metadata)
  }

  private func formatRPCError(_ error: GRPCCore.RPCError) -> String {
    UploadErrorFormatter.formatRPCError(error)
  }
}
