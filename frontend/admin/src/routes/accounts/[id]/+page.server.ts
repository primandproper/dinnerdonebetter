import type { PageServerLoad } from './$types';
import { EntryQuery } from '@primandproper/platform-client/audit/v1';
import { getAccount, listAccountMembers, listAuditEntries } from '$lib/grpc/clients';
import { QueryFilter } from '@primandproper/platform-client/filtering/v1';

export const load: PageServerLoad = async ({ locals, params }) => {
  const session = locals.session;
  const accountId = params.id;
  try {
    const accountRes = (await getAccount(session, { accountId })) as { account?: Record<string, unknown> };
    const account = accountRes?.account ?? null;

    let users: unknown[] = [];
    let auditLog: unknown[] = [];

    if (accountId) {
      try {
        const usersRes = (await listAccountMembers(session, {
          accountId,
          filter: QueryFilter.create({ maxResponseSize: 50 }),
        })) as { results?: unknown[] };
        users = usersRes?.results ?? [];
      } catch {
        // ignore
      }
      try {
        const auditRes = (await listAuditEntries(session, {
          query: EntryQuery.create({ resourceType: 'accounts', resourceId: accountId }),
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
