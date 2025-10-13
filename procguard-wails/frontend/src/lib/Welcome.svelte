<script lang="ts">
  export let navigateTo: (view: string) => void;
</script>

<div class="welcome-container">
  <h2>Chào mừng đến với ProcGuard</h2>
  <p>Chọn một mục để bắt đầu:</p>
  <button class="big-button" on:click={() => navigateTo('app-management')}>
    Quản lý Ứng dụng
  </button>
  <button class="big-button" on:click={() => navigateTo('web-management')}>
    Quản lý Web
  </button>
</div>

<style>
  .welcome-container {
    text-align: center;
  }
  .big-button {
    padding: 1rem 2rem;
    font-size: 1.2rem;
    margin: 1rem;
  }
</style>
