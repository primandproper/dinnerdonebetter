import type { PageServerLoad } from './$types';
import { getSelf, getActiveAccount } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, url }) => {
  const session = locals.session;
  const [self, account] = await Promise.all([getSelf(session), getActiveAccount(session)]);

  const updated = url.searchParams.get('updated') === '1';
  const invited = url.searchParams.get('invited') === '1';
  const deleted = url.searchParams.get('deleted') === '1';
  const error = url.searchParams.get('error');

  let flash: { type: 'info' | 'error'; message: string } | null = null;
  if (updated) flash = { type: 'info', message: 'Profile updated.' };
  else if (invited) flash = { type: 'info', message: 'Invitation sent successfully.' };
  else if (deleted) flash = { type: 'info', message: 'Passkey removed successfully.' };
  else if (error) flash = { type: 'error', message: error };

  return {
    user: self ?? null,
    account: account ?? null,
    flash,
  };
};
