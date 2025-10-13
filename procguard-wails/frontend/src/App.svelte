<script lang="ts">
  import Login from './lib/Login.svelte';
  import Welcome from './lib/Welcome.svelte';
  import AppManagement from './lib/AppManagement.svelte';
  import WebManagement from './lib/WebManagement.svelte';
  import Settings from './lib/Settings.svelte';

  let currentView = 'login'; // Default view

  function navigateTo(view: string) {
    currentView = view;
  }

  // This is a placeholder for the actual login logic
  function handleLogin() {
    // After successful login, navigate to the welcome view
    navigateTo('welcome');
  }

</script>

<nav>
  {#if currentView !== 'login'}
    <button on:click={() => navigateTo('welcome')}>Trang chủ</button>
    <button on:click={() => navigateTo('settings')}>Cài đặt</button>
    <button on:click={() => navigateTo('login')} style="float: right;">Đăng xuất</button>
  {/if}
</nav>

<main>
  {#if currentView === 'login'}
    <Login on:login={handleLogin} />
  {:else if currentView === 'welcome'}
    <Welcome {navigateTo} />
  {:else if currentView === 'app-management'}
    <AppManagement />
  {:else if currentView === 'web-management'}
    <WebManagement />
  {:else if currentView === 'settings'}
    <Settings />
  {/if}
</main>

<style>
  /* Global styles can go here */
main {
    height: 100%;
    width: 100%;
  }
</style>