//
//  PerformRecipeView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import DesignSystem
import SwiftProtobuf
import SwiftUI

struct PerformRecipeView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @State private var viewModel: PerformRecipeViewModel?
  @State private var isInstrumentsVesselsExpanded = false
  @State private var isIngredientsExpanded = false
  @State private var checkedIngredients: Set<String> = []
  // Track by ValidInstrument/ValidVessel ID
  @State private var checkedInstrumentsVessels: Set<String> = []
  @State private var scaleFromMealPlan: Float?

  let recipeID: String
  let highlightedStepIDs: Set<String>?
  let prepTaskContext: PrepTaskContext?
  /// Scale from meal plan (e.g. 0.5 for half recipe). nil = use default 1.0.
  let initialScale: Float?

  struct PrepTaskContext {
    let prepTaskName: String?
    let recipeName: String?
    let eventName: String?
    let eventTime: Date?
  }

  init(
    recipeID: String, highlightedStepIDs: Set<String>? = nil,
    prepTaskContext: PrepTaskContext? = nil,
    initialScale: Float? = nil
  ) {
    self.recipeID = recipeID
    self.highlightedStepIDs = highlightedStepIDs
    self.prepTaskContext = prepTaskContext
    self.initialScale = initialScale
  }

  var body: some View {
    NavigationStack {
      Group {
        if let viewModel = viewModel {
          DSContentState(
            isLoading: viewModel.isLoading,
            loadingMessage: "Loading recipe...",
            error: viewModel.errorMessage,
            onRetry: {
              eventReporterService.reporter.track(event: "perform_recipe_retry", properties: [:])
              await viewModel.loadRecipe()
            },
            content: {
              if let recipe = viewModel.recipe {
                RecipePerformanceContentView(
                  checkedIngredients: $checkedIngredients,
                  checkedInstrumentsVessels: $checkedInstrumentsVessels,
                  isInstrumentsVesselsExpanded: $isInstrumentsVesselsExpanded,
                  isIngredientsExpanded: $isIngredientsExpanded,
                  recipe: recipe,
                  viewModel: viewModel,
                  highlightedStepIDs: highlightedStepIDs,
                  prepTaskContext: prepTaskContext,
                  externalScale: externalScaleBinding
                )
                .navigationTitle(navigationTitle)
                .navigationBarTitleDisplayMode(.inline)
              } else {
                DSLoadingView("Loading...")
              }
            })
        } else {
          DSInitializingView()
        }
      }
      .onAppear {
        if viewModel == nil {
          viewModel = PerformRecipeViewModel(recipeID: recipeID, authManager: authManager)
          eventReporterService.reporter.track(
            event: "perform_recipe_started",
            properties: [
              "recipe_id": recipeID,
              "from_prep_task": prepTaskContext != nil,
            ])
          Task {
            await viewModel?.loadRecipe()
          }
        }
      }
    }
  }

  private var externalScaleBinding: Binding<Float?> {
    if initialScale != nil {
      return Binding(
        get: { scaleFromMealPlan ?? initialScale },
        set: { scaleFromMealPlan = $0 }
      )
    }
    return .constant(nil)
  }

  private var navigationTitle: String {
    if let context = prepTaskContext {
      var parts: [String] = []

      if let prepTaskName = context.prepTaskName, !prepTaskName.isEmpty {
        parts.append(prepTaskName)
      } else {
        parts.append("Prep Task")
      }

      if let recipeName = context.recipeName, !recipeName.isEmpty {
        parts.append("for \(recipeName)")
      }

      if context.eventName != nil, let eventTime = context.eventTime {
        let formatter = DateFormatter()
        formatter.dateStyle = .none
        formatter.timeStyle = .short
        parts.append("at \(formatter.string(from: eventTime))")
      }

      return parts.joined(separator: " ")
    }

    return "Perform Recipe"
  }
}

// MARK: - Preview

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "Test User"
  authManager.userID = "user123"
  authManager.accountID = "account123"

  return PerformRecipeView(recipeID: "test-recipe")
    .environment(authManager)
    .environment(EventReporterService())
}
