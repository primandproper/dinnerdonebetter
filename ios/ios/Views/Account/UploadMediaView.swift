//
//  UploadMediaView.swift
//  ios
//

import PhotosUI
import SwiftUI

/// Generic media upload view: uploads an image to the media registry as the caller's own object.
struct UploadMediaView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @State private var viewModel: MediaUploadViewModel?
  @State private var selectedItem: PhotosPickerItem?

  var body: some View {
    ScrollView {
      VStack(spacing: DSTheme.Spacing.xl) {
        if let viewModel = viewModel {
          DSSection("Upload Photo", subtitle: "Select a photo to upload to the backend") {
            VStack(spacing: DSTheme.Spacing.lg) {
              PhotosPicker(
                selection: $selectedItem,
                matching: .images,
                photoLibrary: .shared()
              ) {
                HStack(spacing: DSTheme.Spacing.md) {
                  Image(systemName: "photo.on.rectangle.angled")
                    .font(.system(size: DSTheme.IconSize.lg))
                    .foregroundColor(DSTheme.Colors.primary)
                  Text("Select Photo")
                    .font(DSTheme.Typography.label)
                    .foregroundColor(DSTheme.Colors.textPrimary)
                }
                .frame(maxWidth: .infinity)
                .padding(DSTheme.Spacing.lg)
                .background(DSTheme.Colors.cardBackground)
                .cornerRadius(DSTheme.Radius.md)
              }
              .disabled(viewModel.isUploading)

              if viewModel.isUploading {
                ProgressView("Uploading...")
                  .padding()
              }

              if let error = viewModel.errorMessage {
                Text(error)
                  .font(DSTheme.Typography.caption)
                  .foregroundColor(DSTheme.Colors.error)
                  .multilineTextAlignment(.leading)
              }

              if let path = viewModel.lastUploadedStoragePath {
                VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
                  Text("Uploaded successfully")
                    .font(DSTheme.Typography.label)
                    .foregroundColor(DSTheme.Colors.success)
                  Text(path)
                    .font(DSTheme.Typography.caption)
                    .foregroundColor(DSTheme.Colors.textSecondary)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(DSTheme.Spacing.md)
                .background(DSTheme.Colors.cardBackground)
                .cornerRadius(DSTheme.Radius.md)
              }
            }
          }
        } else {
          DSInitializingView()
        }
      }
      .dsScreenPadding()
    }
    .navigationTitle("Upload Media")
    .onAppear {
      if viewModel == nil {
        viewModel = MediaUploadViewModel(authManager: authManager)
      }
    }
    .onChange(of: selectedItem) { _, newItem in
      guard let item = newItem, let viewModel = viewModel else { return }
      Task {
        await loadAndUpload(item: item, viewModel: viewModel)
      }
    }
  }

  private func loadAndUpload(item: PhotosPickerItem, viewModel: MediaUploadViewModel) async {
    do {
      guard let data = try await item.loadTransferable(type: Data.self) else {
        viewModel.errorMessage = "Failed to load image data"
        return
      }

      let prepared = ImagePreparationHelper.prepareImageForUpload(data: data)
      let objectName = UUID().uuidString + "." + prepared.fileExtension

      await viewModel.upload(
        imageData: prepared.data,
        contentType: prepared.contentType,
        objectName: objectName
      )

      selectedItem = nil
    } catch {
      viewModel.errorMessage = "Failed to load image: \(error.localizedDescription)"
    }
  }
}

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "Test User"
  authManager.userID = "user123"

  return NavigationStack {
    UploadMediaView()
      .environment(authManager)
  }
}
