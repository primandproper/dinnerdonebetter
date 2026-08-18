package datachangemessagehandler

import (
	"context"

	"github.com/primandproper/platform-go/v11/jobs"
)

// SearchSyncer pairs a search index's topic with the Syncer that applies its events.
//
// One entry per index, because platform-go v10 keys an index event by the topic it arrives on
// rather than by a discriminator inside the payload. The single search-index-requests topic and
// the switch that fanned it out by index type are both gone: a searchsync.Event names a
// document and an operation, and nothing else, so the only thing that can say which index it
// belongs to is where it came from.
//
// Handle is the Syncer's own jobs.Handler. Concurrency, retry with backoff, dead-lettering and
// draining shutdown all belong to the jobs.Pool that runs it, which is why there is so little
// else on this type.
//
// Close is the one thing the Pool cannot own: the Syncer writes through an indexstamp.Stamper,
// whose buffer holds the last interval's worth of last_indexed_at stamps in memory and owns a
// goroutine to flush them. It is closed after the pools drain, so a document indexed by a
// handler that was still running at shutdown is stamped rather than dropped. A nil Close is a
// syncer with nothing to shut down, which is what the tests build.
type SearchSyncer struct {
	Handle jobs.Handler
	Close  func(ctx context.Context) error
	Topic  string
}
