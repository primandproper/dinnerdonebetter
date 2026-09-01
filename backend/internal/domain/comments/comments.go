/*
Package comments is this application's half of platform-go's comments store: the
namespace its table carries, the catalog of things that can be commented on, and
the data change events a write emits.

The store itself is platform-go's. It owns the schema, the paging, the thread
depth, the tenancy column and the erasure, because that half is the same in every
application. What is not the same — and what platform deliberately refuses to
guess at — is which kinds of thing a comment may be about. That catalog is
assembled in internal/build/comments, which is the one layer that may know both
the comments store and the domains whose things are commented on.
*/
package comments

import (
	"github.com/primandproper/platform-go/v13/tenancy"
)

// TablePrefix namespaces the platform-go comments table, rendering ddb_comments.
//
// The platform's own default is the empty prefix, which renders "comments" — the
// exact name the table this replaced carried, and a name generic enough that a
// database shared with anything else would eventually collide. Its DDL says
// CREATE TABLE IF NOT EXISTS, so that collision would be a silent no-op followed
// by a store reading columns that are not there.
const TablePrefix = "ddb"

// The data change events a comment write emits. They are declared in the webhook
// event catalog (internal/domain/webhooks/catalog), so a subscriber is already
// able to ask for them.
const (
	// CommentCreatedServiceEventType indicates a comment was created.
	CommentCreatedServiceEventType = "comment_created"
	// CommentUpdatedServiceEventType indicates a comment was updated.
	CommentUpdatedServiceEventType = "comment_updated"
	// CommentArchivedServiceEventType indicates a comment was archived.
	CommentArchivedServiceEventType = "comment_archived"
)

// Scope is the tenancy every comment in this deployment is filed under.
//
// It is global, and that is a decision rather than a default. A comment is about
// a recipe, a meal, a meal plan or an issue report, and the first of those is
// readable across accounts — so scoping comments by account would make one
// recipe's discussion look different depending on who was reading it, which is
// not what a discussion is. Household-private targets are protected by the
// checks the owning service already runs before it delegates here, not by the
// column.
func Scope() tenancy.Scope { return tenancy.Global() }
