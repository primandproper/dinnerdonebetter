//
//  AssociatedRecipeCard.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import DesignSystem
import SwiftProtobuf
import SwiftUI

struct AssociatedRecipeCard: View {
  let recipe: Mealplanning_Recipe
  var viewModel: PerformRecipeViewModel?
  var parentRecipe: Mealplanning_Recipe?
  @Environment(AuthenticationManager.self) private var authManager

  private var canMarkAsMade: Bool {
    guard viewModel != nil, let parentRecipe else { return false }
    return parentRecipe.associatedRecipes.contains { $0.id == recipe.id }
  }

  private var allStepsComplete: Bool {
    guard let viewModel else { return false }
    return recipe.steps.allSatisfy { viewModel.isStepCompleted(recipeID: recipe.id, stepID: $0.id) }
  }

  var body: some View {
    DSCard {
      HStack {
        VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
          Text(recipe.name.isEmpty ? "Unnamed Recipe" : recipe.name)
            .font(DSTheme.Typography.body)
            .fontWeight(.medium)
            .foregroundColor(DSTheme.Colors.textPrimary)

          if !recipe.description_p.isEmpty {
            Text(recipe.description_p)
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.textSecondary)
              .lineLimit(2)
          }
        }

        Spacer()

        if canMarkAsMade, let viewModel {
          Button(
            action: {
              UIImpactFeedbackGenerator(style: .light).impactOccurred()
              viewModel.markAssociatedRecipeAsCompleted(associatedRecipe: recipe)
            },
            label: {
              HStack(spacing: DSTheme.Spacing.xs) {
                Image(systemName: allStepsComplete ? "checkmark.circle.fill" : "checkmark.circle")
                  .foregroundColor(allStepsComplete ? .green : DSTheme.Colors.primary)
                Text("I made this")
                  .font(DSTheme.Typography.label)
                  .foregroundColor(allStepsComplete ? .green : DSTheme.Colors.primary)
              }
            }
          )
          .buttonStyle(.plain)
          .disabled(allStepsComplete)
        }

        NavigationLink(
          destination: {
            PerformRecipeView(recipeID: recipe.id)
              .environment(authManager)
          },
          label: {
            Image(systemName: "chevron.right")
              .font(.caption)
              .foregroundColor(DSTheme.Colors.textTertiary)
          }
        )
      }
    }
  }
}
