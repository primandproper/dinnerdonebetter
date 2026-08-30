import type { PageServerLoad } from './$types';
import { getIssueReports } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

export const load: PageServerLoad = async ({ locals }) => {
  const token = locals.accessToken;
  if (!token) {
    return { reports: [], error: 'Not authenticated' };
  }
  try {
    const res = (await getIssueReports(token, { filter: QueryFilter.create({ maxResponseSize: 100 }) })) as {
      results?: Array<{ id?: string }>;
    };
    return { reports: res?.results ?? [] };
  } catch (e) {
    return {
      reports: [],
      error: e instanceof Error ? e.message : 'Failed to load issue reports',
    };
  }
};
