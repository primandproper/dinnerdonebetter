import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { QueryFilter } from '@dinnerdonebetter/api-client';
import {
  getActiveAccount,
  getSelf,
  getSentAccountInvitations,
  createAccountInvitation,
  cancelAccountInvitation,
  updateAccountMemberPermissions,
} from '$lib/grpc/clients';

const ACCOUNT_ADMIN_ROLE = 'account_admin';
const ACCOUNT_MEMBER_ROLE = 'account_member';

export const load: PageServerLoad = async ({ locals, url, request }) => {
  const token = locals.oauthToken;
  if (!token) {
    return {
      account: null,
      invitations: [],
      currentUserId: '',
      isAdmin: false,
      baseUrl: '',
      error: null,
      invited: false,
    };
  }

  try {
    const activeRes = await getActiveAccount(token);
    const account = activeRes.result ?? null;

    if (!account) {
      return {
        account: null,
        invitations: [],
        currentUserId: '',
        isAdmin: false,
        baseUrl: buildBaseUrl(request),
        error: null,
        invited: false,
      };
    }

    const selfRes = await getSelf(token);
    const currentUserId = selfRes.result?.id ?? '';

    let isAdmin = false;
    for (const m of account.members ?? []) {
      const userId = m.belongsToUser?.id ?? '';
      if (userId === currentUserId && m.accountRole === ACCOUNT_ADMIN_ROLE) {
        isAdmin = true;
        break;
      }
    }

    const invRes = await getSentAccountInvitations(token, {
      filter: QueryFilter.create({ maxResponseSize: 50 }),
    });
    const invitations = (invRes.results ?? []).filter((inv) => inv.destinationAccount?.id === account.id);

    const baseUrl = buildBaseUrl(request);
    const error = url.searchParams.get('error');
    const invited = url.searchParams.get('invited') === '1';

    return {
      account,
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
    const token = locals.oauthToken;
    if (!token) throw redirect(302, '/login');

    const formData = await request.formData();
    const email = (formData.get('email') as string)?.trim() ?? '';
    const name = (formData.get('name') as string)?.trim() ?? '';
    const note = (formData.get('note') as string)?.trim() ?? '';

    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!email || !emailRegex.test(email)) {
      throw redirect(302, '/account/household-members?error=invalid_email');
    }

    try {
      await createAccountInvitation(token, {
        input: { toEmail: email, toName: name, note, expiresAt: undefined },
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
    const token = locals.oauthToken;
    if (!token) throw redirect(302, '/login');

    const formData = await request.formData();
    const invitationId = (formData.get('invitation_id') as string)?.trim() ?? '';
    if (!invitationId) {
      throw redirect(302, '/account/household-members?error=invalid');
    }

    try {
      await cancelAccountInvitation(token, {
        accountInvitationId: invitationId,
        input: { token: '', note: '' },
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
    const token = locals.oauthToken;
    if (!token) throw redirect(302, '/login');

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
      await updateAccountMemberPermissions(token, {
        userId,
        input: { newRole, reason },
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
