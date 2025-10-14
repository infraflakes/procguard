"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
let isAutostartEnabled = false;
// This variable will hold the Bootstrap Modal instance
let uninstallModalInstance; // Use 'any' to avoid needing full Bootstrap types
document.addEventListener('DOMContentLoaded', () => {
    // Initialize the modal instance once the DOM is ready
    const uninstallModalEl = document.getElementById('uninstall-modal');
    if (uninstallModalEl) {
        // Assuming Bootstrap's JS is loaded and available globally
        uninstallModalInstance = new window.bootstrap.Modal(uninstallModalEl);
    }
    const uninstallForm = document.getElementById('uninstall-form');
    if (uninstallForm) {
        uninstallForm.addEventListener('submit', (event) => __awaiter(void 0, void 0, void 0, function* () {
            event.preventDefault();
            const uninstallPasswordInput = document.getElementById('uninstall-password');
            const uninstallError = document.getElementById('uninstall-error');
            const password = uninstallPasswordInput.value;
            const response = yield fetch('/api/uninstall', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ password }),
            });
            if (response.ok) {
                closeUninstallModal();
                alert('ProcGuard đã được gỡ cài đặt. Trình duyệt sẽ đóng.');
                window.close();
            }
            else {
                uninstallError.textContent =
                    'Gỡ cài đặt thất bại. Vui lòng kiểm tra lại mật khẩu.';
                uninstallError.style.display = 'block';
            }
        }));
    }
});
function loadAutostartStatus() {
    return __awaiter(this, void 0, void 0, function* () {
        const autostartStatusEl = document.getElementById('autostart-status');
        const autostartToggleBtn = document.getElementById('autostart-toggle-btn');
        try {
            const res = yield fetch('/api/settings/autostart/status');
            if (!res.ok) {
                autostartStatusEl.textContent = 'Không hỗ trợ trên HĐH này';
                autostartStatusEl.classList.remove('bg-secondary');
                autostartStatusEl.classList.add('bg-warning');
                autostartToggleBtn.disabled = true;
                return;
            }
            const data = yield res.json();
            isAutostartEnabled = data.enabled;
            autostartStatusEl.textContent = isAutostartEnabled ? 'Đã bật' : 'Đã tắt';
            autostartStatusEl.classList.remove('bg-secondary', 'bg-danger', 'bg-success');
            autostartStatusEl.classList.add(isAutostartEnabled ? 'bg-success' : 'bg-secondary');
            autostartToggleBtn.textContent = isAutostartEnabled
                ? 'Tắt tự động khởi động'
                : 'Bật tự động khởi động';
            autostartToggleBtn.disabled = false;
        }
        catch (e) {
            autostartStatusEl.textContent = 'Lỗi';
            autostartStatusEl.classList.remove('bg-secondary');
            autostartStatusEl.classList.add('bg-danger');
            autostartToggleBtn.disabled = true;
        }
    });
}
function toggleAutostart() {
    return __awaiter(this, void 0, void 0, function* () {
        const autostartToggleBtn = document.getElementById('autostart-toggle-btn');
        const endpoint = isAutostartEnabled
            ? '/api/settings/autostart/disable'
            : '/api/settings/autostart/enable';
        try {
            autostartToggleBtn.disabled = true;
            const res = yield fetch(endpoint, { method: 'POST' });
            if (!res.ok) {
                const errorText = yield res.text();
                alert(`Thao tác thất bại: ${errorText}`);
            }
            else {
                yield loadAutostartStatus(); // Refresh status after action
            }
        }
        catch (e) {
            if (e instanceof Error) {
                alert(`Đã xảy ra lỗi: ${e.message}`);
            }
        }
        finally {
            autostartToggleBtn.disabled = false;
        }
    });
}
// This function is called by the onclick attribute in settings.html
function uninstall() {
    if (uninstallModalInstance) {
        // Clear any previous error messages when opening
        const uninstallError = document.getElementById('uninstall-error');
        if (uninstallError)
            uninstallError.style.display = 'none';
        uninstallModalInstance.show();
    }
}
function closeUninstallModal() {
    if (uninstallModalInstance) {
        const uninstallPasswordInput = document.getElementById('uninstall-password');
        uninstallModalInstance.hide();
        // Reset form state after modal closes
        if (uninstallPasswordInput)
            uninstallPasswordInput.value = '';
    }
}
