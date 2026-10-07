<script lang="ts">
  import '../app.css';
  import { page } from '$app/state';
  import { Link, PageContainer } from '@dinnerdonebetter/ui';

  let { children } = $props();

  let mobileMenuOpen = $state(false);

  /** Routes that use their own width (e.g. recipe creator with PageContainer wide) — don't wrap in layout PageContainer */
  const fullWidthRoutes = ['/recipes/new'];
  const useLayoutContainer = $derived(!fullWidthRoutes.some((p) => page.url.pathname === p));
</script>

<div class="layout-root">
  <header class="layout-header">
    <div class="container">
      <div><Link href="/">Admin</Link></div>
      <nav class="layout-links" class:open={mobileMenuOpen}>
        <Link href="/">Dashboard</Link>
        <Link href="/users">Users</Link>
        <Link href="/accounts">Accounts</Link>
        <Link href="/oauth2_clients">OAuth2 Clients</Link>
        <Link href="/recipes">Recipes</Link>
        <Link href="/recipes/new">Create Recipe</Link>
        <Link href="/waitlists">Waitlists</Link>
        <Link href="/issue_reports">Issue Reports</Link>
        <Link href="/settings">Settings</Link>
        <Link href="/products">Products</Link>
        <Link href="/subscriptions">Subscriptions</Link>
        <Link href="/queue_test">Queue Test</Link>
        <div class="nav-dropdown">
          <span class="nav-dropdown-label">Enumerations</span>
          <div class="nav-dropdown-content">
            <Link href="/valid_ingredients">Ingredients</Link>
            <Link href="/valid_instruments">Instruments</Link>
            <Link href="/valid_preparations">Preparations</Link>
            <Link href="/valid_measurement_units">Measurement Units</Link>
            <Link href="/valid_vessels">Vessels</Link>
            <Link href="/valid_ingredient_states">Ingredient States</Link>
            <Link href="/valid_prep_task_configs">Prep Task Configs</Link>
            <Link href="/measurement_unit_conversion_mismatches">Conversion Mismatches</Link>
          </div>
        </div>
        <Link href="/logout">Logout</Link>
      </nav>
      <button
        type="button"
        class="mobile-menu-btn"
        aria-label="Toggle menu"
        onclick={() => (mobileMenuOpen = !mobileMenuOpen)}
      >
        <span></span>
        <span></span>
        <span></span>
      </button>
    </div>
  </header>

  <main class="layout-main">
    {#if useLayoutContainer}
      <PageContainer>
        {@render children()}
      </PageContainer>
    {:else}
      {@render children()}
    {/if}
  </main>
</div>

<style>
  .layout-root {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }
  .layout-header {
    background: var(--color-surface);
    border-bottom: 1px solid var(--color-border);
    padding: var(--space-md);
  }
  .container {
    max-width: 64rem;
    margin: 0 auto;
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
  }
  .layout-links {
    display: flex;
    gap: var(--space-md);
    align-items: center;
  }
  .nav-dropdown {
    position: relative;
  }
  .nav-dropdown-label {
    color: var(--color-text);
    cursor: pointer;
    padding: var(--space-xs) 0;
  }
  .nav-dropdown-content {
    display: none;
    position: absolute;
    top: 100%;
    left: 0;
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: var(--space-sm);
    min-width: 12rem;
    box-shadow: var(--shadow-md);
    z-index: 50;
    flex-direction: column;
    gap: var(--space-xs);
  }
  .nav-dropdown:hover .nav-dropdown-content {
    display: flex;
  }
  .mobile-menu-btn {
    display: none;
    flex-direction: column;
    gap: 4px;
    background: none;
    border: none;
    cursor: pointer;
    padding: var(--space-sm);
  }
  .mobile-menu-btn span {
    width: 1.25rem;
    height: 2px;
    background: var(--color-text);
  }
  @media (max-width: 768px) {
    .mobile-menu-btn {
      display: flex;
    }
    .layout-links {
      display: none;
      width: 100%;
      flex-direction: column;
      align-items: flex-start;
      padding-top: var(--space-md);
    }
    .layout-links.open {
      display: flex;
    }
    .nav-dropdown-content {
      position: static;
      box-shadow: none;
      border: none;
      padding-left: var(--space-md);
    }
    .nav-dropdown:hover .nav-dropdown-content {
      display: none;
    }
  }
  .layout-main {
    flex: 1;
    padding: var(--space-lg);
  }
</style>
