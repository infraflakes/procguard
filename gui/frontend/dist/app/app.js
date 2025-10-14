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
function search(range) {
    return __awaiter(this, void 0, void 0, function* () {
        const q = document.getElementById('q');
        const sinceDateInput = document.getElementById('since_date');
        const sinceTimeInput = document.getElementById('since_time');
        const untilDateInput = document.getElementById('until_date');
        const untilTimeInput = document.getElementById('until_time');
        const results = document.getElementById('results');
        let since = '';
        let until = '';
        if (range) {
            since = range.since;
            until = range.until;
        }
        else {
            if (sinceDateInput.value && sinceTimeInput.value) {
                since = `${sinceDateInput.value}T${sinceTimeInput.value}`;
            }
            if (untilDateInput.value && untilTimeInput.value) {
                until = `${untilDateInput.value}T${untilTimeInput.value}`;
            }
        }
        let url = '/api/search?q=' + encodeURIComponent(q.value);
        if (since) {
            url += '&since=' + encodeURIComponent(since);
        }
        if (until) {
            url += '&until=' + encodeURIComponent(until);
        }
        const res = yield fetch(url);
        const data = yield res.json();
        if (data && data.length > 0) {
            results.innerHTML = data
                .map((l) => {
                const processName = l[1];
                return `<label class="list-group-item">
                  <input class="form-check-input me-2" type="checkbox" name="search-result-app" value="${processName}">
                  ${l.join(' | ')}
                </label>`;
            })
                .join('');
        }
        else {
            results.innerHTML = '<div class="list-group-item">Không tìm thấy kết quả.</div>';
        }
    });
}
function block() {
    return __awaiter(this, void 0, void 0, function* () {
        const blockStatus = document.getElementById('block-status');
        const selectedApps = Array.from(document.querySelectorAll('input[name="search-result-app"]:checked')).map((cb) => cb.value);
        if (selectedApps.length === 0) {
            alert('Vui lòng chọn một ứng dụng từ kết quả tìm kiếm để chặn.');
            return;
        }
        // Remove duplicates
        const uniqueApps = [...new Set(selectedApps)];
        yield fetch('/api/block', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ names: uniqueApps }),
        });
        blockStatus.innerText = 'Đã chặn: ' + uniqueApps.join(', ');
        setTimeout(() => {
            blockStatus.innerText = '';
        }, 3000);
    });
}
function loadBlocklist() {
    return __awaiter(this, void 0, void 0, function* () {
        const blocklistItems = document.getElementById('blocklist-items');
        const res = yield fetch('/api/blocklist');
        const data = yield res.json();
        if (data && data.length > 0) {
            blocklistItems.innerHTML = data.map((app) => {
                return `<label class="list-group-item">
                <input class="form-check-input me-2" type="checkbox" name="blocked-app" value="${app}">
                ${app}
              </label>`;
            }).join('');
        }
        else {
            blocklistItems.innerHTML = '<div class="list-group-item">Hiện không có ứng dụng nào bị chặn.</div>';
        }
    });
}
function unblockSelected() {
    return __awaiter(this, void 0, void 0, function* () {
        const unblockStatus = document.getElementById('unblock-status');
        const selectedApps = Array.from(document.querySelectorAll('input[name="blocked-app"]:checked')).map((cb) => cb.value);
        if (selectedApps.length === 0) {
            alert('Vui lòng chọn các ứng dụng để bỏ chặn.');
            return;
        }
        yield fetch('/api/unblock', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ names: selectedApps }),
        });
        unblockStatus.innerText = 'Đã bỏ chặn: ' + selectedApps.join(', ');
        setTimeout(() => {
            unblockStatus.innerText = '';
        }, 3000);
        loadBlocklist(); // Refresh the list
    });
}
function clearBlocklist() {
    return __awaiter(this, void 0, void 0, function* () {
        const unblockStatus = document.getElementById('unblock-status');
        if (confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn không?')) {
            yield fetch('/api/blocklist/clear', { method: 'POST' });
            unblockStatus.innerText = 'Đã xóa toàn bộ danh sách chặn.';
            setTimeout(() => {
                unblockStatus.innerText = '';
            }, 3000);
            loadBlocklist(); // Refresh the list
        }
    });
}
function saveBlocklist() {
    return __awaiter(this, void 0, void 0, function* () {
        const response = yield fetch('/api/blocklist/save');
        const blob = yield response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        a.download = 'procguard_blocklist.json';
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
    });
}
function loadBlocklistFile(event) {
    return __awaiter(this, void 0, void 0, function* () {
        var _a;
        const unblockStatus = document.getElementById('unblock-status');
        const file = (_a = event.target.files) === null || _a === void 0 ? void 0 : _a[0];
        if (!file) {
            return;
        }
        const formData = new FormData();
        formData.append('file', file);
        yield fetch('/api/blocklist/load', {
            method: 'POST',
            body: formData,
        });
        unblockStatus.innerText = 'Đã tải lên và hợp nhất danh sách chặn.';
        setTimeout(() => {
            unblockStatus.innerText = '';
        }, 3000);
        loadBlocklist(); // Refresh the list
    });
}
