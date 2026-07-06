//
//  VoteMealPlanViewModel.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import Foundation
import GRPCCore
import SwiftProtobuf
import SwiftUI

/// Represents a ballot for a single event with ranked options
struct EventBallot: Identifiable {
  let id: String  // Event ID
  var rankedOptions: [Mealplanning_MealPlanOption]  // Ordered list (first = rank 0)
  var isLocked: Bool = false
  // Maps option ID to vote ID (for updates)
  var optionVoteIDs: [String: String] = [:]
  // Tracks if the user has abstained from this event
  var isAbstained: Bool = false
  // Maps option ID to abstain vote ID (for updates)
  var abstainVoteIDs: [String: String] = [:]

  /// Check if ballot is complete (has all options)
  func isComplete(totalOptions: Int) -> Bool {
    return rankedOptions.count == totalOptions
  }
}

@Observable
@MainActor
class VoteMealPlanViewModel {
  let mealPlan: Mealplanning_MealPlan
  private let authManager: AuthenticationManager

  // Ballots for each event (keyed by event ID)
  var ballots: [String: EventBallot] = [:]

  // Submission state
  var isSubmitting = false
  var submissionError: String?
  var submissionSuccess = false

  // Voting status for account members (only loaded if user is creator)
  var accountMembers: [Identity_AccountUserMembershipWithUser] = []
  var votingStatus: [String: VotingStatus] = [:]  // Maps user ID to voting status
  var isLoadingVotingStatus = false

  /// Represents the voting status for a user
  struct VotingStatus {
    let hasVoted: Bool
    let hasAbstained: Bool
    let eventsVoted: Set<String>  // Event IDs where user has voted
    let eventsAbstained: Set<String>  // Event IDs where user has abstained
  }

  init(mealPlan: Mealplanning_MealPlan, authManager: AuthenticationManager) {
    self.mealPlan = mealPlan
    self.authManager = authManager

    // Initialize ballots for each event
    for event in mealPlan.events {
      // Check if user has abstained from this event
      var hasAbstained = false
      var abstainVoteIDs: [String: String] = [:]
      for option in event.options {
        if let vote = option.votes.first(where: { $0.byUser == authManager.userID && $0.abstain }) {
          hasAbstained = true
          abstainVoteIDs[option.id] = vote.id
        }
      }

      if hasAbstained {
        // User has abstained - mark ballot as abstained
        ballots[event.id] = EventBallot(
          id: event.id,
          rankedOptions: event.options,  // Keep all options for display
          isLocked: true,  // Locked because abstained
          isAbstained: true,
          abstainVoteIDs: abstainVoteIDs
        )
        continue
      }

      // Check if user has already voted on this event
      var userVotes: [(option: Mealplanning_MealPlanOption, rank: UInt32)] = []
      for option in event.options {
        if let vote = option.votes.first(where: { $0.byUser == authManager.userID && !$0.abstain })
        {
          userVotes.append((option: option, rank: vote.rank))
        }
      }

      if !userVotes.isEmpty {
        // User has voted - initialize with their existing ranking
        // Sort by rank (lower rank = higher preference)
        userVotes.sort { $0.rank < $1.rank }
        let rankedOptions = userVotes.map { $0.option }

        // Build map of option ID to vote ID for updates
        var optionVoteIDs: [String: String] = [:]
        for option in event.options {
          if let vote = option.votes.first(where: { $0.byUser == authManager.userID && !$0.abstain }
          ) {
            optionVoteIDs[option.id] = vote.id
          }
        }

        // Add any options that weren't voted on (shouldn't happen, but be safe)
        let votedOptionIDs = Set(rankedOptions.map { $0.id })
        let unvotedOptions = event.options.filter { !votedOptionIDs.contains($0.id) }

        ballots[event.id] = EventBallot(
          id: event.id,
          rankedOptions: rankedOptions + unvotedOptions,
          isLocked: true,  // Start locked if user has already voted
          optionVoteIDs: optionVoteIDs
        )
      } else {
        // No existing votes - initialize with all options in original order
        ballots[event.id] = EventBallot(
          id: event.id,
          rankedOptions: event.options
        )
      }
    }
  }

  // MARK: - Ballot Management

  /// Get ballot for an event
  func getBallot(for eventID: String) -> EventBallot? {
    return ballots[eventID]
  }

  /// Reorder options in a ballot (drag and drop)
  func reorderOptions(eventID: String, from source: IndexSet, to destination: Int) {
    guard var ballot = ballots[eventID] else { return }

    // Don't allow reordering if locked
    guard !ballot.isLocked else { return }

    ballot.rankedOptions.move(fromOffsets: source, toOffset: destination)
    ballots[eventID] = ballot
  }

