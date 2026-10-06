<script lang="ts">
  import { enhance } from '$app/forms';
  import { PageContainer, FormField, Input, Button, Alert } from '@dinnerdonebetter/ui';

  let { data, form } = $props();
</script>

<PageContainer narrow>
  <h1>Change Password</h1>
  {#if data.required}
    <Alert variant="info">You need to choose a new password before you can continue.</Alert>
  {/if}
  {#if form?.error}
    <Alert variant="error">{form.error}</Alert>
  {/if}

  <form method="POST" use:enhance class="password-form">
    <FormField id="current_password" label="Current Password" required>
      <Input
        id="current_password"
        name="current_password"
        type="password"
        autocomplete="current-password"
        required
        dataTestId="change-password-current"
      />
    </FormField>
    <FormField id="new_password" label="New Password" required>
      <Input
        id="new_password"
        name="new_password"
        type="password"
        autocomplete="new-password"
        required
        dataTestId="change-password-new"
      />
    </FormField>
    <FormField id="confirm_password" label="Confirm New Password" required>
      <Input
        id="confirm_password"
        name="confirm_password"
        type="password"
        autocomplete="new-password"
        required
        dataTestId="change-password-confirm"
      />
    </FormField>
    <FormField id="totp_code" label="Authenticator Code (if enabled)">
      <Input
        id="totp_code"
        name="totp_code"
        type="text"
        autocomplete="one-time-code"
        placeholder="Optional"
        dataTestId="change-password-totp"
      />
    </FormField>
    <Button type="submit">Change Password</Button>
  </form>
</PageContainer>

<style>
  .password-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    margin-top: var(--space-lg);
  }
</style>
