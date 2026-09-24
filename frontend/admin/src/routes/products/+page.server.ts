import type { PageServerLoad } from './$types';
import { listProducts } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

export const load: PageServerLoad = async ({ locals }) => {
  const token = locals.accessToken;
  if (!token) {
    return { products: [], error: 'Not authenticated' };
  }
  try {
    const res = (await listProducts(token, { filter: QueryFilter.create({ maxResponseSize: 100 }) })) as {
      results?: Array<{ id?: string; name?: string }>;
    };
    return { products: res?.results ?? [] };
  } catch (e) {
    return {
      products: [],
      error: e instanceof Error ? e.message : 'Failed to load products',
    };
  }
};
