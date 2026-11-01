//
//  DSKeepScreenAwakeButton.swift
//  ios
//
//  Design System Keep Screen Awake Button - Toggles screen dimming while cooking
//

import DesignSystem
import SwiftUI
import UIKit

/// A button that toggles keeping the device screen awake.
/// When enabled, prevents the screen from dimming or locking while following a recipe.
///
/// Usage:
/// ```swift
/// DSKeepScreenAwakeButton(inline: true)  // Full-width card above recipe steps
/// DSKeepScreenAwakeButton()              // Compact icon for toolbar
/// ```
struct DSKeepScreenAwakeButton: View {
  @State private var isScreenAwake = false
  var inline: Bool = false
  /// Optional callback when the user toggles the keep-awake state (for analytics).
  var onToggle: ((Bool) -> Void)?

  var body: some View {
    Button {
      isScreenAwake.toggle()
      UIApplication.shared.isIdleTimerDisabled = isScreenAwake
      onToggle?(isScreenAwake)
    } label: {
      if inline {
        HStack(spacing: DSTheme.Spacing.md) {
          Image(systemName: isScreenAwake ? "sun.max.fill" : "sun.max")
            .font(DSTheme.Typography.title3)
            .foregroundColor(isScreenAwake ? DSTheme.Colors.primary : DSTheme.Colors.textSecondary)
          Text(isScreenAwake ? "Screen awake (tap to allow sleep)" : "Keep screen awake")
            .font(DSTheme.Typography.body)
            .foregroundColor(isScreenAwake ? DSTheme.Colors.primary : DSTheme.Colors.textPrimary)
          Spacer()
        }
        .padding(DSTheme.Spacing.md)
        .background(
          isScreenAwake ? DSTheme.Colors.primary.opacity(0.1) : DSTheme.Colors.cardBackground
        )
        .cornerRadius(DSTheme.Radius.md)
      } else {
        Image(systemName: isScreenAwake ? "sun.max.fill" : "sun.max")
          .foregroundColor(isScreenAwake ? DSTheme.Colors.primary : DSTheme.Colors.textSecondary)
      }
    }
    .buttonStyle(.plain)
    .accessibilityLabel(isScreenAwake ? "Screen awake (tap to allow sleep)" : "Keep screen awake")
    .accessibilityHint("Prevents the screen from dimming while cooking")
    .onDisappear {
      UIApplication.shared.isIdleTimerDisabled = false
    }
  }
}

// MARK: - Preview

#Preview("DSKeepScreenAwakeButton") {
  VStack(spacing: DSTheme.Spacing.lg) {
    DSKeepScreenAwakeButton(inline: true)
    DSKeepScreenAwakeButton(inline: false)
  }
  .padding()
}
