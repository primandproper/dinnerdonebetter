<script lang="ts">
  import { Heading, Button } from '@primandproper/platform-ui';

  let { form } = $props();
</script>

<Heading level={1}>Queue Test</Heading>
<p class="subtitle">Send a test message through a queue and verify round-trip delivery</p>

<form method="POST" action="?/default">
  <label for="queue_name">Queue</label>
  <select id="queue_name" name="queue_name" required>
    <option value="data_changes">Data Changes</option>
    <option value="outbound_emails">Outbound Emails</option>
    <option value="search_index_requests">Search Index Requests</option>
    <option value="user_data_aggregation">User Data Aggregation</option>
    <option value="webhook_execution_requests">Webhook Execution Requests</option>
  </select>
  <Button type="submit">Send test message</Button>
</form>

{#if form?.success !== undefined && form?.success !== null}
  {#if form.success}
    <p class="result success">Success. Test ID: {form.testId ?? '-'}, Round-trip: {form.roundTripMs ?? 0} ms</p>
  {:else}
    <p class="result error">Failed: {form.error ?? 'Unknown error'}</p>
  {/if}
{:else if form?.error}
  <p class="result error">{form.error}</p>
{/if}

<style>
  .subtitle {
    color: var(--pp-color-text-muted);
    margin-bottom: var(--pp-space-lg);
  }
  form {
    display: flex;
    flex-direction: column;
    gap: var(--pp-space-md);
    max-width: 20rem;
  }
  label {
    font-weight: 500;
  }
  select {
    padding: var(--pp-space-sm) var(--pp-space-md);
    border: 1px solid var(--pp-color-border);
    border-radius: var(--pp-radius-md);
  }
  .result {
    margin-top: var(--pp-space-lg);
  }
  .result.success {
    color: var(--pp-color-success, green);
  }
  .result.error {
    color: var(--pp-color-error, #b91c1c);
  }
</style>
