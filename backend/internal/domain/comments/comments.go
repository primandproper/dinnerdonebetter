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

# Tenancy

Every comment is filed under tenancy.Global(), and that is a decision rather
than a default. A comment is about a recipe, a meal, a meal plan or an issue
report, and the first of those is readable across accounts — so scoping
comments by account would make one recipe's discussion look different depending
on who was reading it, which is not what a discussion is. Household-private
targets are protected by the checks the owning service already runs before it
delegates here, not by the column.
*/
package comments
