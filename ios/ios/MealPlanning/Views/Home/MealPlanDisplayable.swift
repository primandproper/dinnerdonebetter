//
//  MealPlanDisplayable.swift
//  ios
//
//  What the meal-plan display helpers read off a plan, so they can read it off either
//  shape the API hands back.
//

import Foundation
import SwiftProtobuf

/// A meal plan as the cards, headers and date ranges see it.
///
/// `GetMealPlansForAccount` answers with `Mealplanning_MealPlanSummary`, whose events
/// carry no options: a plan's options embed whole meals, which embed whole recipes, and a
/// page of those does not fit in a gRPC message. `GetMealPlan` answers with the whole
/// `Mealplanning_MealPlan`. Both conform here, so a card on the home screen and a header
/// on a detail screen format identically without the home screen fetching whole plans.
///
/// Anything a summary genuinely cannot answer is projected onto it server-side rather
/// than reached for through this protocol -- see `chosenMealDisplayName`.
protocol MealPlanDisplayable {
  var id: String { get }
  var notes: String { get }
  var status: Mealplanning_MealPlanStatus { get }
  var votingDeadline: SwiftProtobuf.Google_Protobuf_Timestamp { get }
  var groceryListInitialized: Bool { get }
  var tasksCreated: Bool { get }

  /// The plan's events, in whichever shape this plan carries them.
  var displayEvents: [MealPlanEventDisplayable] { get }
}

/// A meal plan event as the display helpers see it: its own columns, plus the one thing
/// they read out of its options.
protocol MealPlanEventDisplayable {
  var id: String { get }
  var startsAt: SwiftProtobuf.Google_Protobuf_Timestamp { get }
  var endsAt: SwiftProtobuf.Google_Protobuf_Timestamp { get }
  var mealName: Mealplanning_MealPlanEventName { get }

  /// The name to show for the meal voting settled on, or nil while the event is still
  /// awaiting a decision.
  ///
  /// A whole plan derives this from the chosen option's meal; a summary reads the
  /// `chosen_meal_name` the server projected onto it, because the options it would
  /// otherwise derive it from are what the summary exists to drop.
  var chosenMealDisplayName: String? { get }
}

// MARK: - The whole meal plan

extension Mealplanning_MealPlan: MealPlanDisplayable {
  var displayEvents: [MealPlanEventDisplayable] { events }
}

extension Mealplanning_MealPlanEvent: MealPlanEventDisplayable {
  var chosenMealDisplayName: String? {
    guard let chosen = options.first(where: { $0.chosen }) else { return nil }

    let meal = chosen.meal
    if !meal.name.isEmpty {
      return meal.name
    }

    let recipeNames = meal.components.compactMap { component -> String? in
      component.recipe.name.isEmpty ? nil : component.recipe.name
    }
    return recipeNames.isEmpty ? nil : recipeNames.joined(separator: ", ")
  }
}

// MARK: - The summary the list endpoint answers with

extension Mealplanning_MealPlanSummary: MealPlanDisplayable {
  var displayEvents: [MealPlanEventDisplayable] { events }
}

extension Mealplanning_MealPlanEventSummary: MealPlanEventDisplayable {
  var chosenMealDisplayName: String? {
    hasChosenMealName && !chosenMealName.isEmpty ? chosenMealName : nil
  }
}
