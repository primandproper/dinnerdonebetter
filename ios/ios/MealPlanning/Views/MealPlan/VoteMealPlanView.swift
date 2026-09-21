//
//  VoteMealPlanView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import SwiftProtobuf
import SwiftUI

struct VoteMealPlanView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @Environment(\.dismiss) private var dismiss
  @State private var viewModel: VoteMealPlanViewModel?
  @State private var editMode: EditMode = .active

  let mealPlan: Mealplanning_MealPlan

  init(mealPlan: Mealplanning_MealPlan) {
    self.mealPlan = mealPlan
    // Note: viewModel will be initialized in onAppear to access @Environment authManager
  }

  var body: some View {
    NavigationStack {
      Group {
        if let viewModel = viewModel {
          VStack(spacing: 0) {
            // Static header with meal plan details
            mealPlanHeader(viewModel: viewModel)

            // Side-scrolling events
            GeometryReader { geometry in
              ScrollViewReader { proxy in
                ScrollView(.horizontal, showsIndicators: true) {
                  HStack(spacing: 16) {
                    ForEach(Array(viewModel.mealPlan.events.enumerated()), id: \.element.id) {
                      index, event in
                      eventVotingView(
                        event: event, index: index, viewModel: viewModel, scrollProxy: proxy
                      )
                      .frame(width: geometry.size.width - 32)
                      .id(index)
                    }
                  }
                  .padding(.horizontal, 16)
                }
              }
            }

            // Voting status section (only shown if user is creator)
            if viewModel.isCreator {
              votingStatusSection(viewModel: viewModel)
            }

            // Submit button (only shown when all ballots are complete and locked)
            if viewModel.canSubmit {
              submitButton(viewModel: viewModel)
                .padding()
            }
          }
          .navigationTitle("Vote on Meal Plan")
          .navigationBarTitleDisplayMode(.inline)
          .toolbar {
            ToolbarItem(placement: .navigationBarLeading) {
              Button("Cancel") {
                eventReporterService.reporter.track(event: "vote_cancelled", properties: [:])
                dismiss()
              }
            }
          }
          .alert("Error", isPresented: .constant(viewModel.submissionError != nil)) {
            Button("OK") {
              viewModel.submissionError = nil
            }
          } message: {
            if let error = viewModel.submissionError {
              Text(error)
            }
          }
          .alert(
            "Success",
            isPresented: Binding(
              get: { viewModel.submissionSuccess },
              set: { viewModel.submissionSuccess = $0 }
            )
          ) {
            Button("OK") {
              dismiss()
            }
          } message: {
            Text(
              viewModel.isUpdateMode
                ? "Your votes have been updated successfully!"
                : "Your votes have been submitted successfully!"
            )
          }
        } else {
          DSInitializingView()
        }
      }
      .onAppear {
        if viewModel == nil {
          viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)
          eventReporterService.reporter.track(
            event: "vote_started",
            properties: [
              "meal_plan_id": mealPlan.id,
              "events_count": mealPlan.events.count,
            ])
        }
        if let viewModel = viewModel, viewModel.isCreator {
          Task {
            await viewModel.loadVotingStatus()
          }
        }
      }
    }
  }

  // MARK: - Meal Plan Header

  private func deadlineDateView(deadline: SwiftProtobuf.Google_Protobuf_Timestamp) -> some View {
    let deadlineDate = HomeViewModel.timestampToDate(deadline)
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    formatter.timeStyle = .short

    return HStack {
      Image(systemName: "clock")
        .foregroundColor(.secondary)
        .font(.caption)
      Text(formatter.string(from: deadlineDate))
        .font(.caption)
        .foregroundColor(.secondary)
    }
  }

  private func mealPlanHeader(viewModel: VoteMealPlanViewModel) -> some View {
    VStack(alignment: .leading, spacing: 16) {
      // Title and time range
      VStack(alignment: .leading, spacing: 4) {
        Text(viewModel.mealPlan.notes.isEmpty ? "Meal Plan" : viewModel.mealPlan.notes)
          .font(.title2)
          .fontWeight(.bold)

        Text(MealPlanningHomeHelpers.formatMealPlanTimeRange(viewModel.mealPlan))
          .font(.subheadline)
          .foregroundColor(.secondary)
      }

      DSDivider()

      // Countdown timer section
      VStack(alignment: .leading, spacing: 8) {
        Text("Voting Deadline")
          .font(.headline)

        // Deadline date/time with countdown timer
        HStack(spacing: 12) {
          deadlineDateView(deadline: viewModel.mealPlan.votingDeadline)
          VoteDeadlineCountdown(deadline: viewModel.mealPlan.votingDeadline)
        }
      }

      // Event counter (if multiple events)
      if viewModel.mealPlan.events.count > 1 {
        DSDivider()
        HStack {
          Image(systemName: "calendar")
            .foregroundColor(.secondary)
            .font(.caption)
          Text("\(viewModel.mealPlan.events.count) events")
            .font(.caption)
            .foregroundColor(.secondary)
          Spacer()
          Text("Swipe horizontally to view all")
            .font(.caption)
            .foregroundColor(.secondary)
            .italic()
        }
      }
    }
    .padding()
    .background(Color(.systemGray6))
  }

  // MARK: - Event Voting View

  private func eventVotingView(
    event: Mealplanning_MealPlanEvent, index: Int, viewModel: VoteMealPlanViewModel,
    scrollProxy: ScrollViewProxy
  ) -> some View {
    ScrollView {
      VStack(alignment: .leading, spacing: 16) {
        // Event header
        eventHeader(event: event)

        // Abstain button (only show if not already abstained and not already voted)
        if let ballot = viewModel.getBallot(for: event.id),
          !ballot.isAbstained,
          !viewModel.hasUserVotedOnEvent(eventID: event.id)
        {
          abstainButton(
            event: event, index: index, viewModel: viewModel, scrollProxy: scrollProxy
          )
        }

        // Lock status
        lockStatusView(event: event, viewModel: viewModel)

        // Instructions
        instructionsView(event: event, viewModel: viewModel)

        // Ranked options list (drag and drop)
        rankedOptionsList(event: event, viewModel: viewModel, editMode: $editMode)

        // Lock/Update button
        if let ballot = viewModel.getBallot(for: event.id) {
          LockButtonView(
            event: event,
            ballot: ballot,
            isUpdateMode: viewModel.hasUserVotedOnEvent(eventID: event.id)
              && viewModel.isDeadlineActive
          ) {
            eventReporterService.reporter.track(
              event: "vote_lock_toggled",
              properties: ["event_id": event.id])
            viewModel.toggleLock(eventID: event.id)
          }
          .padding(.top, 8)
        }
      }
      .padding()
    }
  }

  private func abstainButton(
    event: Mealplanning_MealPlanEvent, index: Int, viewModel: VoteMealPlanViewModel,
    scrollProxy: ScrollViewProxy
  ) -> some View {
    Button(
      action: {
        eventReporterService.reporter.track(
          event: "vote_abstain_tapped",
          properties: ["event_id": event.id])
        Task {
          let success = await viewModel.abstainFromEvent(eventID: event.id)
          if success {
            // Scroll to next event if available
            let nextIndex = index + 1
            if nextIndex < viewModel.mealPlan.events.count {
              withAnimation {
                scrollProxy.scrollTo(nextIndex, anchor: .leading)
              }
            }
          }
        }
      },
      label: {
        HStack {
          Image(systemName: "hand.raised.fill")
          Text("Abstain from Voting")
        }
        .font(.headline)
        .frame(maxWidth: .infinity)
        .padding()
        .background(Color.red)
        .foregroundColor(.white)
        .cornerRadius(10)
      }
    )
  }

  private func eventHeader(event: Mealplanning_MealPlanEvent) -> some View {
    let date = HomeViewModel.timestampToDate(event.startsAt)
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    formatter.timeStyle = .short

    return VStack(alignment: .leading, spacing: 8) {
      Text(MealPlanningUtils.formatMealName(event.mealName))
        .font(.title2)
        .fontWeight(.bold)

      Text(formatter.string(from: date))
        .font(.subheadline)
        .foregroundColor(.secondary)
    }
  }

  private func lockStatusView(event: Mealplanning_MealPlanEvent, viewModel: VoteMealPlanViewModel)
    -> some View
  {
    if let ballot = viewModel.getBallot(for: event.id) {
      return AnyView(
        HStack {
          if ballot.isAbstained {
            Label("Abstained", systemImage: "hand.raised.fill")
              .font(.subheadline)
              .foregroundColor(.red)
          } else if ballot.isLocked {
            Label("Locked", systemImage: "lock.fill")
              .font(.subheadline)
              .foregroundColor(.green)
          } else {
            Label("Unlocked", systemImage: "lock.open.fill")
              .font(.subheadline)
              .foregroundColor(.orange)
          }

          Spacer()

          if !ballot.isAbstained && !ballot.isComplete(totalOptions: event.options.count) {
            Text("\(ballot.rankedOptions.count) of \(event.options.count) options ranked")
              .font(.caption)
              .foregroundColor(.secondary)
          }
        }
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(8)
      )
    } else {
      return AnyView(EmptyView())
    }
  }

  private func instructionsView(
    event: Mealplanning_MealPlanEvent, viewModel: VoteMealPlanViewModel
  ) -> some View {
    let ballot = viewModel.getBallot(for: event.id)

    return VStack(alignment: .leading, spacing: 4) {
      if ballot?.isAbstained == true {
        Text("You have abstained from voting")
          .font(.headline)
        Text("You chose not to vote on any options for this event.")
          .font(.subheadline)
          .foregroundColor(.secondary)
      } else if viewModel.hasUserVotedOnEvent(eventID: event.id) && viewModel.isDeadlineActive {
        Text("Update your preferences")
          .font(.headline)
        Text("Unlock to reorder your votes. Your first choice should be at the top.")
          .font(.subheadline)
          .foregroundColor(.secondary)
      } else {
        Text("Rank your preferences")
          .font(.headline)
        Text("Drag options to reorder them. Your first choice should be at the top.")
          .font(.subheadline)
          .foregroundColor(.secondary)
      }
    }
    .padding()
    .background(Color(.systemGray6))
    .cornerRadius(8)
  }

  // MARK: - Voting Status Section

  private func votingStatusSection(viewModel: VoteMealPlanViewModel) -> some View {
    VStack(alignment: .leading, spacing: 12) {
      DSDivider()
        .padding(.horizontal)

      Text("Voting Status")
        .font(.headline)
        .padding(.horizontal)

      if viewModel.isLoadingVotingStatus {
        HStack {
          Spacer()
          ProgressView()
            .padding()
          Spacer()
        }
      } else if viewModel.accountMembers.isEmpty {
        Text("No account members found")
          .font(.subheadline)
          .foregroundColor(.secondary)
          .padding(.horizontal)
      } else {
        ScrollView(.horizontal, showsIndicators: false) {
          HStack(spacing: 12) {
            ForEach(viewModel.accountMembers, id: \.id) { member in
              if member.hasBelongsToUser {
                memberVotingStatusCard(
                  member: member,
                  status: viewModel.votingStatus[member.user.id],
                  totalEvents: viewModel.mealPlan.events.count
                )
              }
            }
          }
          .padding(.horizontal)
        }
      }
    }
    .padding(.vertical, 8)
    .background(Color(.systemGray6))
  }

  private func memberVotingStatusCard(
    member: Primandproper_Platform_Identity_V1_MembershipWithUser,
    status: VoteMealPlanViewModel.VotingStatus?,
    totalEvents: Int
  ) -> some View {
    let userName =
      member.user.displayName.isEmpty
      ? "\(member.user.firstName) \(member.user.lastName)".trimmingCharacters(
        in: .whitespaces)
      : member.user.displayName

    let displayName = userName.isEmpty ? "Unknown User" : userName

    return VStack(alignment: .leading, spacing: 8) {
      Text(displayName)
        .font(.subheadline)
        .fontWeight(.semibold)

      if let status = status {
        if status.hasVoted || status.hasAbstained {
          VStack(alignment: .leading, spacing: 4) {
            if status.hasVoted {
              Label("Voted", systemImage: "checkmark.circle.fill")
                .font(.caption)
                .foregroundColor(.green)
            }
            if status.hasAbstained {
              Label("Abstained", systemImage: "hand.raised.fill")
                .font(.caption)
                .foregroundColor(.orange)
            }
            Text(
              "\(status.eventsVoted.count + status.eventsAbstained.count) of \(totalEvents) events"
            )
            .font(.caption2)
            .foregroundColor(.secondary)
          }
        } else {
          Label("Not voted", systemImage: "circle")
            .font(.caption)
            .foregroundColor(.secondary)
        }
      } else {
        Label("Not voted", systemImage: "circle")
          .font(.caption)
          .foregroundColor(.secondary)
      }
    }
    .padding(12)
    .frame(minWidth: 120)
    .background(Color(.systemBackground))
    .cornerRadius(8)
    .shadow(color: Color.black.opacity(0.1), radius: 2, x: 0, y: 1)
  }

  // MARK: - Submit Button

  private func submitButton(viewModel: VoteMealPlanViewModel) -> some View {
    Button(
      action: {
        Task {
          let success = await viewModel.submitVotes()
          if success {
            let eventsVoted = viewModel.mealPlan.events.filter {
              guard let ballot = viewModel.getBallot(for: $0.id) else { return false }
              return !ballot.isAbstained
            }.count
            let eventsAbstained = viewModel.mealPlan.events.filter {
              viewModel.getBallot(for: $0.id)?.isAbstained == true
            }.count
            eventReporterService.reporter.track(
              event: "vote_submitted",
              properties: [
                "meal_plan_id": viewModel.mealPlan.id,
                "events_voted": eventsVoted,
                "events_abstained": eventsAbstained,
              ])
            // Alert will be shown via viewModel.submissionSuccess
          }
        }
      },
      label: {
        HStack {
          if viewModel.isSubmitting {
            ProgressView()
              .progressViewStyle(CircularProgressViewStyle(tint: .white))
          }
          Text(
            viewModel.isSubmitting
              ? (viewModel.isUpdateMode ? "Updating..." : "Submitting...")
              : (viewModel.isUpdateMode ? "Update Votes" : "Submit Votes")
          )
          .fontWeight(.semibold)
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(viewModel.isSubmitting ? Color.gray : Color.green)
        .foregroundColor(.white)
        .cornerRadius(10)
      }
    )
    .disabled(viewModel.isSubmitting)
  }
}

