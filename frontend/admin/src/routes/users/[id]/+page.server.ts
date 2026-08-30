import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { QueryFilter } from '@dinnerdonebetter/api-client';
import {
  getUser,
  getAccountsForUser,
  getAuditLogEntriesForUser,
  adminSetPasswordChangeRequired,
} from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params, url }) => {
  const token = locals.accessToken;
  const userId = params.id;
  if (!token) {
    return {
      user: null,
      accounts: [],
      auditLog: [],
      subscriptions: [],
      error: 'Not authenticated',
      passwordChangeUpdated: false,
    };
  }
  try {
    const userRes = (await getUser(token, { userId })) as { result?: Record<string, unknown> };
    const user = userRes?.result ?? null;

    let accounts: unknown[] = [];
    let auditLog: unknown[] = [];
    const subscriptions: unknown[] = [];

    if (user?.id) {
      try {
        const accRes = (await getAccountsForUser(token, {
          userId: user.id as string,
          filter: QueryFilter.create({ maxResponseSize: 50 }),
        })) as { results?: unknown[] };
        accounts = accRes?.results ?? [];
      } catch {
        // ignore
      }
      try {
        const auditRes = (await getAuditLogEntriesForUser(token, {
          userId: user.id as string,
          filter: QueryFilter.create({ maxResponseSize: 20 }),
        })) as { results?: unknown[] };
        auditLog = auditRes?.results ?? [];
      } catch {
        // ignore
      }
      // Subscriptions are per-account; we could load for each account or show a message
    }

    const passwordChangeUpdated = url.searchParams.get('password_change_updated') === '1';
    const error = url.searchParams.get('error') ?? null;

    return { user, accounts, auditLog, subscriptions, passwordChangeUpdated, error };
  } catch (e) {
    return {
      user: null,
      accounts: [],
      auditLog: [],
      subscriptions: [],
      error: e instanceof Error ? e.message : 'Failed to load user',
      passwordChangeUpdated: false,
    };
  }
};

export const actions: Actions = {
  'require-password-change': async ({ locals, params }) => {
    const token = locals.accessToken;
    if (!token) throw redirect(302, '/login');

    const userId = params.id;

    try {
      await adminSetPasswordChangeRequired(token, { targetUserId: userId, requiresPasswordChange: true });
      throw redirect(302, `/users/${userId}?password_change_updated=1`);
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, `/users/${userId}?error=password_change_failed`);
    }
  },

  'clear-password-change': async ({ locals, params }) => {
    const token = locals.accessToken;
    if (!token) throw redirect(302, '/login');

    const userId = params.id;

    try {
      await adminSetPasswordChangeRequired(token, { targetUserId: userId, requiresPasswordChange: false });
      throw redirect(302, `/users/${userId}?password_change_updated=1`);
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, `/users/${userId}?error=password_change_failed`);
    }
  },
};