  /// Lock/unlock a ballot for an event
  func toggleLock(eventID: String) {
    guard var ballot = ballots[eventID] else { return }

    // Can only lock if ballot is complete
    let event = mealPlan.events.first { $0.id == eventID }
    guard let event = event else { return }

    if !ballot.isLocked && !ballot.isComplete(totalOptions: event.options.count) {
      // Cannot lock incomplete ballot
      return
    }

    ballot.isLocked.toggle()
    ballots[eventID] = ballot
  }

  /// Check if all ballots are complete and locked (or abstained)
  var canSubmit: Bool {
    // Can't submit if there are no events
    guard !mealPlan.events.isEmpty else { return false }

    for event in mealPlan.events {
      guard let ballot = ballots[event.id] else { return false }
      // Ballot is valid if it's abstained, or if it's locked and complete
      if !ballot.isAbstained
        && (!ballot.isLocked || !ballot.isComplete(totalOptions: event.options.count))
      {
        return false
      }
    }
    return true
  }

  /// Check if user has already voted on this meal plan
  var hasUserVoted: Bool {
    for event in mealPlan.events {
      for option in event.options
      where option.votes.contains(where: { $0.byUser == authManager.userID && !$0.abstain }) {
        return true
      }
    }
    return false
  }

  /// Check if voting deadline has not expired
  var isDeadlineActive: Bool {
    let deadline = HomeViewModel.timestampToDate(mealPlan.votingDeadline)
    return deadline > Date()
  }

  /// Check if we're in update mode (user has voted and deadline hasn't expired)
  var isUpdateMode: Bool {
    return hasUserVoted && isDeadlineActive
  }

  /// Check if user has voted on a specific event
  func hasUserVotedOnEvent(eventID: String) -> Bool {
    guard let event = mealPlan.events.first(where: { $0.id == eventID }) else { return false }
    return event.options.contains { option in
      option.votes.contains(where: { $0.byUser == authManager.userID && !$0.abstain })
    }
  }

  /// Check if user has abstained from a specific event
  func hasUserAbstainedFromEvent(eventID: String) -> Bool {
    return ballots[eventID]?.isAbstained ?? false
  }

  /// Check if the current user is the creator of the meal plan
  var isCreator: Bool {
    return mealPlan.createdByUser == authManager.userID
  }

  /// Load voting status for all account members (only if user is creator)
  func loadVotingStatus() async {
    guard isCreator else { return }

    isLoadingVotingStatus = true

    do {
      // Get account details to get members
      var accountRequest = Identity_GetAccountRequest()
      accountRequest.accountID = mealPlan.belongsToAccount

      let accountResponse = try await authManager.authenticatedCall("getAccount", idempotent: true)
      {
        client, metadata, options in
        try await client.identity.getAccount(accountRequest, metadata: metadata, options: options)
      }

      guard accountResponse.hasResult else {
        print("⚠️ Account not found")
        isLoadingVotingStatus = false
        return
      }

      let account = accountResponse.result
      accountMembers = account.members

      // Determine voting status for each member
      var statusMap: [String: VotingStatus] = [:]

      for member in account.members {
        guard member.hasBelongsToUser else { continue }
        let userID = member.belongsToUser.id

        var eventsVoted: Set<String> = []
        var eventsAbstained: Set<String> = []

        // Check each event
        for event in mealPlan.events {
          var hasVotedInEvent = false
          var hasAbstainedInEvent = false

          // Check all options in the event
          for option in event.options {
            for vote in option.votes where vote.byUser == userID {
              if vote.abstain {
                hasAbstainedInEvent = true
              } else {
                hasVotedInEvent = true
              }
              break
            }
            // If we found a vote (either regular or abstain), we can stop checking this event
            if hasVotedInEvent || hasAbstainedInEvent {
              break
            }
          }

          if hasVotedInEvent {
            eventsVoted.insert(event.id)
          } else if hasAbstainedInEvent {
            eventsAbstained.insert(event.id)
          }
        }

        let hasVoted = !eventsVoted.isEmpty
        let hasAbstained = !eventsAbstained.isEmpty

        statusMap[userID] = VotingStatus(
          hasVoted: hasVoted,
          hasAbstained: hasAbstained,
          eventsVoted: eventsVoted,
          eventsAbstained: eventsAbstained
        )
      }

      votingStatus = statusMap
    } catch {
      // Non-fatal: leave voting status as-is (session errors handled by authenticatedCall)
    }

    isLoadingVotingStatus = false
  }

