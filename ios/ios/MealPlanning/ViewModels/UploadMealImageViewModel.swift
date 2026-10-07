//
//  UploadMealImageViewModel.swift
//  ios
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import PlatformClient
import SwiftUI

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

      let parts = MediaUploadParts.make(name: objectName, contentType: contentType, data: imageData)

      _ = try await authManager.authenticatedCall("uploadMealImage") { client, metadata, _ in
        try await client.mealPlanning.uploadMealImage(
          metadata: metadata,
          options: uploadOptions,
          requestProducer: { writer in
            // The meal planning RPC carries the registry's upload stream — a header, then
            // chunks — and names what it is attached to itself.
            for part in parts {
              var request = Mealplanning_UploadMealMediaRequest()
              request.mealID = self.mealID
              request.upload = part
              try await writer.write(request)
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
