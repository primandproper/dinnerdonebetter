import type { PageServerLoad } from './$types';
import { verifyEmailAddress } from '@primandproper/platform-client';

export const load: PageServerLoad = async ({ url, locals }) => {
  const token = url.searchParams.get('t') ?? '';

  if (!token) {
    return {
      success: false,
      message:
        'This verification link is invalid. Please check your email for the correct link or sign in to request a new one.',
    };
  }

  try {
    const result = await verifyEmailAddress(locals.session, token);
    if (result.kind === 'verified') {
      return {
        success: true,
        message: 'Your email has been verified successfully.',
      };
    }
    return {
      success: false,
      message: 'This verification link is invalid or has expired. Please sign in to request a new verification email.',
    };
  } catch {
    return {
      success: false,
      message: 'This verification link is invalid or has expired. Please sign in to request a new verification email.',
    };
  }
};
