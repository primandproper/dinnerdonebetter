package datachangemessagehandler

import (
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
// draining shutdown all belong to the jobs.Pool that runs it, which is why there is nothing
// else on this type.
type SearchSyncer struct {
	Handle jobs.Handler
	Topic  string
}
