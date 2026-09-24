<script lang="ts">
  import { enhance } from '$app/forms';
  import { PageContainer, FormField, Input, Button, Alert, Link } from '@dinnerdonebetter/ui';
  import type { Account } from '@primandproper/platform-client/identity/v1';

  let { data } = $props();
  const account = data?.account as Account | null | undefined;
  const isAdmin = data?.isAdmin ?? false;
  const error = data?.error as string | null | undefined;
  const updated = data?.updated ?? false;

  const errorMessages: Record<string, string> = {
    invalid: 'Invalid input. Please check your entries.',
    invalid_name: 'Household name is required.',
    update_failed: 'Failed to save household details. Please try again.',
    server: 'Something went wrong. Please try again.',
  };
  const displayError = error ? (errorMessages[error] ?? 'Something went wrong.') : null;
</script>

<PageContainer>
  <h1>Household Details</h1>
  <p><Link href="/account/settings" class="back-link">Back to Account Settings</Link></p>

  {#if updated}
    <Alert variant="info">Household details updated.</Alert>
  {/if}
  {#if displayError}
    <Alert variant="error">{displayError}</Alert>
  {/if}

  {#if !account}
    <p>Create or join a household to edit household details.</p>
  {:else if !isAdmin}
    <p>Only household admins can edit household details.</p>
  {:else}
    <form method="POST" action="?/update" use:enhance class="details-form">
      <FormField id="name" label="Household Name" required>
        <Input
          id="name"
          name="name"
          type="text"
          value={account.name ?? ''}
          required
          dataTestId="household-details-name"
        />
      </FormField>
      <FormField id="contact_phone" label="Contact Phone">
        <Input
          id="contact_phone"
          name="contact_phone"
          type="tel"
          value={account.billingAddress?.phone ?? ''}
          dataTestId="household-details-contact-phone"
        />
      </FormField>
      <fieldset class="address-section">
        <legend>Address</legend>
        <FormField id="address_line_1" label="Address Line 1">
          <Input
            id="address_line_1"
            name="address_line_1"
            type="text"
            value={account.billingAddress?.line1 ?? ''}
            dataTestId="household-details-address-line-1"
          />
        </FormField>
        <FormField id="address_line_2" label="Address Line 2">
          <Input
            id="address_line_2"
            name="address_line_2"
            type="text"
            value={account.billingAddress?.line2 ?? ''}
            dataTestId="household-details-address-line-2"
          />
        </FormField>
        <FormField id="city" label="City">
          <Input
            id="city"
            name="city"
            type="text"
            value={account.billingAddress?.city ?? ''}
            dataTestId="household-details-city"
          />
        </FormField>
        <FormField id="state" label="State / Province">
          <Input
            id="state"
            name="state"
            type="text"
            value={account.billingAddress?.state ?? ''}
            dataTestId="household-details-state"
          />
        </FormField>
        <FormField id="zip_code" label="ZIP / Postal Code">
          <Input
            id="zip_code"
            name="zip_code"
            type="text"
            value={account.billingAddress?.postalCode ?? ''}
            dataTestId="household-details-zip-code"
          />
        </FormField>
        <FormField id="country" label="Country">
          <Input
            id="country"
            name="country"
            type="text"
            value={account.billingAddress?.country ?? ''}
            dataTestId="household-details-country"
          />
        </FormField>
      </fieldset>
      <Button type="submit">Save Changes</Button>
    </form>
  {/if}
</PageContainer>

<style>
  .back-link {
    font-size: 0.875rem;
  }
  .details-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
    margin-top: var(--space-lg);
  }
  .address-section {
    padding: var(--space-md);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }
  .address-section legend {
    font-size: 0.875rem;
    font-weight: var(--font-weight-medium);
    margin-bottom: var(--space-md);
  }
  .address-section .form-field {
    margin-bottom: var(--space-md);
  }
</style>
