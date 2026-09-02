<script lang="ts">
  import { enhance } from '$app/forms';
  import { PageContainer, FormField, Button, Alert, Link } from '@dinnerdonebetter/ui';
  import type { ConfigurableSetting } from '$lib/settings/resolutions';

  let { data } = $props();
  const configurableSettings = (data?.configurableSettings ?? []) as ConfigurableSetting[];
  const error = data?.error as string | null | undefined;
  const updated = data?.updated ?? false;

  const errorMessages: Record<string, string> = {
    invalid: 'Invalid input. Please try again.',
    update_failed: 'Failed to save preference. Please try again.',
    server: 'Something went wrong. Please try again.',
  };
  const displayError = error ? (errorMessages[error] ?? 'Something went wrong.') : null;

  function humanReadable(str: string): string {
    if (!str) return str;
    return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
  }

  function humanReadableName(name: string): string {
    return name.replace(/_/g, ' ');
  }
</script>

<PageContainer>
  <h1>Preferences</h1>
  <p><Link href="/account/settings" class="back-link">Back to Account Settings</Link></p>

  {#if updated}
    <Alert variant="info">Preferences updated.</Alert>
  {/if}
  {#if displayError}
    <Alert variant="error">{displayError}</Alert>
  {/if}

  {#if configurableSettings.length === 0 && !displayError}
    <p class="muted">No preferences to configure.</p>
  {:else}
    <div class="settings-list">
      {#each configurableSettings as item (item.setting.id)}
        {#if item.setting.enumeration?.length}
          <div class="setting-card">
            <form method="POST" action="?/update" use:enhance class="setting-form">
              <input
                type="hidden"
                name="setting_name"
                value={item.setting.name}
                data-testid="preference-setting-name"
              />
              <FormField
                id="pref-{item.setting.id}"
                label={humanReadableName(item.setting.name) +
                  (item.setting.description ? ` — ${item.setting.description}` : '')}
              >
                <select
                  name="value"
                  id="pref-{item.setting.id}"
                  class="pref-select"
                  data-testid="preference-value-{item.setting.id}"
                  onchange={(e) => (e.currentTarget as HTMLSelectElement).form?.requestSubmit()}
                >
                  {#each item.setting.enumeration ?? [] as opt (opt)}
                    <option value={opt} selected={opt === item.currentValue}>
                      {humanReadable(opt)}
                    </option>
                  {/each}
                </select>
              </FormField>
              <Button type="submit" variant="default">Save</Button>
            </form>
          </div>
        {/if}
      {/each}
    </div>
  {/if}
</PageContainer>

<style>
  .back-link {
    font-size: 0.875rem;
  }
  .muted {
    color: var(--color-muted, #666);
    text-align: center;
    padding: var(--space-xl);
  }
  .settings-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    margin-top: var(--space-lg);
  }
  .setting-card {
    padding: var(--space-md);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }
  .setting-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }
  @media (min-width: 640px) {
    .setting-form {
      flex-direction: row;
      align-items: center;
      justify-content: space-between;
    }
  }
  .pref-select {
    padding: var(--space-sm) var(--space-md);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    font-size: 0.875rem;
    min-width: 8rem;
  }
</style>
