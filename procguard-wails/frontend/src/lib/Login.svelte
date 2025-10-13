<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { HasPassword, Login as wailsLogin, SetPassword } from '../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher();

  let hasPassword = false;
  let password = '';
  let newPassword = '';
  let confirmPassword = '';
  let errorMessage = '';

  onMount(async () => {
    hasPassword = await HasPassword();
  });

  async function handleLogin() {
    const success = await wailsLogin(password);
    if (success) {
      dispatch('login');
    } else {
      errorMessage = 'Sai mật khẩu';
    }
  }

  async function handleSetPassword() {
    if (newPassword !== confirmPassword) {
      errorMessage = 'Mật khẩu không khớp';
      return;
    }
    try {
      await SetPassword(newPassword);
      dispatch('login');
    } catch (error) {
      errorMessage = 'Lỗi đặt mật khẩu';
    }
  }
</script>

<div class="login-container">
  <h1>{hasPassword ? 'Đăng nhập' : 'Tạo mật khẩu'}</h1>
  <form on:submit|preventDefault={hasPassword ? handleLogin : handleSetPassword}>
    {#if hasPassword}
      <div class="form-group">
        <label for="password">Mật khẩu</label>
        <input type="password" id="password" bind:value={password} required />
      </div>
    {:else}
      <div class="form-group">
        <label for="new-password">Mật khẩu mới</label>
        <input type="password" id="new-password" bind:value={newPassword} required />
      </div>
      <div class="form-group">
        <label for="confirm-password">Xác nhận mật khẩu</label>
        <input type="password" id="confirm-password" bind:value={confirmPassword} required />
      </div>
    {/if}
    <button type="submit">Tiếp tục</button>
  </form>
  {#if errorMessage}
    <p class="error" style="display: block;">{errorMessage}</p>
  {/if}
</div>

<style>
  .login-container {
    background-color: #fff;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    width: 300px;
    margin: auto;
  }
  h1 {
    text-align: center;
    color: #333;
  }
  .form-group {
    margin-bottom: 1rem;
  }
  label {
    display: block;
    margin-bottom: 0.5rem;
    color: #555;
  }
  input[type='password'] {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #ccc;
    border-radius: 4px;
    box-sizing: border-box;
  }
  button {
    width: 100%;
    padding: 0.75rem;
    border: none;
    border-radius: 4px;
    background-color: #007bff;
    color: #fff;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
    box-sizing: border-box;
  }
  button:hover {
    background-color: #0056b3;
  }
  .error {
    color: #d93025;
    text-align: center;
    margin-top: 1rem;
  }
</style>
