/**
 * The caller's own controls are gated on what GetPrincipal says they may do in the active
 * account, not on the names of the roles they hold: which role grants what is the server's
 * mapping, and a copy of it here would drift from it. The names below are platform's
 * defaults for the methods behind each control (identity/grpc/permissions.go). A control
 * is only a hint either way, since the server still refuses the call.
 */

import type { GetPrincipalResponse } from '@primandproper/platform-client/identity/v1';

export const Permission = {
  /** UpdateAccount: renaming the account and changing its billing address. */
  updateAccount: 'identity.accounts.update',
  /** SetMembershipRoles and RemoveMember. */
  manageMembers: 'identity.members.manage',
  /** Invite and CancelInvitation. */
  inviteMembers: 'identity.invitations.send',
} as const;

export type Permission = (typeof Permission)[keyof typeof Permission];

/**
 * holds reports whether the principal may do `permission` in its active account. A server
 * that states no permissions at all is not one that grants none, which GetPrincipal says a
 * client must not read it as, so its controls are shown and the server decides each call.
 */
export function holds(principal: Pick<GetPrincipalResponse, 'permissions'>, permission: Permission): boolean {
  return principal.permissions?.permissions.includes(permission) ?? true;
}
