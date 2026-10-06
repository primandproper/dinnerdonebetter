<script lang="ts">
  import { Heading, Text, Link, Card } from '@dinnerdonebetter/ui';

  let { data } = $props();
</script>

<Heading level={1}>Issue Reports</Heading>
<p class="subtitle">Every account's issue reports</p>

{#if data?.error}
  <p class="error">{data.error}</p>
{:else if data?.reports && data.reports.length > 0}
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Account</th>
          <th>Kind</th>
          <th>Status</th>
          <th>Subject</th>
          <th>Filed</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each data.reports as report (report.id)}
          <tr>
            <td><code>{report.id}</code></td>
            <td><code>{report.accountID || '-'}</code></td>
            <td>{report.kind || '-'}</td>
            <td>{report.status}</td>
            <td>{report.subject || '-'}</td>
            <td>{report.createdAt ? new Date(report.createdAt).toLocaleString() : '-'}</td>
            <td><Link href="/issue_reports/{report.id}">View</Link></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else}
  <Card>
    <Text>No issue reports found.</Text>
  </Card>
{/if}

<style>
  .subtitle {
    color: var(--color-text-muted);
    margin-bottom: var(--space-lg);
  }
  .error {
    color: var(--color-error, #b91c1c);
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
    padding: var(--space-sm) var(--space-md);
    text-align: left;
    border-bottom: 1px solid var(--color-border);
  }
</style>
