/*
Package privacy is the waitlists domain's contribution to a subject access
request: what a person's waitlist signups say about them, and what an erasure
takes away.

There is an eraser here, unlike most of this application's domains, and the
reason is structural. Every belongs_to_user column in this repository's own
schema cascades from the user row, so one identity eraser covers them; the
platform's waitlist_signups has no such column and cannot have one — see
internal/repositories/postgres/migrations. Without this, erasing a user would
leave their address on every list they had joined.
*/
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	platformwaitlists "github.com/primandproper/platform-go/v14/waitlists"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/pointer"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// NewCollector builds the waitlists collector: every signup belonging to the
// subject, paged to the end and encoded, or nothing if they joined no list.
//
// A withdrawn signup is not among them, and that is the store's answer rather
// than this package's: a withdrawal blanks the subject reference, so the row
// that remembers a suppression no longer says whose it was. What it holds after
// that is a digest of an address and nothing else, which is not data about an
// identifiable person to export.
// The reader is taken at construction because dataprivacy.Collector.Collect is
// handed no executor. The request scope is not consulted: every signup this
// deployment writes is in waitlists.Scope.
func NewCollector(store platformwaitlists.SignupStore, reader database.SQLQueryExecutor) platformdataprivacy.Collector {
	return platformdataprivacy.CollectorFor(func(ctx context.Context, _ tenancy.Scope, subject platformdataprivacy.Subject, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[platformwaitlists.Signup], error) {
		return store.ListSignupsForSubject(ctx, reader, waitlists.Scope(), waitlists.SubjectFor(subject.ID), filter)
	})
}

// NewEraser builds the waitlists eraser: every signup belonging to the subject
// is withdrawn.
//
// # Why withdrawal is the erasure
//
// A withdrawal is exactly what an erasure of a signup has to be. It blanks the
// contact, the notes and the subject reference — everything the row said about a
// person — and keeps the digest of the address, which is what makes the
// suppression outlive the address it is about. Deleting the row instead would
// free the unique key, so the next signup from that address would quietly
// succeed: an account deleted at somebody's request would be an address that can
// be added back to a mailing list. Erasing somebody and then re-subscribing them
// is not an erasure.
//
// What is retained is therefore an unsalted digest of an email address and the
// fact that it is suppressed. That is reported in the outcome rather than left
// implicit, because it is the sort of thing that gets asked about afterwards.
//
// # What platform-go #458 closed
//
// Both of this eraser's departures from the Eraser contract are gone in v14.
// Withdraw takes the caller's database.Tx, so a withdrawal now commits with the
// rest of the subject's erasure or not at all; and ListSignupsForSubject pages
// archived rows when the filter asks, so an administratively archived signup —
// which still holds the address it was made with — is reached like any other.
// Both reads below therefore run on the erasure's own transaction rather than on
// a handle of their own.
func NewEraser(store platformwaitlists.SignupStore) platformdataprivacy.Eraser {
	return platformdataprivacy.EraserFunc(func(ctx context.Context, tx database.Tx, _ tenancy.Scope, subject platformdataprivacy.Subject) (platformdataprivacy.ErasureOutcome, error) {
		outcome := platformdataprivacy.ErasureOutcome{}

		signups, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[platformwaitlists.Signup], error) {
			everything := *filter
			everything.IncludeArchived = pointer.To(true)

			return store.ListSignupsForSubject(ctx, tx, waitlists.Scope(), waitlists.SubjectFor(subject.ID), &everything)
		})
		if err != nil {
			return outcome, platformerrors.Wrap(err, "reading the subject's waitlist signups")
		}

		for i := range signups {
			signup := signups[i]

			if _, err = store.Withdraw(ctx, tx, waitlists.Scope(), signup.ListID, signup.ID); err != nil {
				return outcome, platformerrors.Wrapf(err, "withdrawing waitlist signup %q", signup.ID)
			}

			outcome.Anonymized++
		}

		if outcome.Anonymized > 0 {
			outcome.Retained = map[string]string{
				"waitlist_signups": "an irreversible digest of the address each signup was made with, kept so that " +
					"the withdrawal this erasure performed cannot be undone by a later signup from the same address",
			}
		}

		return outcome, nil
	})
}
