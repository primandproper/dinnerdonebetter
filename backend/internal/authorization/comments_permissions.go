package authorization

import (
	commentsgrpc "github.com/primandproper/platform-go/v14/comments/grpc"
)

// The comment permissions are platform's, re-exported under the names this
// application's policy already spells.
//
// They are re-exported rather than declared, because the surface that enforces
// them is platform's: commentsgrpc.Permissions() maps each RPC to one of these,
// and a constant declared here with a different string would be a policy that
// grants something no method asks for. Naming them here keeps PlatformPolicy
// reading in one vocabulary.
//
// The strings changed with the adoption — "create.comments" became
// "comments.create" — because platform namespaces the domain first so that two
// composed domains cannot collide on a bare verb. Nothing is deployed, so the
// policy the migrator seeds simply seeds the new strings; see
// migrations/migration_files for the seed and docs/identity.md for the policy.
//
// ModerateCommentsPermission has no local predecessor. It gates the two reads
// that cross a discussion — everything said about a kind of thing — which the
// service this replaced did not expose at all, so it is granted to nobody until
// there is a moderation queue to grant it to.
const (
	// CreateCommentsPermission is a permission.
	CreateCommentsPermission = commentsgrpc.PermissionCreateComments
	// ReadCommentsPermission is a permission.
	ReadCommentsPermission = commentsgrpc.PermissionReadComments
	// UpdateCommentsPermission is a permission.
	UpdateCommentsPermission = commentsgrpc.PermissionUpdateComments
	// ArchiveCommentsPermission is a permission.
	ArchiveCommentsPermission = commentsgrpc.PermissionArchiveComments
	// ModerateCommentsPermission gates reading every comment about a kind of thing.
	ModerateCommentsPermission = commentsgrpc.PermissionModerateComments
)
