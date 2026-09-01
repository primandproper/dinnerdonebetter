//
//  LoadedMealPlanView.swift
//  ios
//
//  Fetches a whole meal plan by ID and hands it to a screen that needs one.
//

import SwiftProtobuf
import SwiftUI

/// Loads a whole meal plan by ID, then renders `content` with it.
///
/// `GetMealPlansForAccount` answers with `Mealplanning_MealPlanSummary`, whose events
/// carry no options -- an option embeds a whole meal, whose components embed whole
/// recipes, and a page of those does not fit in a gRPC message. Screens that read a
/// plan's options (voting, the detail screen's meal cards) therefore fetch the plan
/// itself, which is what `GetMealPlan` is for.
///
/// Screens that only need the plan's own columns and its event dates take a
/// `MealPlanDisplayable` instead and need none of this.
struct LoadedMealPlanView<Content: View>: View {
  @Environment(AuthenticationManager.self) private var authManager

  let mealPlanID: String
  @ViewBuilder var content: (Mealplanning_MealPlan) -> Content

  @State private var mealPlan: Mealplanning_MealPlan?
  @State private var loadError: String?
  @State private var isLoading = true

  var body: some View {
    Group {
      if let mealPlan {
        content(mealPlan)
      } else if isLoading {
        DSInitializingView()
      } else {
        DSErrorView(
          loadError ?? "Please try again.",
          title: "Couldn't load meal plan",
          onRetry: { await load() }
        )
      }
    }
    .task {
      await load()
    }
  }

  private func load() async {
    isLoading = true
    loadError = nil

    do {
      var request = Mealplanning_GetMealPlanRequest()
      request.mealPlanID = mealPlanID

      let response = try await authManager.authenticatedCall("getMealPlan", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getMealPlan(request, metadata: metadata, options: options)
      }

      mealPlan = response.result
    } catch {
      loadError = ErrorDisplayFormatter.format(error, context: "load meal plan").message
    }

    isLoading = false
  }
}
