//
//  UploadRecipeImageViewModel.swift
//  ios
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import PlatformClient
import SwiftUI

private let uploadChunkSize = 64 * 1024  // 64 KB
private let uploadBucketName = "recipes"

@Observable
@MainActor
class UploadRecipeImageViewModel {
  var isUploading = false
  var errorMessage: String?
  var didSucceed = false

  private let recipeID: String
  private let authManager: AuthenticationManager

  init(recipeID: String, authManager: AuthenticationManager) {
    self.recipeID = recipeID
    self.authManager = authManager
  }

  func uploadRecipeImage(imageData: Data, contentType: String, objectName: String) async {
    isUploading = true
    errorMessage = nil
    didSucceed = false

    do {

      var uploadOptions = GRPCCore.CallOptions.defaults
      uploadOptions.timeout = .seconds(60)

      _ = try await authManager.authenticatedCall("uploadRecipeImage") {
        client, metadata, _ in
        try await client.mealPlanning.uploadRecipeImage(
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

            var mediaReq = Mealplanning_UploadRecipeMediaRequest()
            mediaReq.recipeID = self.recipeID
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

              var chunkMediaReq = Mealplanning_UploadRecipeMediaRequest()
              chunkMediaReq.recipeID = self.recipeID
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
        print("❌ Recipe image upload RPC error: \(error.code), \(error.serverMessage)")
      } else {
        errorMessage = "Upload failed: \(error.localizedDescription)"
      }
    }

    isUploading = false
  }

}
