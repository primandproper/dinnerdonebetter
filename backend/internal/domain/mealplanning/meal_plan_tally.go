package mealplanning

import (
	"context"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"

	"resenje.org/schulze"
)

// ErrAlreadyFinalized is returned when a meal plan that has already been finalized is tallied
// again.
var ErrAlreadyFinalized = platformerrors.New("meal plan already finalized")

type (
	// MealPlanElectorate answers who votes on an account's meal plans.
	//
	// It is a seam rather than a read the repository makes for itself because who belongs to an
	// account is identity's to say, and the tally that needs it is a rule, not a query.
	MealPlanElectorate interface {
		MembersOfAccount(ctx context.Context, accountID string) ([]string, error)
	}

	// MealPlanEventDecision is the option a tally chose for one event.
	MealPlanEventDecision struct {
		_ struct{} `json:"-"`

		MealPlanEventID  string `json:"mealPlanEventID"`
		MealPlanOptionID string `json:"mealPlanOptionID"`
		Tiebroken        bool   `json:"tiebroken"`
	}

	// MealPlanTally is what counting a meal plan's ballots decided. Writing it down is the
	// repository's job; deciding it is TallyMealPlan's.
	MealPlanTally struct {
		_ struct{} `json:"-"`

		// Decisions holds the option chosen for each event this tally settled. An event that
		// was already settled, offers no options, or is still waiting on a ballot is absent.
		Decisions []*MealPlanEventDecision `json:"decisions"`
		// AwaitingVotesFrom names the members whose ballots held an event open. It is for
		// telling somebody why a plan did not finalize, and decides nothing.
		AwaitingVotesFrom []string `json:"awaitingVotesFrom"`
		// Finalized reports whether the plan is finalized by this tally.
		Finalized bool `json:"finalized"`
	}
)

// RandomTiebreak is the tiebreak production tallies use: one of n equally ranked options, at
// random. Chance is the fairest arbiter a tie has, and nothing about it needs to be unpredictable
// to an adversary.
func RandomTiebreak(n int) int {
	/* #nosec: G404 */
	return rand.IntN(n)
}

// InitialMealPlanStatus is the status a new meal plan starts in. A plan offering a choice on any
// event waits for votes; one whose every event names at most one meal has nothing to vote on, so
// it is finalized from the start.
func InitialMealPlanStatus(events []*MealPlanEventDatabaseCreationInput) MealPlanStatus {
	for _, event := range events {
		if len(event.Options) > 1 {
			return MealPlanStatusAwaitingVotes
		}
	}

	return MealPlanStatusFinalized
}

// TallyMealPlan counts a meal plan's ballots and decides what, if anything, voting settled.
//
// The electorate is every member of the plan's account. Before the voting deadline an event is
// decided only once every member has voted on it, and events are decided in order: the first one
// still waiting on a ballot holds every later one open too. Once the deadline has passed every
// open event is decided on whatever ballots it has. The plan is finalized when the deadline has
// passed or nothing was left waiting.
//
// An event voting has already settled, or one that offers no options, is left alone. An event
// nobody voted on chooses nothing, and once the deadline has passed the plan finalizes around it.
//
// tiebreak picks one of n equally ranked winners. It is a parameter so that the tally is a pure
// function of its inputs; production passes a random choice.
func TallyMealPlan(mealPlan *MealPlan, electorate []string, now time.Time, tiebreak func(n int) int) (*MealPlanTally, error) {
	if mealPlan == nil || tiebreak == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if strings.EqualFold(mealPlan.Status, string(MealPlanStatusFinalized)) {
		return nil, ErrAlreadyFinalized
	}

	deadlinePassed := mealPlan.VotingDeadline.Before(now)
	tally := &MealPlanTally{
		Decisions:         []*MealPlanEventDecision{},
		AwaitingVotesFrom: []string{},
	}

	waiting := false
	for _, event := range mealPlan.Events {
		if len(event.Options) == 0 || eventIsDecided(event) {
			continue
		}

		missing := membersWithoutBallots(event, electorate)
		tally.AwaitingVotesFrom = append(tally.AwaitingVotesFrom, missing...)
		if len(missing) > 0 {
			waiting = true
		}

		if waiting && !deadlinePassed {
			continue
		}

		winner, tiebroken, chosen, err := decideEventWinner(event.Options, tiebreak)
		if err != nil {
			return nil, platformerrors.Wrapf(err, "counting the ballots of meal plan event %q", event.ID)
		}

		if chosen {
			tally.Decisions = append(tally.Decisions, &MealPlanEventDecision{
				MealPlanEventID:  event.ID,
				MealPlanOptionID: winner,
				Tiebroken:        tiebroken,
			})
		}
	}

	tally.Finalized = deadlinePassed || !waiting

	return tally, nil
}

