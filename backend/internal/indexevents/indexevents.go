/*
Package indexevents says which writes feed which search index.

A write that changes an indexed row owes the index an event. That obligation used to be an
option on the emit call — every repository method that touched an indexed entity passed
WithIndexUpsert or WithIndexDelete by hand, at forty call sites — and an obligation expressed as
an option is one a call site can forget. A repository method that omitted it compiled, reviewed
clean, and left the index stale until the next scheduled rebuild, with no test that would catch
it and no metric that would show it.

So the obligation is no longer a parameter. The rules are registered on the outbox Writer once,
as platform's searchsync side effect, which runs inside every Enqueue and derives the index
events from the data change messages the caller was already sending.

The matching, the refusal of a message with no document ID, and the per-document ordering key
are platform's (searchsync.NewSideEffect). What is this application's is the table: which of its
event types feed which index, and under which key the document's ID travels. And the table is
not written here. Each domain keeps its own rows beside its index names — identity's in
internal/services/identity/indexing, the meal planning domain's in
internal/domain/mealplanning/searchindex — and this package is the list that merges them. Adding
an indexed entity to a domain means adding rows to that domain's table; adding a domain means one
entry in the list below.

An event type absent from the merged table produces no index event, which is right — most of
them should not.
*/
package indexevents

import (
	"slices"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/platform-go/v15/searchsync"
)

// SideEffectName identifies this effect on the Writer. It appears in the error a duplicate or
// nil registration is refused with.
const SideEffectName = "search-index"

// tables is every domain's rows, in the order they are merged. Order carries no meaning:
// platform refuses a rule registered twice, so no row can shadow another.
func tables() [][]searchsync.Rule {
	return [][]searchsync.Rule{
		identityindexing.IndexRules(),
		// Domain: mealplanning
		searchindex.IndexRules(),
	}
}

// Rules is the merged table, as a fresh slice: platform validates and keeps what it is handed,
// and a shared slice would let one caller's mutation change what every other caller registered.
func Rules() []searchsync.Rule {
	return slices.Concat(tables()...)
}

// NewSideEffect builds the outbox side effect that derives this application's index events.
//
// Register it once on the outbox Writer. It then runs inside every Enqueue, on the caller's
// executor, so the index events are written by the same statement as the row change and commit
// with it — an index event cannot outlive a rolled-back write, and a committed write cannot lose
// its index event. The one failure it refuses rather than skips is a tabled event with no
// document ID under the key its rule names; skipping that would put back the failure this
// package removes.
func NewSideEffect() (outbox.SideEffect, error) {
	return searchsync.NewSideEffect(Rules())
}

// RulesFor returns the rows for an event type, in table order. An untabled event type has none.
func RulesFor(eventType string) []searchsync.Rule {
	var matched []searchsync.Rule

	rules := Rules()
	for i := range rules {
		if rules[i].EventType == eventType {
			matched = append(matched, rules[i])
		}
	}

	return matched
}

// EventTypes reports every event type the merged table covers, each once, for tests and for
// anything that wants to assert the set rather than read it.
func EventTypes() []string {
	rules := Rules()

	out := make([]string, 0, len(rules))
	for i := range rules {
		if !slices.Contains(out, rules[i].EventType) {
			out = append(out, rules[i].EventType)
		}
	}

	return out
}
