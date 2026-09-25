import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { QueryFilter } from '@dinnerdonebetter/api-client';
import {
  getActiveAccount,
  getSelf,
  listAccountMembers,
  listInvitationsFromUser,
  invite,
  cancelInvitation,
  setMembershipRoles,
} from '$lib/grpc/clients';

const ACCOUNT_ADMIN_ROLE = 'account_admin';
const ACCOUNT_MEMBER_ROLE = 'account_member';

export const load: PageServerLoad = async ({ locals, url, request }) => {
  const session = locals.session;
  try {
    const activeRes = await getActiveAccount(session);
    const account = activeRes.result ?? null;

    if (!account) {
      return {
        account: null,
        members: [],
        invitations: [],
        currentUserId: '',
        isAdmin: false,
        baseUrl: buildBaseUrl(request),
        error: null,
        invited: false,
      };
    }

    const selfRes = await getSelf(session);
    const currentUserId = selfRes.result?.id ?? '';

    // The roster is a read of its own now. platform's Account carries no member list —
    // an account with thirty members would otherwise be thirty users on every read of it
    // — so who is in it is a paged read, and a membership carries a set of roles rather
    // than a single one.
    const membersRes = await listAccountMembers(session, {
      accountId: account.id,
      filter: QueryFilter.create({ maxResponseSize: 50 }),
    });
    const members = membersRes.results ?? [];

    let isAdmin = false;
    for (const m of members) {
      if (m.user?.id === currentUserId && (m.membership?.roles ?? []).includes(ACCOUNT_ADMIN_ROLE)) {
        isAdmin = true;
        break;
      }
    }

    const invRes = await listInvitationsFromUser(session, {
      filter: QueryFilter.create({ maxResponseSize: 50 }),
    });
    const invitations = (invRes.results ?? []).filter((inv) => inv.belongsToAccount === account.id);

    const baseUrl = buildBaseUrl(request);
    const error = url.searchParams.get('error');
    const invited = url.searchParams.get('invited') === '1';

    return {
      account,
      members,
      invitations,
      currentUserId,
      isAdmin,
      baseUrl,
      error,
      invited,
    };
  } catch {
    return {
      account: null,
      members: [],
      invitations: [],
      currentUserId: '',
      isAdmin: false,
      baseUrl: buildBaseUrl(request),
      error: 'server',
      invited: false,
    };
  }
};

function buildBaseUrl(request: Request): string {
  const url = new URL(request.url);
  return `${url.protocol}//${url.host}`;
}

const _errorMessages: Record<string, string> = {
  invalid: 'Invalid input. Please check your entries.',
  invalid_email: 'Please enter a valid email address.',
  invalid_role: 'Invalid role selected.',
  invitation_failed: 'Failed to send invitation. Please try again.',
  cancel_failed: 'Failed to cancel invitation.',
  role_update_failed: 'Failed to update member role.',
  server: 'Something went wrong. Please try again.',
};

export const actions: Actions = {
  'send-invitation': async ({ request, locals }) => {
    const session = locals.session;
    const formData = await request.formData();
    const email = (formData.get('email') as string)?.trim() ?? '';
    const name = (formData.get('name') as string)?.trim() ?? '';
    const note = (formData.get('note') as string)?.trim() ?? '';

    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!email || !emailRegex.test(email)) {
      throw redirect(302, '/account/household-members?error=invalid_email');
    }

    try {
      // The account is named on the request rather than taken from the session, and the
      // roles the invitation promises come from here: what somebody was invited to is
      // what they get, and an acceptance cannot ask for more.
      const activeRes = await getActiveAccount(session);
      await invite(session, {
        accountId: activeRes.result?.id ?? '',
        toEmail: email,
        toName: name,
        note,
        roles: [ACCOUNT_MEMBER_ROLE],
        expiresAt: undefined,
      });
      throw redirect(302, '/account/household-members?invited=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/household-members?error=invitation_failed');
    }
  },
  'cancel-invitation': async ({ request, locals }) => {
    const session = locals.session;
    const formData = await request.formData();
    const invitationId = (formData.get('invitation_id') as string)?.trim() ?? '';
    if (!invitationId) {
      throw redirect(302, '/account/household-members?error=invalid');
    }

    try {
      // No token: withdrawing is the sender's act, and the secret half of the link is
      // the recipient's. The sender never holds it — no read returns one.
      await cancelInvitation(session, {
        invitationId,
        statusNote: '',
      });
      throw redirect(302, '/account/household-members');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/household-members?error=cancel_failed');
    }
  },
  'update-role': async ({ request, locals }) => {
    const session = locals.session;
    const formData = await request.formData();
    const userId = (formData.get('user_id') as string)?.trim() ?? '';
    const newRole = (formData.get('new_role') as string)?.trim() ?? '';
    const reason = (formData.get('reason') as string)?.trim() ?? '';

    if (!userId || !newRole || !reason) {
      throw redirect(302, '/account/household-members?error=invalid');
    }
    if (newRole !== ACCOUNT_ADMIN_ROLE && newRole !== ACCOUNT_MEMBER_ROLE) {
      throw redirect(302, '/account/household-members?error=invalid_role');
    }

    try {
      // Roles are replaced rather than merged, and the account is named: a caller adding
      // one reads the membership and writes the union, which is visible here rather than
      // hidden in a setter that could not express a revocation.
      const activeRes = await getActiveAccount(session);
      await setMembershipRoles(session, {
        accountId: activeRes.result?.id ?? '',
        userId,
        roles: [newRole],
      });
      throw redirect(302, '/account/household-members');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/household-members?error=role_update_failed');
    }
  },
};
