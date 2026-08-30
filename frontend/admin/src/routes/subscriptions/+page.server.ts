import type { PageServerLoad } from './$types';
import { getSubscriptionsForAccount } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

export const load: PageServerLoad = async ({ locals, url }) => {
  const token = locals.accessToken;
  if (!token) {
    return { items: [], accountId: null, error: 'Not authenticated' };
  }
  const accountId = url.searchParams.get('account_id')?.trim() ?? '';
  if (!accountId) {
    return { items: [], accountId: null, error: null };
  }
  try {
    const res = (await getSubscriptionsForAccount(token, {
      accountId,
      filter: DEFAULT_LIST_FILTER,
    })) as {
      results?: Array<{
        id?: string;
        belongsToAccount?: string;
        productId?: string;
        externalSubscriptionId?: string;
        status?: string;
        currentPeriodStart?: Date | string;
        currentPeriodEnd?: Date | string;
      }>;
    };
    return { items: res?.results ?? [], accountId, error: null };
  } catch (e) {
    return {
      items: [],
      accountId,
      error: e instanceof Error ? e.message : 'Failed to load subscriptions',
    };
  }
};
