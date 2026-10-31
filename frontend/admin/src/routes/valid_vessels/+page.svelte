<script lang="ts">
  import { Heading, Text, Link, Card } from '@primandproper/platform-ui';

  let { data } = $props();
</script>

<Heading level={1}>Valid Vessels</Heading>
<p class="subtitle">List and manage valid vessels</p>

<form method="get" action="/valid_vessels" class="search-form">
  <label for="q">Search</label>
  <input id="q" name="q" type="search" placeholder="Filter by name…" />
  <button type="submit">Search</button>
</form>

{#if data?.error}
  <p class="error">{data.error}</p>
{:else if data?.items && data.items.length > 0}
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Name</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each data.items as item, i ((item as Record<string, unknown>).id ?? i)}
          <tr>
            <td><code>{(item as Record<string, unknown>).id ?? '-'}</code></td>
            <td>{(item as Record<string, unknown>).name ?? '-'}</td>
            <td><Link href="/valid_vessels/{(item as Record<string, unknown>).id}">View</Link></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else}
  <Card>
    <Text>No valid vessels found.</Text>
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
  .search-form {
    display: flex;
    align-items: center;
    gap: var(--pp-space-sm);
    margin-bottom: var(--pp-space-lg);
  }
  .search-form input {
    padding: var(--pp-space-xs) var(--pp-space-sm);
    min-width: 12rem;
  }
</style>
