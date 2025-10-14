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
function searchWebLogs(range) {
    return __awaiter(this, void 0, void 0, function* () {
        const webSinceDateInput = document.getElementById('web_since_date');
        const webSinceTimeInput = document.getElementById('web_since_time');
        const webUntilDateInput = document.getElementById('web_until_date');
        const webUntilTimeInput = document.getElementById('web_until_time');
        let since = '';
        let until = '';
        if (range) {
            since = range.since;
            until = range.until;
        }
        else {
            if (webSinceDateInput.value && webSinceTimeInput.value) {
                since = `${webSinceDateInput.value}T${webSinceTimeInput.value}`;
            }
            if (webUntilDateInput.value && webUntilTimeInput.value) {
                until = `${webUntilDateInput.value}T${webUntilTimeInput.value}`;
            }
        }
        yield loadWebLogs(since, until);
    });
}
function loadWebLogs() {
    return __awaiter(this, arguments, void 0, function* (since = '', until = '') {
        const webLogItems = document.getElementById('web-log-items');
        let url = '/api/web-logs';
        const params = new URLSearchParams();
        if (since) {
            params.append('since', since);
        }
        if (until) {
            params.append('until', until);
        }
        const queryString = params.toString();
        if (queryString) {
            url += `?${queryString}`;
        }
        const res = yield fetch(url);
        const data = yield res.json();
        webLogItems.innerHTML = '';
        if (data && data.length > 0) {
            webLogItems.innerHTML = data
                .map((l) => {
                const urlString = l[1];
                let domain = '';
                try {
                    const url = new URL(urlString);
                    domain = url.hostname;
                }
                catch (e) {
                    // Ignore invalid URLs
                }
                return `<label class="list-group-item">
                  <input class="form-check-input me-2" type="checkbox" name="web-log-domain" value="${domain}">
                  ${l.join(' | ')}
                </label>`;
            })
                .join('');
        }
        else {
            webLogItems.innerHTML = '<div class="list-group-item">Chưa có lịch sử truy cập web.</div>';
        }
    });
}
function blockSelectedWebsites() {
    return __awaiter(this, void 0, void 0, function* () {
        const selectedDomains = Array.from(document.querySelectorAll('input[name="web-log-domain"]:checked')).map((cb) => cb.value);
        if (selectedDomains.length === 0) {
            alert('Vui lòng chọn một trang web để chặn.');
            return;
        }
        const uniqueDomains = [...new Set(selectedDomains)];
        for (const domain of uniqueDomains) {
            yield fetch('/api/web-blocklist/add', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ domain: domain }),
            });
        }
        alert('Các trang web đã chọn đã được thêm vào danh sách chặn.');
        // Uncheck all boxes
        document.querySelectorAll('input[name="web-log-domain"]:checked').forEach((cb) => (cb.checked = false));
    });
}
function loadWebBlocklist() {
    return __awaiter(this, void 0, void 0, function* () {
        const webBlocklistItems = document.getElementById('web-blocklist-items');
        const res = yield fetch('/api/web-blocklist');
        const data = yield res.json();
        webBlocklistItems.innerHTML = '';
        if (data && data.length > 0) {
            webBlocklistItems.innerHTML = data.map((domain) => {
                return `
        <div class="list-group-item d-flex justify-content-between align-items-center">
          <label class="flex-grow-1 mb-0">
            <input class="form-check-input me-2" type="checkbox" name="blocked-website" value="${domain}">
            ${domain}
          </label>
          <button class="btn btn-sm btn-outline-danger" onclick="removeWebBlocklist('${domain}')">&times;</button>
        </div>
      `;
            }).join('');
        }
        else {
            webBlocklistItems.innerHTML = '<div class="list-group-item">Hiện không có trang web nào bị chặn.</div>';
        }
    });
}
function removeWebBlocklist(domain) {
    return __awaiter(this, void 0, void 0, function* () {
        if (confirm(`Bạn có chắc chắn muốn bỏ chặn ${domain} không?`)) {
            yield fetch('/api/web-blocklist/remove', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ domain: domain }),
            });
            loadWebBlocklist();
        }
    });
}
function unblockSelectedWebsites() {
    return __awaiter(this, void 0, void 0, function* () {
        const unblockWebStatus = document.getElementById('unblock-web-status');
        const selectedWebsites = Array.from(document.querySelectorAll('input[name="blocked-website"]:checked')).map((cb) => cb.value);
        if (selectedWebsites.length === 0) {
            alert('Vui lòng chọn các trang web để bỏ chặn.');
            return;
        }
        for (const domain of selectedWebsites) {
            yield fetch('/api/web-blocklist/remove', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ domain: domain }),
            });
        }
        unblockWebStatus.innerText = 'Đã bỏ chặn: ' + selectedWebsites.join(', ');
        setTimeout(() => {
            unblockWebStatus.innerText = '';
        }, 3000);
        loadWebBlocklist(); // Refresh the list
    });
}
function clearWebBlocklist() {
    return __awaiter(this, void 0, void 0, function* () {
        const unblockWebStatus = document.getElementById('unblock-web-status');
        if (confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn web không?')) {
            yield fetch('/api/web-blocklist/clear', { method: 'POST' });
            unblockWebStatus.innerText = 'Đã xóa toàn bộ danh sách chặn web.';
            setTimeout(() => {
                unblockWebStatus.innerText = '';
            }, 3000);
            loadWebBlocklist(); // Refresh the list
        }
    });
}
function saveWebBlocklist() {
    return __awaiter(this, void 0, void 0, function* () {
        const response = yield fetch('/api/web-blocklist/save');
        const blob = yield response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        a.download = 'procguard_web_blocklist.json';
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
    });
}
function loadWebBlocklistFile(event) {
    return __awaiter(this, void 0, void 0, function* () {
        var _a;
        const unblockWebStatus = document.getElementById('unblock-web-status');
        const file = (_a = event.target.files) === null || _a === void 0 ? void 0 : _a[0];
        if (!file) {
            return;
        }
        const formData = new FormData();
        formData.append('file', file);
        yield fetch('/api/web-blocklist/load', {
            method: 'POST',
            body: formData,
        });
        unblockWebStatus.innerText = 'Đã tải lên và hợp nhất danh sách chặn web.';
        setTimeout(() => {
            unblockWebStatus.innerText = '';
        }, 3000);
        loadWebBlocklist(); // Refresh the list
    });
}