// eventIsDecided reports whether voting has already settled an event.
func eventIsDecided(event *MealPlanEvent) bool {
	for _, option := range event.Options {
		if option.Chosen {
			return true
		}
	}

	return false
}

// membersWithoutBallots returns the members of the electorate who have not voted on any of an
// event's options.
func membersWithoutBallots(event *MealPlanEvent, electorate []string) []string {
	voted := map[string]bool{}
	for _, option := range event.Options {
		for _, vote := range option.Votes {
			voted[vote.ByUser] = true
		}
	}

	missing := []string{}
	for _, member := range electorate {
		if !voted[member] {
			missing = append(missing, member)
		}
	}

	return missing
}

// decideEventWinner runs a Schulze election over one event's ballots. It reports false when the
// event has no ballots to count.
func decideEventWinner(options []*MealPlanOption, tiebreak func(n int) int) (winner string, tiebroken, chosen bool, err error) {
	candidateSet := map[string]struct{}{}
	ballots := map[string]schulze.Ballot[string]{}

	for _, option := range options {
		for _, vote := range option.Votes {
			if ballots[vote.ByUser] == nil {
				ballots[vote.ByUser] = schulze.Ballot[string]{}
			}

			if !vote.Abstain {
				ballots[vote.ByUser][vote.BelongsToMealPlanOption] = int(vote.Rank)
			}

			candidateSet[vote.BelongsToMealPlanOption] = struct{}{}
		}
	}

	if len(candidateSet) == 0 {
		return "", false, false, nil
	}

	// Sorted so that the same ballots produce the same results in the same order, which is what
	// lets the tiebreak be the only source of chance.
	candidates := make([]string, 0, len(candidateSet))
	for candidate := range candidateSet {
		candidates = append(candidates, candidate)
	}
	slices.Sort(candidates)

	voting := schulze.NewVoting(candidates)
	for _, ballot := range ballots {
		// Vote refuses a ballot naming a candidate the election does not know, and every
		// candidate came from these ballots, so this is not expected to fail.
		if _, err = voting.Vote(ballot); err != nil {
			return "", false, false, err
		}
	}

	results, _, tie := voting.Compute()
	if tie {
		return breakTie(results, tiebreak), true, true, nil
	}

	if len(results) > 0 {
		return results[0].Choice, false, true, nil
	}

	return "", false, false, nil
}

// breakTie picks among the results with the most wins.
func breakTie(results []schulze.Result[string], tiebreak func(n int) int) string {
	var (
		highest int
		leaders []string
	)

	for _, result := range results {
		switch {
		case result.Wins == highest:
			leaders = append(leaders, result.Choice)
		case result.Wins > highest:
			highest = result.Wins
			leaders = []string{result.Choice}
		}
	}

	return leaders[tiebreak(len(leaders))]
}

// FinalizeMealPlan tallies a meal plan's ballots over its account's electorate and records what
// the tally decided. It is the one procedure behind both a member asking to finalize a plan and
// the finalization saga doing it once the deadline passes, so the two cannot disagree about what
// finalizing means.
//
// It returns ErrAlreadyFinalized for a plan that is already finalized, and otherwise the tally,
// whose Finalized field says whether this call finalized the plan.
func FinalizeMealPlan(
	ctx context.Context,
	plans MealPlanDataManager,
	electorate MealPlanElectorate,
	mealPlanID, accountID string,
	now time.Time,
	tiebreak func(n int) int,
) (*MealPlanTally, error) {
	if mealPlanID == "" || accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	members, err := electorate.MembersOfAccount(ctx, accountID)
	if err != nil {
		return nil, platformerrors.Wrapf(err, "reading the electorate of account %q", accountID)
	}

	mealPlan, err := plans.GetMealPlan(ctx, mealPlanID, accountID)
	if err != nil {
		return nil, platformerrors.Wrapf(err, "reading meal plan %q", mealPlanID)
	}

	tally, err := TallyMealPlan(mealPlan, members, now, tiebreak)
	if err != nil {
		return nil, err
	}

	if err = plans.RecordMealPlanTally(ctx, mealPlan, tally); err != nil {
		return nil, platformerrors.Wrapf(err, "recording the tally of meal plan %q", mealPlanID)
	}

	return tally, nil
}