  /// Abstain from voting on a specific event
  func abstainFromEvent(eventID: String) async -> Bool {
    guard let event = mealPlan.events.first(where: { $0.id == eventID }) else {
      return false
    }

    // Check if already abstained
    if hasUserAbstainedFromEvent(eventID: eventID) {
      return true
    }

    do {
      // Create abstain votes for all options in this event
      var request = Mealplanning_CreateMealPlanOptionVoteRequest()
      request.mealPlanID = mealPlan.id
      request.mealPlanEventID = event.id

      var input = Mealplanning_MealPlanOptionVoteCreationRequestInput()

      // Create abstain votes for all options
      for option in event.options {
        var voteInput = Mealplanning_MealPlanOptionVoteCreationInput()
        voteInput.belongsToMealPlanOption = option.id
        voteInput.rank = 0  // Rank doesn't matter for abstain votes
        voteInput.notes = ""
        voteInput.abstain = true  // Mark as abstain

        input.votes.append(voteInput)
      }

      request.input = input

      // Submit abstain votes for this event
      _ = try await authManager.authenticatedCall("createMealPlanOptionVote") {
        client, metadata, options in
        try await client.mealPlanning.createMealPlanOptionVote(
          request, metadata: metadata, options: options)
      }

      // Update the ballot to reflect abstention
      // Note: We'd need to extract vote IDs from the response, but for now we'll mark it as abstained
      // The vote IDs will be populated when the meal plan is refreshed
      if var ballot = ballots[event.id] {
        ballot.isAbstained = true
        ballot.isLocked = true
        ballots[event.id] = ballot
      }

      return true
    } catch {
      return false
    }
  }

  // MARK: - Vote Submission

  /// Submit all votes for all events
  func submitVotes() async -> Bool {
    guard canSubmit else {
      submissionError = "All ballots must be complete and locked before submitting"
      return false
    }

    isSubmitting = true
    submissionError = nil
    submissionSuccess = false

    do {
      // Submit or update votes for each event
      for event in mealPlan.events {
        guard let ballot = ballots[event.id] else {
          continue
        }

        // Skip abstained events - they're already submitted
        if ballot.isAbstained {
          continue
        }

        let hasExistingVotes = hasUserVotedOnEvent(eventID: event.id)

        if hasExistingVotes {
          // Update existing votes
          for (index, option) in ballot.rankedOptions.enumerated() {
            guard let voteID = ballot.optionVoteIDs[option.id] else {
              // This shouldn't happen, but if it does, skip this option
              print("⚠️ Warning: No vote ID found for option \(option.id) in event \(event.id)")
              continue
            }

            var updateRequest = Mealplanning_UpdateMealPlanOptionVoteRequest()
            updateRequest.mealPlanID = mealPlan.id
            updateRequest.mealPlanEventID = event.id
            updateRequest.mealPlanOptionID = option.id
            updateRequest.mealPlanOptionVoteID = voteID

            var updateInput = Mealplanning_MealPlanOptionVoteUpdateRequestInput()
            updateInput.rank = UInt32(index)  // Rank 0 = first choice, 1 = second choice, etc.
            updateInput.notes = ""
            updateInput.abstain = false
            updateInput.belongsToMealPlanOption = option.id

            updateRequest.input = updateInput

            // Update vote for this option
            _ = try await authManager.authenticatedCall("updateMealPlanOptionVote") {
              client, metadata, options in
              try await client.mealPlanning.updateMealPlanOptionVote(
                updateRequest, metadata: metadata, options: options)
            }
          }
        } else {
          // Create new votes
          var request = Mealplanning_CreateMealPlanOptionVoteRequest()
          request.mealPlanID = mealPlan.id
          request.mealPlanEventID = event.id

          var input = Mealplanning_MealPlanOptionVoteCreationRequestInput()

          // Create votes with ranks based on order in rankedOptions
          for (index, option) in ballot.rankedOptions.enumerated() {
            var voteInput = Mealplanning_MealPlanOptionVoteCreationInput()
            voteInput.belongsToMealPlanOption = option.id
            voteInput.rank = UInt32(index)  // Rank 0 = first choice, 1 = second choice, etc.
            voteInput.notes = ""
            voteInput.abstain = false

            input.votes.append(voteInput)
          }

          request.input = input

          // Submit votes for this event
          _ = try await authManager.authenticatedCall("createMealPlanOptionVote") {
            client, metadata, options in
            try await client.mealPlanning.createMealPlanOptionVote(
              request, metadata: metadata, options: options)
          }
        }
      }

      submissionSuccess = true
      isSubmitting = false
      return true
    } catch {
      submissionError = "Failed to submit votes: \(error.localizedDescription)"
      isSubmitting = false
      return false
    }
  }
}
