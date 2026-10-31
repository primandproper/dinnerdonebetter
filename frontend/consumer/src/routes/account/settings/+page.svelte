<script lang="ts">
  import { PageContainer, Link } from '@primandproper/platform-ui';

  let { data } = $props();
  const hasAccount = data?.hasAccount ?? false;
</script>

<PageContainer>
  <h1>Account Settings</h1>
  <p>Manage your account preferences here.</p>
  <nav class="account-links">
    <Link href="/account/household-members" class="account-link">
      <span class="account-link-title">Household Members</span>
      <span class="account-link-subtitle">Members and invitations</span>
    </Link>
    {#if hasAccount}
      <Link href="/account/household-details" class="account-link">
        <span class="account-link-title">Household Details</span>
        <span class="account-link-subtitle">Edit household details</span>
      </Link>
    {/if}
    <Link href="/account/passkeys" class="account-link">
      <span class="account-link-title">Passkeys</span>
      <span class="account-link-subtitle">Add, view, and remove passkeys</span>
    </Link>
    <Link href="/account/sessions" class="account-link">
      <span class="account-link-title">Sessions</span>
      <span class="account-link-subtitle">View and revoke active sessions</span>
    </Link>
    <Link href="/account/preferences" class="account-link">
      <span class="account-link-title">Preferences</span>
      <span class="account-link-subtitle">Configure preferences</span>
    </Link>
    <Link href="/account/profile" class="account-link">
      <span class="account-link-title">Profile</span>
      <span class="account-link-subtitle">Photo, name, and account details</span>
    </Link>
  </nav>
</PageContainer>

<style>
  .account-links {
    display: flex;
    flex-direction: column;
    gap: var(--pp-space-md);
    margin-top: var(--pp-space-lg);
  }
  .account-link {
    display: block;
    padding: var(--pp-space-md);
    border: 1px solid var(--pp-color-border);
    border-radius: var(--pp-radius-md);
    background: var(--pp-color-surface);
    text-decoration: none;
    color: inherit;
  }
  .account-link:hover {
    background: var(--color-surface-hover, #f5f5f5);
  }
  .account-link-title {
    display: block;
    font-weight: var(--pp-font-weight-medium);
  }
  .account-link-subtitle {
    display: block;
    font-size: 0.875rem;
    color: var(--color-muted, #666);
    margin-top: 0.25rem;
  }
</style>
