<script lang="ts">
  import '../app.css';
  import favicon from '$lib/assets/favicon.svg';
  import { Link } from '@primandproper/platform-ui';
  import { COMPANY_NAME } from '@dinnerdonebetter/branding';

  let { children } = $props();
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

<div class="layout-root">
  <header class="layout-header">
    <div class="container">
      <div><Link href="/">{COMPANY_NAME}</Link></div>
      <div class="layout-links">
        <Link href="/account/settings">Account</Link>
        <Link href="/logout">Sign Out</Link>
      </div>
    </div>
  </header>

  <main class="layout-main">
    {@render children()}
  </main>

  <footer class="layout-footer">
    <p>© {new Date().getFullYear()} {COMPANY_NAME}. All rights reserved.</p>
    <div class="layout-footer-links">
      <Link href="/privacy-policy">Privacy Policy</Link>
      <Link href="/terms-of-service">Terms of Service</Link>
    </div>
  </footer>
</div>

<style>
  .layout-root {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }
  .layout-header {
    position: sticky;
    top: 0;
    z-index: 50;
    background: var(--pp-color-surface);
    border-bottom: 1px solid var(--pp-color-border);
    padding: var(--pp-space-md);
  }
  .container {
    max-width: 70vw;
    margin-left: 15vw;
    margin-right: auto;
    display: flex;
    justify-content: space-between;
    align-items: center;
    align-content: space-between;
  }

  @media (max-width: 768px) {
    .container {
      max-width: 100%;
      margin-left: 0;
      margin-right: 0;
    }
  }
  .container > div:first-child {
    font-weight: var(--pp-font-weight-medium);
    font-size: 1.25rem;
  }
  .layout-links {
    display: flex;
    gap: var(--pp-space-md);
  }
  .layout-main {
    flex: 1;
    min-height: 50vh;
  }
  .layout-footer {
    border-top: 1px solid var(--pp-color-border);
    background: var(--pp-color-surface);
    padding: var(--pp-space-md);
    margin-top: auto;
  }
  .layout-footer p,
  .layout-footer-links {
    margin: 0;
    font-size: 0.875rem;
  }
  .layout-footer-links {
    display: flex;
    gap: var(--pp-space-md);
    margin-top: var(--pp-space-sm);
  }
</style>
