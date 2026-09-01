//
//  HomeViewHelpers.swift
//  ios
//
//  Meal planning home screen helper functions.
//

import SwiftProtobuf
import SwiftUI

enum MealPlanningHomeHelpers {
  private static func timestampToDate(_ timestamp: SwiftProtobuf.Google_Protobuf_Timestamp) -> Date
  {
    let seconds = TimeInterval(timestamp.seconds)
    let nanos = TimeInterval(timestamp.nanos) / 1_000_000_000.0
    return Date(timeIntervalSince1970: seconds + nanos)
  }

  static func formatMealPlanTimeRange(_ mealPlan: any MealPlanDisplayable) -> String {
    guard !mealPlan.displayEvents.isEmpty else {
      return ""
    }

    let earliestStart =
      mealPlan.displayEvents.map { timestampToDate($0.startsAt) }.min() ?? Date()
    let latestEnd = mealPlan.displayEvents.map { timestampToDate($0.endsAt) }.max() ?? Date()

    let dateFormatter = DateFormatter()
    dateFormatter.dateStyle = .medium
    dateFormatter.timeStyle = .none

    let startString = dateFormatter.string(from: earliestStart)

    let calendar = Calendar.current
    if calendar.isDate(earliestStart, inSameDayAs: latestEnd) {
      let timeFormatter = DateFormatter()
      timeFormatter.dateStyle = .none
      timeFormatter.timeStyle = .short
      return
        "\(startString) • \(timeFormatter.string(from: earliestStart)) - \(timeFormatter.string(from: latestEnd))"
    } else {
      let endString = dateFormatter.string(from: latestEnd)
      return "\(startString) - \(endString)"
    }
  }

  /// Compact date range for subtitle when title already conveys meal info (e.g. "Mar 12–14" or "Tue, Thu 7:00 PM").
  static func formatMealPlanTimeRangeCompact(_ mealPlan: any MealPlanDisplayable) -> String {
    guard !mealPlan.displayEvents.isEmpty else {
      return ""
    }

    let eventDates = mealPlan.displayEvents.map { timestampToDate($0.startsAt) }.sorted()
    let earliestStart = eventDates.first ?? Date()
    let latestEnd = mealPlan.displayEvents.map { timestampToDate($0.endsAt) }.max() ?? Date()

    let dateFormatter = DateFormatter()
    dateFormatter.dateFormat = "MMM d"

    let calendar = Calendar.current
    if calendar.isDate(earliestStart, inSameDayAs: latestEnd) {
      let timeFormatter = DateFormatter()
      timeFormatter.dateStyle = .none
      timeFormatter.timeStyle = .short
      return
        "\(dateFormatter.string(from: earliestStart)) • \(timeFormatter.string(from: earliestStart))"
    }

    if eventDates.count <= 3 {
      let weekdayFormatter = DateFormatter()
      weekdayFormatter.dateFormat = "EEE"
      let timeFormatter = DateFormatter()
      timeFormatter.dateStyle = .none
      timeFormatter.timeStyle = .short
      let parts = eventDates.map {
        "\(weekdayFormatter.string(from: $0)) \(timeFormatter.string(from: $0))"
      }
      return parts.joined(separator: ", ")
    }

    let startString = dateFormatter.string(from: earliestStart)
    let endString = dateFormatter.string(from: latestEnd)
    return "\(startString)–\(endString)"
  }

  /// Display names for the meals voting settled on, skipping events still undecided.
  static func chosenMealDisplayNames(from mealPlan: any MealPlanDisplayable) -> [String] {
    mealPlan.displayEvents.compactMap { $0.chosenMealDisplayName }
  }

  /// Whether notes is the auto-generated default from the wizard.
  static func isDefaultMealPlanTitle(_ notes: String, mealPlan: any MealPlanDisplayable) -> Bool {
    let trimmed = notes.trimmingCharacters(in: .whitespacesAndNewlines)
    guard !trimmed.isEmpty else { return true }

    let dateRange = formatMealPlanTimeRange(mealPlan)
    guard !dateRange.isEmpty else { return false }

    if trimmed == "Dinners \(dateRange)" {
      return true
    }

    if mealPlan.displayEvents.count == 1, trimmed.hasPrefix("Meal Plan for ") {
      let startDate =
        mealPlan.displayEvents.map { timestampToDate($0.startsAt) }.min() ?? Date()
      let formatter = DateFormatter()
      formatter.dateStyle = .medium
      formatter.timeStyle = .none
      return trimmed == "Meal Plan for \(formatter.string(from: startDate))"
    }

    return false
  }

  /// Whether the meal plan's date range is entirely in the past (end date < now).
  static func isMealPlanInPast(_ mealPlan: any MealPlanDisplayable) -> Bool {
    guard !mealPlan.displayEvents.isEmpty else { return false }
    let latestEnd =
      mealPlan.displayEvents.map { timestampToDate($0.endsAt) }.max() ?? Date.distantFuture
    return latestEnd < Date()
  }

  static func mealPlanDisplayTitle(_ mealPlan: any MealPlanDisplayable, fallback: String) -> String
  {
    let title = mealPlan.notes.trimmingCharacters(in: .whitespacesAndNewlines)

    if !title.isEmpty && !isDefaultMealPlanTitle(title, mealPlan: mealPlan) {
      return title
    }

    let names = chosenMealDisplayNames(from: mealPlan)
    if !names.isEmpty {
      return names.joined(separator: " & ")
    }

    guard !mealPlan.displayEvents.isEmpty else {
      return fallback
    }

    let earliestStart =
      mealPlan.displayEvents.map { timestampToDate($0.startsAt) }.min() ?? Date()
    let latestEnd = mealPlan.displayEvents.map { timestampToDate($0.endsAt) }.max() ?? Date()
    let dateFormatter = DateFormatter()
    dateFormatter.dateFormat = "MMM d"
    let calendar = Calendar.current
    let compactRange: String
    if calendar.isDate(earliestStart, inSameDayAs: latestEnd) {
      compactRange = dateFormatter.string(from: earliestStart)
    } else {
      compactRange =
        "\(dateFormatter.string(from: earliestStart))–\(dateFormatter.string(from: latestEnd))"
    }
    let eventCount = mealPlan.displayEvents.count
    let mealType = eventCount == 1 ? "dinner" : "\(eventCount) dinners"
    return "\(mealType.capitalized) • \(compactRange)"
  }
}
