//
//  UploadAvatarView.swift
//  ios
//

import DesignSystem
import PhotosUI
import SwiftUI
import UIKit

struct UploadAvatarView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @State private var viewModel: UploadAvatarViewModel?
  @State private var selectedItem: PhotosPickerItem?
  @Environment(\.dismiss) private var dismiss

  var body: some View {
    ScrollView {
      VStack(spacing: DSTheme.Spacing.xl) {
        if let viewModel = viewModel {
          DSSection("Profile Photo", subtitle: "Select a photo to set as your profile picture") {
            VStack(spacing: DSTheme.Spacing.lg) {
              PhotosPicker(
                selection: $selectedItem,
                matching: .images,
                photoLibrary: .shared()
              ) {
                HStack(spacing: DSTheme.Spacing.md) {
                  Image(systemName: "person.crop.circle.badge.plus")
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

              if viewModel.didSucceed {
                HStack(spacing: DSTheme.Spacing.sm) {
                  Image(systemName: "checkmark.circle.fill")
                    .foregroundColor(DSTheme.Colors.success)
                  Text("Profile photo updated successfully")
                    .font(DSTheme.Typography.label)
                    .foregroundColor(DSTheme.Colors.success)
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
    .navigationTitle("Profile Photo")
    .onAppear {
      if viewModel == nil {
        viewModel = UploadAvatarViewModel(authManager: authManager)
      }
    }
    .onChange(of: selectedItem) { _, newItem in
      guard let item = newItem, let viewModel = viewModel else { return }
      Task {
        await loadAndUpload(item: item, viewModel: viewModel)
      }
    }
    .onChange(of: viewModel?.didSucceed) { _, didSucceed in
      if didSucceed == true {
        dismiss()
      }
    }
  }

  private func loadAndUpload(item: PhotosPickerItem, viewModel: UploadAvatarViewModel) async {
    do {
      guard let data = try await item.loadTransferable(type: Data.self) else {
        viewModel.errorMessage = "Failed to load image data"
        return
      }

      let prepared = ImagePreparationHelper.prepareImageForUpload(data: data)
      let objectName = UUID().uuidString + "." + prepared.fileExtension

      await viewModel.uploadAvatar(
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
    UploadAvatarView()
      .environment(authManager)
  }
}
