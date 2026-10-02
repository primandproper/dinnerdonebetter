//
//  UploadMealImageViewModel.swift
//  ios
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import PlatformClient
import SwiftUI

private let uploadChunkSize = 64 * 1024  // 64 KB
private let uploadBucketName = "meals"

@Observable
@MainActor
class UploadMealImageViewModel {
  var isUploading = false
  var errorMessage: String?
  var didSucceed = false

  private let mealID: String
  private let authManager: AuthenticationManager

  init(mealID: String, authManager: AuthenticationManager) {
    self.mealID = mealID
    self.authManager = authManager
  }

  func uploadMealImage(imageData: Data, contentType: String, objectName: String) async {
    isUploading = true
    errorMessage = nil
    didSucceed = false

    do {

      var uploadOptions = GRPCCore.CallOptions.defaults
      uploadOptions.timeout = .seconds(60)

      _ = try await authManager.authenticatedCall("uploadMealImage") { client, metadata, _ in
        try await client.mealPlanning.uploadMealImage(
          metadata: metadata,
          options: uploadOptions,
          requestProducer: { writer in
            // 1. Send metadata
            var meta = UploadedMedia_UploadMetadata()
            meta.bucket = uploadBucketName
            meta.objectName = objectName
            meta.contentType = contentType

            var uploadReq = UploadedMedia_UploadRequest()
            uploadReq.payload = .metadata(meta)

            var mediaReq = Mealplanning_UploadMealMediaRequest()
            mediaReq.mealID = self.mealID
            mediaReq.upload = uploadReq
            try await writer.write(mediaReq)

            // 2. Send chunks
            var offset = 0
            while offset < imageData.count {
              let end = min(offset + uploadChunkSize, imageData.count)
              let chunk = imageData.subdata(in: offset..<end)
              offset = end

              var chunkUploadReq = UploadedMedia_UploadRequest()
              chunkUploadReq.payload = .chunk(chunk)

              var chunkMediaReq = Mealplanning_UploadMealMediaRequest()
              chunkMediaReq.mealID = self.mealID
              chunkMediaReq.upload = chunkUploadReq
              try await writer.write(chunkMediaReq)
            }
          }
        )
      }

      didSucceed = true
    } catch {
      if let error = error.platformError {
        let statusMessage = UploadErrorFormatter.formatRPCError(error)
        errorMessage = "Upload failed: \(statusMessage)"
        print("❌ Meal image upload RPC error: \(error.code), \(error.serverMessage)")
      } else {
        errorMessage = "Upload failed: \(error.localizedDescription)"
      }
    }

    isUploading = false
  }

}
