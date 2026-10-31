//
//  RecipeScaleInput.swift
//  ios
//
//  Scale input with TextField, +/- buttons, and optional portions display.
//

import DesignSystem
import SwiftUI

struct RecipeScaleInput: View {
  @Binding var scale: Float
  /// When provided, shows scaled yield (e.g. "~4 servings").
  var estimatedPortionsMin: Float?
  var estimatedPortionsMax: Float?
  @State private var scaleText: String = "1.0"
  @FocusState private var isScaleFocused: Bool

  private let range: ClosedRange<Float> = 0.25...4.0
  private let step: Float = 0.25

  init(
    scale: Binding<Float>, estimatedPortionsMin: Float? = nil, estimatedPortionsMax: Float? = nil
  ) {
    _scale = scale
    self.estimatedPortionsMin = estimatedPortionsMin
    self.estimatedPortionsMax = estimatedPortionsMax
  }

  var body: some View {
    HStack(spacing: DSTheme.Spacing.md) {
      Text("Scale:")
        .font(DSTheme.Typography.label)
        .foregroundColor(DSTheme.Colors.textPrimary)

      HStack(spacing: DSTheme.Spacing.sm) {
        TextField("1.0", text: $scaleText)
          .keyboardType(.decimalPad)
          .textFieldStyle(.roundedBorder)
          .font(DSTheme.Typography.body)
          .frame(width: 80)
          .focused($isScaleFocused)
          .onSubmit {
            updateScaleFromText()
          }
          .onChange(of: isScaleFocused) { _, isFocused in
            if !isFocused {
              updateScaleFromText()
            }
          }
          .onChange(of: scaleText) { _, newValue in
            var filtered = newValue.filter { $0.isNumber || $0 == "." }
            let parts = filtered.split(separator: ".", omittingEmptySubsequences: false)
            if parts.count > 2 {
              filtered = String(parts[0]) + "." + parts.dropFirst().joined()
            }
            if filtered != newValue {
              scaleText = filtered
            }
          }

        Text("x")
          .font(DSTheme.Typography.body)
          .foregroundColor(DSTheme.Colors.textSecondary)

        DSStepperButtons(
          onDecrement: { adjustScale(by: -step) }, onIncrement: { adjustScale(by: step) })
      }

      if let portionsMin = estimatedPortionsMin, portionsMin > 0 {
        Text(
          "(~\(PortionsFormatter.formatScaled(min: portionsMin, max: estimatedPortionsMax, scale: scale)) servings)"
        )
        .font(DSTheme.Typography.caption)
        .foregroundColor(DSTheme.Colors.textSecondary)
      }
    }
    .onAppear {
      scaleText = String(format: "%.2f", scale)
    }
    .onChange(of: scale) { _, newValue in
      if !isScaleFocused {
        scaleText = String(format: "%.2f", newValue)
      }
    }
  }

  private func updateScaleFromText() {
    if let parsed = Float(scaleText), parsed > 0 {
      let clamped = min(max(parsed, range.lowerBound), range.upperBound)
      scale = clamped
      scaleText = String(format: "%.2f", clamped)
    } else {
      scaleText = String(format: "%.2f", scale)
    }
  }

  private func adjustScale(by delta: Float) {
    let next = min(max(scale + delta, range.lowerBound), range.upperBound)
    scale = next
    scaleText = String(format: "%.2f", next)
  }
}
