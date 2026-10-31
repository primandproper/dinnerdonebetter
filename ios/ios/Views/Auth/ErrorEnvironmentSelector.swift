//
//  ErrorEnvironmentSelector.swift
//  ios
//
//  DEBUG environment picker shown beneath a server-down error
//

import DesignSystem
import SwiftUI

/// A compact button that opens the environment picker, so a debug build that can't reach its
/// server can switch to another one from the error screen. Passed to `DSContentState` as its
/// `errorAccessory`, since the design system knows nothing about `AppEnvironment`.
///
/// Usage:
/// ```swift
/// DSContentState(
///   isLoading: viewModel.isLoading,
///   error: viewModel.errorMessage,
///   errorAccessory: ErrorEnvironmentSelector.accessory(if: viewModel.isServerDownError)
/// ) { ... }
/// ```
struct ErrorEnvironmentSelector: View {
  #if DEBUG
    @State private var showEnvironmentPicker = false
    @State private var selectedEnvironment: AppEnvironment = APIConfiguration.currentEnvironment
  #endif

  /// The selector when `show` is true in a DEBUG build; otherwise no accessory at all.
  static func accessory(if show: Bool) -> AnyView? {
    #if DEBUG
      show ? AnyView(ErrorEnvironmentSelector()) : nil
    #else
      nil
    #endif
  }

  var body: some View {
    #if DEBUG
      Button {
        showEnvironmentPicker = true
      } label: {
        HStack(spacing: DSTheme.Spacing.xs) {
          Image(systemName: selectedEnvironment.iconName)
            .font(.system(size: 12))
          Text(selectedEnvironment.displayName)
            .font(DSTheme.Typography.caption)
          Image(systemName: "chevron.up")
            .font(.system(size: 10, weight: .semibold))
        }
        .foregroundColor(DSTheme.Colors.textSecondary)
        .padding(.horizontal, DSTheme.Spacing.md)
        .padding(.vertical, DSTheme.Spacing.sm)
        .background(DSTheme.Colors.cardBackground)
        .cornerRadius(DSTheme.Radius.full)
      }
      .padding(.top, DSTheme.Spacing.sm)
      .sheet(isPresented: $showEnvironmentPicker) {
        EnvironmentPickerSheet(
          selectedEnvironment: $selectedEnvironment,
          onDismiss: { showEnvironmentPicker = false }
        )
        .presentationDetents([.medium])
      }
    #else
      EmptyView()
    #endif
  }
}