// MARK: - Lock Button View

struct LockButtonView: View {
  let event: Mealplanning_MealPlanEvent
  let ballot: EventBallot
  let isUpdateMode: Bool
  let onToggle: () -> Void

  var body: some View {
    Button(action: onToggle) {
      HStack {
        if isUpdateMode {
          Image(systemName: ballot.isLocked ? "arrow.clockwise" : "lock.open.fill")
          Text(ballot.isLocked ? "Update Vote" : "Unlock to Update")
        } else {
          Image(systemName: ballot.isLocked ? "lock.fill" : "lock.open.fill")
          Text(ballot.isLocked ? "Unlock Ballot" : "Lock Ballot")
        }
      }
      .font(.headline)
      .frame(maxWidth: .infinity)
      .padding()
      .background(buttonColor)
      .foregroundColor(.white)
      .cornerRadius(10)
    }
    .disabled(!ballot.isComplete(totalOptions: event.options.count))
  }

  private var buttonColor: Color {
    if isUpdateMode {
      return ballot.isLocked ? Color.blue : Color.orange
    } else {
      return ballot.isLocked ? Color.orange : Color.blue
    }
  }
}

// MARK: - Vote Deadline Countdown

struct VoteDeadlineCountdown: View {
  let deadline: SwiftProtobuf.Google_Protobuf_Timestamp
  @State private var timeRemaining: TimeInterval = 0
  @State private var timer: Timer?

