<script lang="ts">
  import { enhance } from '$app/forms';
  import { PageContainer, FormField, Input, Button, Alert, Link } from '@dinnerdonebetter/ui';
  import type { User } from '@dinnerdonebetter/api-client/primandproper/platform/identity/v1/identity';

  // Two things this page used to do are gone, and both went with the surface underneath them.
  //
  // The profile photo — preview, upload form and the whole callback that captured the
  // stored path — read `user.avatar` and posted to `?/update-avatar`. platform's User has
  // no avatar: an image is a row in the upload registry and a reference to it, so
  // restoring this means an RPC on the media surface that attaches an object to a user.
  //
  // The birthday input read `user.birthday` and posted to `?/update-details`, which has
  // never read it. The column now lives in this repo's own side table, because the
  // directory has nowhere to put it, and nothing serves it over the wire yet. Restoring
  // it means an RPC over that table.
  //
  // Both were still on the page after the cutover: the photo upload 404'd, and the
  // birthday silently discarded whatever was typed into it.

  let { data } = $props();
  const user = data?.user as User | null | undefined;
  const error = data?.error as string | null | undefined;
  const updated = data?.updated ?? false;

  const errorMessages: Record<string, string> = {
    invalid_username: 'Username is required.',
    invalid_first_name: 'First name is required.',
    invalid_password: 'Password is required to update details.',
    invalid_input: 'Invalid input. Please check your entries.',
    update_failed: 'Failed to save. Please try again.',
    server: 'Something went wrong. Please try again.',
  };
  const displayError = error ? (errorMessages[error] ?? 'Something went wrong.') : null;

  const username = user?.username ?? '';
  const firstName = user?.firstName ?? '';
  const lastName = user?.lastName ?? '';
</script>

<PageContainer>
  <h1>Profile</h1>
  <p><Link href="/account/settings" class="back-link">Back to Account Settings</Link></p>

  {#if updated}
    <Alert variant="info">Profile updated.</Alert>
  {/if}
  {#if displayError}
    <Alert variant="error">{displayError}</Alert>
  {/if}

  {#if user === null && error === 'server'}
    <p>Unable to load profile. Please try again.</p>
  {:else}
    <div class="profile-sections">
      <section class="profile-section">
        <h2>Username</h2>
        <form method="POST" action="?/update-username" use:enhance>
          <FormField id="username" label="Username" required>
            <Input
              id="username"
              name="username"
              type="text"
              autocomplete="username"
              value={username}
              required
              dataTestId="profile-username"
            />
          </FormField>
          <Button type="submit">Update Username</Button>
        </form>
      </section>

      <section class="profile-section">
        <h2>Profile Details</h2>
        <p class="profile-hint">Updating your name requires your password for security.</p>
        <form method="POST" action="?/update-details" use:enhance>
          <FormField id="first_name" label="First Name" required>
            <Input
              id="first_name"
              name="first_name"
              type="text"
              autocomplete="given-name"
              value={firstName}
              required
              dataTestId="profile-first-name"
            />
          </FormField>
          <FormField id="last_name" label="Last Name">
            <Input
              id="last_name"
              name="last_name"
              type="text"
              autocomplete="family-name"
              value={lastName}
              dataTestId="profile-last-name"
            />
          </FormField>
          <FormField id="current_password" label="Current Password" required>
            <Input
              id="current_password"
              name="current_password"
              type="password"
              autocomplete="current-password"
              required
              dataTestId="profile-current-password"
            />
          </FormField>
          <FormField id="totp_token" label="Authenticator Code (if enabled)">
            <Input
              id="totp_token"
              name="totp_token"
              type="text"
              autocomplete="one-time-code"
              placeholder="Optional"
              dataTestId="profile-totp-token"
            />
          </FormField>
          <Button type="submit">Update Details</Button>
        </form>
      </section>
    </div>
  {/if}
</PageContainer>

<style>
  .back-link {
    font-size: 0.875rem;
  }
  .profile-sections {
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
    margin-top: var(--space-lg);
  }
  .profile-section {
    padding: var(--space-md);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }
  .profile-section h2 {
    font-size: 0.875rem;
    font-weight: var(--font-weight-medium);
    margin: 0 0 var(--space-md);
  }
  .profile-hint {
    font-size: 0.875rem;
    color: var(--color-muted, #666);
    margin: 0 0 var(--space-md);
  }
  .profile-section form {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }
</style>
