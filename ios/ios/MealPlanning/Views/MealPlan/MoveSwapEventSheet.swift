//
//  MoveSwapEventSheet.swift
//  ios
//
//  Created by Auto.
//

import SwiftProtobuf
import SwiftUI

struct MoveSwapEventSheet: View {
  @Environment(\.dismiss) private var dismiss
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService

  let mealPlan: Mealplanning_MealPlan
  let event: Mealplanning_MealPlanEvent
  let onSuccess: () async -> Void
  var startInMoveMode = false
  var startInSwapMode = false

  @State private var mode: Mode = .choose
  @State private var selectedDate = Date()
  @State private var selectedSwapEvent: Mealplanning_MealPlanEvent?
  @State private var isUpdating = false
  @State private var errorMessage: String?

  private enum Mode {
    case choose
    case move
    case swap
  }

  private var otherEvents: [Mealplanning_MealPlanEvent] {
    mealPlan.events.filter { $0.id != event.id }
  }

  private var canSwap: Bool {
    otherEvents.count >= 1
  }

  private var datesWithOtherEvents: Set<DateComponents> {
    let cal = Calendar.current
    return Set(
      otherEvents.map { evt in
        cal.dateComponents(
          [.year, .month, .day], from: HomeViewModel.timestampToDate(evt.startsAt))
      }
    )
  }

  private func selectedDateHasOtherEvent() -> Bool {
    let cal = Calendar.current
    let selectedDay = cal.dateComponents([.year, .month, .day], from: selectedDate)
    return datesWithOtherEvents.contains(selectedDay)
  }

  private func selectedDateIsSameAsEvent() -> Bool {
    let cal = Calendar.current
    let eventDay = cal.dateComponents(
      [.year, .month, .day], from: HomeViewModel.timestampToDate(event.startsAt))
    let selectedDay = cal.dateComponents([.year, .month, .day], from: selectedDate)
    return eventDay.year == selectedDay.year && eventDay.month == selectedDay.month
      && eventDay.day == selectedDay.day
  }

  var body: some View {
    NavigationStack {
      Group {
        switch mode {
        case .choose:
          chooseModeContent
        case .move:
          moveModeContent
        case .swap:
          swapModeContent
        }
      }
      .navigationTitle("Reschedule Meal")
      .navigationBarTitleDisplayMode(.inline)
      .toolbar {
        ToolbarItem(placement: .cancellationAction) {
          Button("Cancel") {
            eventReporterService.reporter.track(event: "move_swap_cancelled", properties: [:])
            dismiss()
          }
        }
        if mode != .choose {
          ToolbarItem(placement: .confirmationAction) {
            Button("Back") {
              mode = .choose
              errorMessage = nil
            }
          }
        }
      }
      .alert("Update Failed", isPresented: .constant(errorMessage != nil)) {
        Button("OK") { errorMessage = nil }
      } message: {
        if let msg = errorMessage {
          Text(msg)
        }
      }
      .disabled(isUpdating)
    }
    .onAppear {
      eventReporterService.reporter.track(event: "move_swap_sheet_opened", properties: [:])
      if startInMoveMode {
        mode = .move
        selectedDate = HomeViewModel.timestampToDate(event.startsAt)
      } else if startInSwapMode {
        mode = .swap
      }
    }
  }

