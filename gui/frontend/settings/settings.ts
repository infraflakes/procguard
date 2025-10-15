let isAutostartEnabled = false;

// This variable will hold the Bootstrap Modal instance
let uninstallModalInstance: any; // Use 'any' to avoid needing full Bootstrap types

document.addEventListener('DOMContentLoaded', () => {
  // Initialize the modal instance once the DOM is ready
  const uninstallModalEl = document.getElementById('uninstall-modal');
  if (uninstallModalEl) {
    // Assuming Bootstrap's JS is loaded and available globally
    uninstallModalInstance = new (window as any).bootstrap.Modal(
      uninstallModalEl
    );
  }

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
      autostartStatusEl.classList.remove('bg-secondary');
      autostartStatusEl.classList.add('bg-warning');
      autostartToggleBtn.disabled = true;
      return;
    }
    const data = await res.json();
    isAutostartEnabled = data.enabled;

    autostartStatusEl.textContent = isAutostartEnabled ? 'Đã bật' : 'Đã tắt';
    autostartStatusEl.classList.remove(
      'bg-secondary',
      'bg-danger',
      'bg-success'
    );
    autostartStatusEl.classList.add(
      isAutostartEnabled ? 'bg-success' : 'bg-secondary'
    );

    autostartToggleBtn.textContent = isAutostartEnabled
      ? 'Tắt tự động khởi động'
      : 'Bật tự động khởi động';
    autostartToggleBtn.disabled = false;
  } catch (e) {
    autostartStatusEl.textContent = 'Lỗi';
    autostartStatusEl.classList.remove('bg-secondary');
    autostartStatusEl.classList.add('bg-danger');
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

// This function is called by the onclick attribute in settings.html
function uninstall(): void {
  if (uninstallModalInstance) {
    // Clear any previous error messages when opening
    const uninstallError = document.getElementById(
      'uninstall-error'
    ) as HTMLParagraphElement;
    if (uninstallError) uninstallError.style.display = 'none';
    uninstallModalInstance.show();
  }
}

function closeUninstallModal(): void {
  if (uninstallModalInstance) {
    const uninstallPasswordInput = document.getElementById(
      'uninstall-password'
    ) as HTMLInputElement;
    uninstallModalInstance.hide();
    // Reset form state after modal closes
    if (uninstallPasswordInput) uninstallPasswordInput.value = '';
  }
}
