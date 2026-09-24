import type { PageServerLoad } from './$types';
import { getIssueReport } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params }) => {
  const token = locals.accessToken;
  const id = params.id;
  if (!token) {
    return { report: null, error: 'Not authenticated' };
  }
  try {
    const res = (await getIssueReport(token, { reportId: id })) as { result?: Record<string, unknown> };
    return { report: res?.result ?? null };
  } catch (e) {
    return {
      report: null,
      error: e instanceof Error ? e.message : 'Failed to load issue report',
    };
  }
};