  private var chooseModeContent: some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.lg) {
      Text("What would you like to do?")
        .font(.headline)

      Button {
        mode = .move
        selectedDate = HomeViewModel.timestampToDate(event.startsAt)
      } label: {
        HStack {
          Image(systemName: "calendar")
          Text("Move to another day")
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(10)
      }
      .buttonStyle(.plain)

      if canSwap {
        Button {
          mode = .swap
        } label: {
          HStack {
            Image(systemName: "arrow.triangle.2.circlepath")
            Text("Swap with another event")
          }
          .frame(maxWidth: .infinity, alignment: .leading)
          .padding()
          .background(Color(.systemGray6))
          .cornerRadius(10)
        }
        .buttonStyle(.plain)
      }
    }
    .padding()
  }

  private var moveModeContent: some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.lg) {
      Text("Select new date")
        .font(.headline)

      DatePicker(
        "Date",
        selection: $selectedDate,
        in: datePickerRange,
        displayedComponents: .date
      )
      .datePickerStyle(.graphical)

      if let notice = moveDateNotice {
        HStack(spacing: DSTheme.Spacing.sm) {
          Image(systemName: "info.circle.fill")
            .foregroundColor(.secondary)
          Text(notice)
            .font(.subheadline)
            .foregroundColor(.secondary)
        }
        .padding(.vertical, DSTheme.Spacing.sm)
        .padding(.horizontal, DSTheme.Spacing.md)
        .background(Color(.systemGray6))
        .cornerRadius(8)
      }

      Button {
        Task { await performMove() }
      } label: {
        if isUpdating {
          ProgressView()
            .frame(maxWidth: .infinity)
        } else {
          Text("Move")
            .frame(maxWidth: .infinity)
        }
      }
      .buttonStyle(.borderedProminent)
      .disabled(isUpdating || selectedDateIsSameAsEvent() || selectedDateHasOtherEvent())
    }
    .padding()
  }

  private var moveDateNotice: String? {
    if selectedDateIsSameAsEvent() {
      return "This is the current date. Choose a different day."
    }
    if let evt = eventOnDate(selectedDate) {
      return
        "\(MealPlanningUtils.formatMealName(evt.mealName)) is already scheduled that day. Use \"Swap with another event\" to exchange places."
    }
    return nil
  }

  private func eventOnDate(_ date: Date) -> Mealplanning_MealPlanEvent? {
    otherEvents.first { evt in
      let cal = Calendar.current
      let evtDay = cal.dateComponents(
        [.year, .month, .day], from: HomeViewModel.timestampToDate(evt.startsAt))
      let day = cal.dateComponents([.year, .month, .day], from: date)
      return evtDay.year == day.year && evtDay.month == day.month && evtDay.day == day.day
    }
  }

  private var swapModeContent: some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.lg) {
      Text("Select event to swap with")
        .font(.headline)

      List(otherEvents, id: \.id) { otherEvent in
        Button {
          selectedSwapEvent = otherEvent
          Task { await performSwap(with: otherEvent) }
        } label: {
          HStack {
            VStack(alignment: .leading, spacing: 4) {
              Text(MealPlanningUtils.formatMealName(otherEvent.mealName))
                .font(.subheadline)
                .fontWeight(.medium)
              Text(formatEventDate(otherEvent))
                .font(.caption)
                .foregroundColor(.secondary)
            }
            Spacer()
            if isUpdating && selectedSwapEvent?.id == otherEvent.id {
              ProgressView()
            }
          }
        }
        .disabled(isUpdating)
      }
    }
    .padding()
  }

  private var datePickerRange: ClosedRange<Date> {
    let cal = Calendar.current
    let planStart =
      mealPlan.events.map { HomeViewModel.timestampToDate($0.startsAt) }.min() ?? Date()
    let planEnd = mealPlan.events.map { HomeViewModel.timestampToDate($0.endsAt) }.max() ?? Date()
    let minDate = cal.date(byAdding: .weekOfYear, value: -2, to: planStart) ?? planStart
    let maxDate = cal.date(byAdding: .weekOfYear, value: 2, to: planEnd) ?? planEnd
    return minDate...maxDate
  }

  private func formatEventDate(_ evt: Mealplanning_MealPlanEvent) -> String {
    let date = HomeViewModel.timestampToDate(evt.startsAt)
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    formatter.timeStyle = .short
    return formatter.string(from: date)
  }

  private func performMove() async {
    if selectedDateIsSameAsEvent() {
      errorMessage = "That's the same day. No change needed."
      return
    }
    if selectedDateHasOtherEvent() {
      errorMessage =
        "That date already has a meal scheduled. Use \"Swap with another event\" to exchange places instead."
      return
    }

    isUpdating = true
    errorMessage = nil

    do {
      let originalStart = HomeViewModel.timestampToDate(event.startsAt)
      let originalEnd = HomeViewModel.timestampToDate(event.endsAt)
      let cal = Calendar.current

      let dayComponents = cal.dateComponents([.year, .month, .day], from: selectedDate)
      let timeStartComponents = cal.dateComponents([.hour, .minute, .second], from: originalStart)
      let timeEndComponents = cal.dateComponents([.hour, .minute, .second], from: originalEnd)

      var newStartComponents = DateComponents()
      newStartComponents.year = dayComponents.year
      newStartComponents.month = dayComponents.month
      newStartComponents.day = dayComponents.day
      newStartComponents.hour = timeStartComponents.hour ?? 0
      newStartComponents.minute = timeStartComponents.minute ?? 0
      newStartComponents.second = timeStartComponents.second ?? 0

      var newEndComponents = DateComponents()
      newEndComponents.year = dayComponents.year
      newEndComponents.month = dayComponents.month
      newEndComponents.day = dayComponents.day
      newEndComponents.hour = timeEndComponents.hour ?? 0
      newEndComponents.minute = timeEndComponents.minute ?? 0
      newEndComponents.second = timeEndComponents.second ?? 0

      guard let newStart = cal.date(from: newStartComponents),
        let newEnd = cal.date(from: newEndComponents)
      else {
        errorMessage = "Invalid date"
        isUpdating = false
        return
      }

      var startsAt = SwiftProtobuf.Google_Protobuf_Timestamp()
      startsAt.seconds = Int64(newStart.timeIntervalSince1970)
      startsAt.nanos = Int32(
        (newStart.timeIntervalSince1970 - Double(startsAt.seconds)) * 1_000_000_000)

      var endsAt = SwiftProtobuf.Google_Protobuf_Timestamp()
      endsAt.seconds = Int64(newEnd.timeIntervalSince1970)
      endsAt.nanos = Int32((newEnd.timeIntervalSince1970 - Double(endsAt.seconds)) * 1_000_000_000)

      var input = Mealplanning_MealPlanEventUpdateRequestInput()
      input.startsAt = startsAt
      input.endsAt = endsAt
      input.belongsToMealPlan = mealPlan.id

      var request = Mealplanning_UpdateMealPlanEventRequest()
      request.mealPlanID = mealPlan.id
      request.mealPlanEventID = event.id
      request.input = input

      _ = try await authManager.authenticatedCall("updateMealPlanEvent") {
        client, metadata, options in
        try await client.mealPlanning.updateMealPlanEvent(
          request, metadata: metadata, options: options)
      }

      eventReporterService.reporter.track(event: "move_swap_completed", properties: [:])
      await onSuccess()
      NotificationCenter.default.post(name: .mealPlanEventsUpdated, object: nil)
      dismiss()
    } catch {
      errorMessage = error.localizedDescription
    }

    isUpdating = false
  }

  private func performSwap(with otherEvent: Mealplanning_MealPlanEvent) async {
    isUpdating = true
    errorMessage = nil

    do {
      var request = Mealplanning_SwapMealPlanEventsRequest()
      request.mealPlanID = mealPlan.id
      request.mealPlanEventIDA = event.id
      request.mealPlanEventIDB = otherEvent.id

      _ = try await authManager.authenticatedCall("swapMealPlanEvents") {
        client, metadata, options in
        try await client.mealPlanning.swapMealPlanEvents(
          request, metadata: metadata, options: options)
      }

      eventReporterService.reporter.track(event: "move_swap_completed", properties: [:])
      await onSuccess()
      NotificationCenter.default.post(name: .mealPlanEventsUpdated, object: nil)
      dismiss()
    } catch {
      errorMessage = error.localizedDescription
    }

    isUpdating = false
  }
}
