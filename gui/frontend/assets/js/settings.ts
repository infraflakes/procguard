let isAutostartEnabled = false;

async function loadAutostartStatus(): Promise<void> {
  const autostartStatusEl = document.getElementById(
    'autostart-status'
  ) as HTMLSpanElement;
  const autostartToggleBtn = document.getElementById(
    'autostart-toggle-btn'
  ) as HTMLButtonElement;
  try {
    const res = await fetch('/api/settings/autostart/status');
    if (!res.ok) {
      autostartStatusEl.textContent = 'Không hỗ trợ trên HĐH này';
      autostartToggleBtn.disabled = true;
      return;
    }
    const data = await res.json();
    isAutostartEnabled = data.enabled;
    autostartStatusEl.textContent = isAutostartEnabled ? 'Đã bật' : 'Đã tắt';
    autostartToggleBtn.textContent = isAutostartEnabled
      ? 'Tắt tự động khởi động'
      : 'Bật tự động khởi động';
    autostartToggleBtn.disabled = false;
  } catch (e) {
    autostartStatusEl.textContent = 'Lỗi khi tải trạng thái';
    autostartToggleBtn.disabled = true;
  }
}

async function toggleAutostart(): Promise<void> {
  const autostartToggleBtn = document.getElementById(
    'autostart-toggle-btn'
  ) as HTMLButtonElement;
  const endpoint = isAutostartEnabled
    ? '/api/settings/autostart/disable'
    : '/api/settings/autostart/enable';
  try {
    autostartToggleBtn.disabled = true;
    const res = await fetch(endpoint, { method: 'POST' });
    if (!res.ok) {
      const errorText = await res.text();
      alert(`Thao tác thất bại: ${errorText}`);
    } else {
      await loadAutostartStatus(); // Refresh status after action
    }
  } catch (e) {
    if (e instanceof Error) {
      alert(`Đã xảy ra lỗi: ${e.message}`);
    }
  } finally {
    autostartToggleBtn.disabled = false;
  }
}

async function uninstall(): Promise<void> {
  const uninstallModal = document.getElementById(
    'uninstall-modal'
  ) as HTMLDivElement;
  uninstallModal.style.display = 'block';
}

function closeUninstallModal(): void {
  const uninstallModal = document.getElementById(
    'uninstall-modal'
  ) as HTMLDivElement;
  const uninstallError = document.getElementById(
    'uninstall-error'
  ) as HTMLParagraphElement;
  const uninstallPasswordInput = document.getElementById(
    'uninstall-password'
  ) as HTMLInputElement;
  uninstallModal.style.display = 'none';
  uninstallError.style.display = 'none';
  uninstallPasswordInput.value = '';
}

document.addEventListener('DOMContentLoaded', () => {
  const uninstallForm = document.getElementById('uninstall-form');
  if (uninstallForm) {
    uninstallForm.addEventListener('submit', async (event: Event) => {
      event.preventDefault();
      const uninstallPasswordInput = document.getElementById(
        'uninstall-password'
      ) as HTMLInputElement;
      const uninstallError = document.getElementById(
        'uninstall-error'
      ) as HTMLParagraphElement;
      const password = uninstallPasswordInput.value;

      const response = await fetch('/api/uninstall', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password }),
      });

      if (response.ok) {
        closeUninstallModal();
        alert('ProcGuard đã được gỡ cài đặt. Trình duyệt sẽ đóng.');
        window.close();
      } else {
        uninstallError.textContent =
          'Gỡ cài đặt thất bại. Vui lòng kiểm tra lại mật khẩu.';
        uninstallError.style.display = 'block';
      }
    });
  }
});
