<script lang="ts">
  import { enhance, applyAction } from '$app/forms';
  import { goto, invalidateAll } from '$app/navigation';
  import { env as publicEnv } from '$env/dynamic/public';
  import { PageContainer, FormField, Input, Button, Alert, Link } from '@dinnerdonebetter/ui';
  import { buildClientSideLogger } from '@dinnerdonebetter/logger/browser';
  import type { User } from '@dinnerdonebetter/api-client/identity/identity_messages';

  const logger = buildClientSideLogger('profile');

  let { data, form } = $props();
  const user = data?.user as User | null | undefined;
  const error = data?.error as string | null | undefined;
  const updated = data?.updated ?? (form as { updated?: boolean } | undefined)?.updated ?? false;
  // Prefer public env (available on client); fallback to server-passed value
  const avatarMediaBaseUrl = publicEnv.PUBLIC_AVATAR_MEDIA_URL_PREFIX ?? data?.avatarMediaBaseUrl ?? '';

  const errorMessages: Record<string, string> = {
    invalid_username: 'Username is required.',
    invalid_first_name: 'First name is required.',
    invalid_password: 'Password is required to update details.',
    invalid_input: 'Invalid input. Please check your entries.',
    update_failed: 'Failed to save. Please try again.',
    avatar_upload_failed: 'Failed to upload profile photo. Use a PNG, JPEG, or GIF under 5 MB.',
    server: 'Something went wrong. Please try again.',
  };
  const displayError = error ? (errorMessages[error] ?? 'Something went wrong.') : null;

  const username = user?.username ?? '';
  const firstName = user?.firstName ?? '';
  const lastName = user?.lastName ?? '';
  const birthdayStr = user?.birthday ? new Date(user.birthday).toISOString().slice(0, 10) : '';

  let avatarImageError = $state(false);
  // Hold the latest path returned from upload so the avatar updates even if form prop doesn't propagate
  let latestAvatarPath = $state<string | null>(null);

  // Use path from upload result (latestAvatarPath), form, or loaded user — $derived so they update when latestAvatarPath changes
  const avatarStoragePath = $derived(
    latestAvatarPath ??
      (form as { avatarStoragePath?: string } | undefined)?.avatarStoragePath ??
      user?.avatar?.storagePath ??
      (user?.avatar as { storage_path?: string } | undefined)?.storage_path ??
      '',
  );
  const avatarUrl = $derived(
    avatarMediaBaseUrl && avatarStoragePath ? `${avatarMediaBaseUrl.replace(/\/$/, '')}/${avatarStoragePath}` : null,
  );

  const initials =
    user?.firstName && user?.lastName
      ? `${user.firstName[0]}${user.lastName[0]}`.toUpperCase()
      : (user?.username?.[0]?.toUpperCase() ?? '?');

  function onAvatarImgError() {
    avatarImageError = true;
  }
  $effect(() => {
    if ((form as { avatarStoragePath?: string } | undefined)?.avatarStoragePath) {
      avatarImageError = false;
    }
  });

  function handleAvatarFormSubmit() {
    return async (opts: { result: { type: string; location?: string; data?: unknown } }) => {
      const { result } = opts;
      logger.with({ type: result?.type, data: (result as { data?: unknown })?.data }).debug('avatar form callback ran');
      if (result.type === 'redirect' && (result as { location?: string }).location) {
        // Server redirect URL; no-app-paths resolve() available for dynamic redirect
        // eslint-disable-next-line svelte/no-navigation-without-resolve -- redirect from form action
        await goto((result as { location: string }).location, { invalidateAll: true });
        return;
      }
      // Capture new path from result so we can show it immediately (form prop may not update in time)
      if (
        result.type === 'success' &&
        result.data &&
        typeof result.data === 'object' &&
        'avatarStoragePath' in result.data
      ) {
        const path = (result.data as { avatarStoragePath: string }).avatarStoragePath;
        logger.with({ path }).debug('captured latest avatar path');
        latestAvatarPath = path;
        avatarImageError = false;
      }
      await applyAction(opts.result as Parameters<typeof applyAction>[0]);
      if (result.type === 'success') {
        await invalidateAll();
      }
    };
  }
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
        <p class="profile-hint">Updating your name or birthday requires your password for security.</p>
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
          <FormField id="birthday" label="Birthday">
            <Input id="birthday" name="birthday" type="date" value={birthdayStr} dataTestId="profile-birthday" />
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

      <section class="profile-section">
        <h2>Profile Photo</h2>
        <div class="profile-avatar-row">
          <div class="profile-avatar-preview" aria-hidden="true">
            {#if avatarUrl && !avatarImageError}
              <img src={avatarUrl} alt="" class="profile-avatar-img" onerror={onAvatarImgError} />
            {:else}
              <span class="profile-avatar-initials">{initials}</span>
            {/if}
          </div>
          <form
            id="avatar-form"
            method="POST"
            action="?/update-avatar"
            enctype="multipart/form-data"
            use:enhance={handleAvatarFormSubmit}
            class="profile-avatar-form"
          >
            <FormField id="avatar" label="Choose a photo (PNG, JPEG, or GIF, max 5 MB)">
              <input
                id="avatar"
                name="avatar"
                type="file"
                accept="image/png,image/jpeg,image/gif"
                data-testid="profile-avatar-input"
                class="profile-avatar-file-input"
                onchange={(e) => {
                  const input = e.currentTarget as HTMLInputElement;
                  const formEl = document.getElementById('avatar-form') as HTMLFormElement;
                  logger.with({ hasForm: !!formEl, fileCount: input.files?.length }).debug('avatar file selected');
                  if (formEl && input.files?.length) {
                    formEl.requestSubmit();
                  }
                }}
              />
            </FormField>
          </form>
        </div>
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
  .profile-avatar-row {
    display: flex;
    align-items: flex-start;
    gap: var(--space-lg);
    flex-wrap: wrap;
  }
  .profile-avatar-preview {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    overflow: hidden;
    background: var(--color-surface-muted, #eee);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .profile-avatar-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .profile-avatar-initials {
    font-size: 1.5rem;
    font-weight: var(--font-weight-medium);
    color: var(--color-muted, #666);
  }
  .profile-avatar-form {
    flex: 1;
    min-width: 200px;
  }
  /* Hide the selected filename so only the browse button is visible */
  .profile-avatar-file-input {
    color: transparent;
    width: min-content;
  }
  .profile-avatar-file-input::file-selector-button,
  .profile-avatar-file-input::-webkit-file-upload-button {
    color: initial;
  }
</style>
