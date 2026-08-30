import type { PageServerLoad } from './$types';
import { getAccount, getUsersForAccount, getAuditLogEntriesForAccount } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

export const load: PageServerLoad = async ({ locals, params }) => {
  const token = locals.accessToken;
  const accountId = params.id;
  if (!token) {
    return { account: null, users: [], auditLog: [], error: 'Not authenticated' };
  }
  try {
    const accountRes = (await getAccount(token, { accountId })) as { result?: Record<string, unknown> };
    const account = accountRes?.result ?? null;

    let users: unknown[] = [];
    let auditLog: unknown[] = [];

    if (accountId) {
      try {
        const usersRes = (await getUsersForAccount(token, {
          accountId,
          filter: QueryFilter.create({ maxResponseSize: 50 }),
        })) as { results?: unknown[] };
        users = usersRes?.results ?? [];
      } catch {
        // ignore
      }
      try {
        const auditRes = (await getAuditLogEntriesForAccount(token, {
          accountId,
          filter: QueryFilter.create({ maxResponseSize: 20 }),
        })) as { results?: unknown[] };
        auditLog = auditRes?.results ?? [];
      } catch {
        // ignore
      }
    }

    return { account, users, auditLog };
  } catch (e) {
    return {
      account: null,
      users: [],
      auditLog: [],
      error: e instanceof Error ? e.message : 'Failed to load account',
    };
  }
};