  var body: some View {
    HStack(spacing: 6) {
      Image(systemName: "clock.fill")
        .font(.caption)
      Text(formattedTimeRemaining)
        .font(.caption)
        .fontWeight(.semibold)
        .monospacedDigit()
    }
    .foregroundColor(timeRemainingColor)
    .onAppear {
      updateTimeRemaining()
      startTimer()
    }
    .onDisappear {
      stopTimer()
    }
  }

  private var formattedTimeRemaining: String {
    if timeRemaining <= 0 {
      return "Expired"
    }

    let days = Int(timeRemaining) / 86400
    let hours = (Int(timeRemaining) % 86400) / 3600
    let minutes = (Int(timeRemaining) % 3600) / 60
    let seconds = Int(timeRemaining) % 60

    if days > 0 {
      return String(format: "%dd %02d:%02d:%02d", days, hours, minutes, seconds)
    } else if hours > 0 {
      return String(format: "%02d:%02d:%02d", hours, minutes, seconds)
    } else {
      return String(format: "%02d:%02d", minutes, seconds)
    }
  }

  private var timeRemainingColor: Color {
    if timeRemaining <= 0 {
      return .red
    } else if timeRemaining < 3600 {  // Less than 1 hour
      return .red
    } else if timeRemaining < 86400 {  // Less than 1 day
      return .orange
    } else {
      return .primary
    }
  }

