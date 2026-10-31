<script lang="ts">
  import { Heading, Text, Link, Card } from '@primandproper/platform-ui';

  let { data } = $props();

  function formatDate(d: Date | string | undefined): string {
    if (!d) return '-';
    const date = typeof d === 'string' ? new Date(d) : d;
    return date.toISOString().slice(0, 10);
  }
</script>

<Heading level={1}>Subscriptions</Heading>
<p class="subtitle">View subscriptions by account</p>

<form method="get" action="/subscriptions" class="search-form">
  <label for="account_id">Account ID</label>
  <input id="account_id" name="account_id" type="text" placeholder="Enter account ID…" value={data?.accountId ?? ''} />
  <button type="submit">Load</button>
</form>

{#if data?.error}
  <p class="error">{data.error}</p>
{:else if data?.accountId && (!data?.items || data.items.length === 0)}
  <Card>
    <Text>No subscriptions found for this account.</Text>
  </Card>
{:else if data?.items && data.items.length > 0}
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Status</th>
          <th>Product</th>
          <th>External ID</th>
          <th>Period</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each data.items as item ((item as Record<string, unknown>).id)}
          {@const row = item as Record<string, unknown>}
          <tr>
            <td><code>{row.id ?? '-'}</code></td>
            <td>{row.status ?? '-'}</td>
            <td><code class="id">{row.productId ?? '-'}</code></td>
            <td><code class="id">{(row.externalSubscriptionId as string) ?? '-'}</code></td>
            <td
              >{formatDate(row.currentPeriodStart as Date | string)} → {formatDate(
                row.currentPeriodEnd as Date | string,
              )}</td
            >
            <td><Link href="/subscriptions/{row.id}">View</Link></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else}
  <Card>
    <Text>Enter an account ID above to view subscriptions.</Text>
  </Card>
{/if}

<style>
  .subtitle {
    color: var(--pp-color-text-muted);
    margin-bottom: var(--pp-space-lg);
  }
  .error {
    color: var(--pp-color-error, #b91c1c);
  }
  .search-form {
    display: flex;
    align-items: center;
    gap: var(--pp-space-sm);
    margin-bottom: var(--pp-space-lg);
  }
  .search-form input {
    padding: var(--pp-space-xs) var(--pp-space-sm);
    min-width: 16rem;
  }
  .table-wrap {
    overflow-x: auto;
  }
  .data-table {
    width: 100%;
    border-collapse: collapse;
  }
  .data-table th,
  .data-table td {
    padding: var(--pp-space-sm) var(--pp-space-md);
    text-align: left;
    border-bottom: 1px solid var(--pp-color-border);
  }
  .id {
    font-size: 0.9em;
  }
</style>
