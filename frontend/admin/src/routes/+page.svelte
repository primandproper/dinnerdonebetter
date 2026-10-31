<script lang="ts">
  import { Heading, Card } from '@primandproper/platform-ui';

  let { data } = $props();
</script>

<Heading level={1}>Dashboard</Heading>
<p class="dashboard-subtitle">Welcome to the admin dashboard</p>

<div class="stats-grid">
  <Card>
    <div class="stat-label">Users</div>
    <div class="stat-value">{data?.userCount ?? '-'}</div>
  </Card>
  <Card>
    <div class="stat-label">Accounts</div>
    <div class="stat-value">{data?.accountCount ?? '-'}</div>
  </Card>
  <Card>
    <div class="stat-label">Recipes</div>
    <div class="stat-value">{data?.recipeCount ?? '-'}</div>
  </Card>
</div>

<style>
  .dashboard-subtitle {
    color: var(--pp-color-text-muted);
    margin-bottom: var(--pp-space-lg);
  }
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
    gap: var(--pp-space-lg);
  }
  .stat-label {
    font-size: var(--pp-font-size-sm);
    color: var(--pp-color-text-muted);
    margin-bottom: var(--pp-space-xs);
  }
  .stat-value {
    font-size: 1.5rem;
    font-weight: 600;
  }
</style>
