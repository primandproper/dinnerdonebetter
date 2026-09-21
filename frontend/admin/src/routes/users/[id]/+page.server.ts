import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { QueryFilter } from '@dinnerdonebetter/api-client';
import { getUser, listAccountsForUser, getAuditLogEntriesForUser } from '$lib/grpc/clients';

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
    const userRes = (await getUser(token, { userId })) as { user?: Record<string, unknown> };
    const user = userRes?.user ?? null;

    let accounts: unknown[] = [];
    let auditLog: unknown[] = [];
    const subscriptions: unknown[] = [];

    if (user?.id) {
      try {
        const accRes = (await listAccountsForUser(token, {
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

// The two "require a password change" actions are gone with the RPC they called.
//
// platform's identity service exposes three operator writes — an account status, a set of
// service roles, an archival — and this is not one of them, although the Service and the
// Store both have SetUserRequiresPasswordChange behind them. Restoring the button means an
// RPC upstream rather than a second directory here; see docs/platform-go-v14-adoption.md.
