import { redirect } from '@sveltejs/kit';
import { signOut } from '@primandproper/platform-client';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
  // signOut never rejects, and clears the cookie whether or not the server heard it.
  await signOut(locals.session);
  throw redirect(302, '/login');
};
