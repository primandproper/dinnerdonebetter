//
//  UploadAvatarViewModel.swift
//  ios
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import PlatformClient
import SwiftUI

/// Uploading a user's avatar.
///
/// An avatar is an upload to the media registry attached to its uploader — belongs_to
/// {user, their ID} — which is the one attachment the registry authorizes without asking
/// anybody. The newest object attached to a user is their avatar, and ListObjectsBySubject is
/// the read that hands it back; platform's User carries no avatar field.
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

    let userID = authManager.userID
    guard !userID.isEmpty else {
      errorMessage = "Upload failed: \(UploadErrorFormatter.notSignedIn)"
      isUploading = false
      return
    }

    do {

      var uploadOptions = GRPCCore.CallOptions.defaults
      uploadOptions.timeout = .seconds(60)

      var subject = Primandproper_Platform_Mediaregistry_V1_Subject()
      subject.type = mediaRegistryUserSubjectType
      subject.id = userID

      let parts = MediaUploadParts.make(
        name: objectName, contentType: contentType, data: imageData, belongsTo: subject)

      _ = try await authManager.authenticatedCall("uploadAvatar") { client, metadata, _ in
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

      didSucceed = true
    } catch {
      if let error = error.platformError {
        let statusMessage = UploadErrorFormatter.formatRPCError(error)
        errorMessage = "Upload failed: \(statusMessage)"
        print("❌ Avatar upload RPC error: \(error.code), \(error.serverMessage)")
      } else {
        errorMessage = "Upload failed: \(error.localizedDescription)"
      }
    }

    isUploading = false
  }
}
