//
//  SelectedMealCard.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import DesignSystem
import SwiftProtobuf
import SwiftUI

// MARK: - Selected Meal Card

struct SelectedMealCard: View {
  let meal: Mealplanning_Meal
  let scale: Float
  let onScaleChange: (Float) -> Void
  let onRemove: () -> Void
  let isRegularWidth: Bool

  @State private var scaleText: String = ""
  @FocusState private var isScaleFocused: Bool

  var body: some View {
    if isRegularWidth {
      horizontalLayout
    } else {
      verticalLayout
    }
  }

  private var horizontalLayout: some View {
    DSCard(style: .outlined) {
      HStack(alignment: .top, spacing: DSTheme.Spacing.lg) {
        VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
          Text(meal.name.isEmpty ? "Unnamed Meal" : meal.name)
            .font(DSTheme.Typography.label)
            .foregroundColor(DSTheme.Colors.textPrimary)

          if !meal.components.isEmpty {
            let recipeNames = meal.components.compactMap { component -> String? in
              component.recipe.name.isEmpty ? nil : component.recipe.name
            }
            if !recipeNames.isEmpty {
              Text(recipeNames.joined(separator: ", "))
                .font(DSTheme.Typography.caption)
                .foregroundColor(DSTheme.Colors.textSecondary)
            }
          }

          Label(
            "\(PortionsFormatter.formatScaled(min: meal.minEstimatedPortions, max: meal.hasMaxEstimatedPortions ? meal.maxEstimatedPortions : nil, scale: scale)) servings",
            systemImage: "person.2"
          )
          .font(DSTheme.Typography.caption)
          .foregroundColor(DSTheme.Colors.textSecondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)

        // Scale control
        VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
          Text("Scale")
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.textSecondary)
          HStack(spacing: DSTheme.Spacing.sm) {
            DSTextField("1.0", text: $scaleText, type: .number)
              .frame(width: 100)
              .focused($isScaleFocused)
              .onSubmit {
                updateScaleFromText()
              }
              .onChange(of: isScaleFocused) { _, isFocused in
                if !isFocused {
                  updateScaleFromText()
                }
              }
            Text("x")
              .font(DSTheme.Typography.body)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }
        }
        .frame(maxWidth: 200)

        DSIconButton("xmark.circle.fill", style: .destructive) {
          onRemove()
        }
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

  private var verticalLayout: some View {
    DSCard(style: .outlined) {
      VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
        HStack {
          VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
            Text(meal.name.isEmpty ? "Unnamed Meal" : meal.name)
              .font(DSTheme.Typography.label)
              .foregroundColor(DSTheme.Colors.textPrimary)

            if !meal.components.isEmpty {
              let recipeNames = meal.components.compactMap { component -> String? in
                component.recipe.name.isEmpty ? nil : component.recipe.name
              }
              if !recipeNames.isEmpty {
                Text(recipeNames.joined(separator: ", "))
                  .font(DSTheme.Typography.caption)
                  .foregroundColor(DSTheme.Colors.textSecondary)
              }
            }

            Label(
              "\(PortionsFormatter.formatScaled(min: meal.minEstimatedPortions, max: meal.hasMaxEstimatedPortions ? meal.maxEstimatedPortions : nil, scale: scale)) servings",
              systemImage: "person.2"
            )
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.textSecondary)
          }

          Spacer()

          DSIconButton("xmark.circle.fill", style: .destructive) {
            onRemove()
          }
        }

        // Scale control
        VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
          Text("Scale")
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.textSecondary)
          HStack(spacing: DSTheme.Spacing.sm) {
            DSTextField("1.0", text: $scaleText, type: .number)
              .focused($isScaleFocused)
              .onSubmit {
                updateScaleFromText()
              }
              .onChange(of: isScaleFocused) { _, isFocused in
                if !isFocused {
                  updateScaleFromText()
                }
              }
            Text("x")
              .font(DSTheme.Typography.body)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }
        }
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
    if let parsedValue = Float(scaleText.trimmingCharacters(in: .whitespacesAndNewlines)) {
      if parsedValue > 0 {
        onScaleChange(parsedValue)
        scaleText = String(format: "%.2f", parsedValue)
      } else {
        scaleText = String(format: "%.2f", scale)
      }
    } else {
      scaleText = String(format: "%.2f", scale)
    }
  }

}
