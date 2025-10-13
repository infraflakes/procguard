<script lang="ts">
  import { onMount } from 'svelte';
import { GetAutostartStatus, EnableAutostart, DisableAutostart } from '../../wailsjs/go/main/App';

  let isAutostartEnabled = false;
  let autostartStatusText = 'Không rõ';
  let showUninstallModal = false;
  let uninstallPassword = '';
  let uninstallError = '';

  onMount(async () => {
    await loadAutostartStatus();
  });

  async function loadAutostartStatus() {
    try {
      isAutostartEnabled = await GetAutostartStatus();
      autostartStatusText = isAutostartEnabled ? 'Đã bật' : 'Đã tắt';
    } catch (error) {
      autostartStatusText = 'Không hỗ trợ trên HĐH này';
    }
  }

  async function toggleAutostart() {
    try {
      if (isAutostartEnabled) {
        await DisableAutostart();
      } else {
        await EnableAutostart();
      }
      await loadAutostartStatus();
    } catch (error) {
      alert(`Thao tác thất bại: ${error}`);
    }
  }

  function openUninstallModal() {
    showUninstallModal = true;
  }

  function closeUninstallModal() {
    showUninstallModal = false;
    uninstallPassword = '';
    uninstallError = '';
  }

  async function handleUninstall() {
    // The uninstall logic from the original app is complex and involves
    // killing processes and exiting the app. This is better handled in the Go backend.
    // We will need a dedicated backend function for this.
    // For now, this is a placeholder.
    alert('Uninstall functionality not yet fully implemented in Wails.');
  }
</script>

<div id="settings-view">
  <h2>Cài đặt</h2>
  <div id="settings-controls">
    <div id="autostart-settings">
      <h4>Khởi động cùng Windows</h4>
      <p>Trạng thái: <b>{autostartStatusText}</b></p>
      <button on:click={toggleAutostart}>
        {isAutostartEnabled ? 'Tắt tự động khởi động' : 'Bật tự động khởi động'}
      </button>
    </div>
    <hr />
    <button on:click={openUninstallModal} style="background-color: #d9534f">
      Gỡ cài đặt ProcGuard
    </button>
    <p>
      <b>Cảnh báo:</b> Thao tác này sẽ xóa toàn bộ dữ liệu và gỡ cài đặt
      ProcGuard khỏi hệ thống.
    </p>
  </div>
</div>

{#if showUninstallModal}
  <div class="modal">
    <div class="modal-content">
      <span class="close-button" role="button" tabindex="0" on:click={closeUninstallModal} on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') closeUninstallModal(); }}>&times;</span>
      <h3>Xác nhận gỡ cài đặt</h3>
      <p>Vui lòng nhập mật khẩu của bạn để tiếp tục.</p>
      <form on:submit|preventDefault={handleUninstall}>
        <input type="password" bind:value={uninstallPassword} placeholder="Mật khẩu" required />
        <button type="submit">Xác nhận</button>
      </form>
      {#if uninstallError}
        <p class="error" style="display: block;">{uninstallError}</p>
      {/if}
    </div>
  </div>
{/if}

<style>
  /* Add component-specific styles here */
</style>
