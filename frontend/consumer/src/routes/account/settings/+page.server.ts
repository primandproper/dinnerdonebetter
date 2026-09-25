import type { PageServerLoad } from './$types';
import { getSelf, listAccountsForUser } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

export const load: PageServerLoad = async ({ locals }) => {
  const session = locals.session;
  try {
    const selfRes = await getSelf(session);
    const user = selfRes.result;
    const userId = user?.id ?? '';

    if (!userId) {
      return { hasAccount: false };
    }

    const accountsRes = await listAccountsForUser(session, {
      userId,
      filter: QueryFilter.create({ maxResponseSize: 1 }),
    });
    const hasAccount = (accountsRes.results?.length ?? 0) > 0;
    return { hasAccount };
  } catch {
    return { hasAccount: false };
  }
};