  private func updateTimeRemaining() {
    let deadlineDate = HomeViewModel.timestampToDate(deadline)
    let now = Date()
    timeRemaining = max(0, deadlineDate.timeIntervalSince(now))
  }

  private func startTimer() {
    timer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { _ in
      updateTimeRemaining()
    }
    if let timer = timer {
      RunLoop.main.add(timer, forMode: .common)
    }
  }

  private func stopTimer() {
    timer?.invalidate()
    timer = nil
  }
}

#Preview {
  // This would need a mock meal plan for preview
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "Test User"
  authManager.userID = "user123"
  authManager.accountID = "account123"

  // Create a mock meal plan
  var mealPlan = Mealplanning_MealPlan()
  mealPlan.id = "test-plan"
  mealPlan.notes = "Test Meal Plan"

  var event = Mealplanning_MealPlanEvent()
  event.id = "test-event"
  event.mealName = .dinner

  var option1 = Mealplanning_MealPlanOption()
  option1.id = "option1"
  var meal1 = Mealplanning_Meal()
  meal1.name = "Pasta"
  option1.meal = meal1

  var option2 = Mealplanning_MealPlanOption()
  option2.id = "option2"
  var meal2 = Mealplanning_Meal()
  meal2.name = "Pizza"
  option2.meal = meal2

  event.options = [option1, option2]
  mealPlan.events = [event]

  return VoteMealPlanView(mealPlan: mealPlan)
    .environment(authManager)
}
