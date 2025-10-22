declare function checkExtension(callback?: (success: boolean) => void): void;
declare function showSubView(viewName: string, parentView: string): void;

function showWebManagementView(): void {
  const notInstalledView = document.getElementById(
    'web-extension-not-installed-view'
  );
  const webManTabs = document.getElementById('webManTabs');
  const webManTabsContent = document.getElementById('webManTabsContent');

  if (notInstalledView && webManTabs && webManTabsContent) {
    if ((window as any).isExtensionInstalled) {
      notInstalledView.style.display = 'none';
      webManTabs.style.display = 'flex';
      webManTabsContent.style.display = 'block';
      showSubView('web-log-view', 'web-management-view');
    } else {
      notInstalledView.style.display = 'block';
      webManTabs.style.display = 'none';
      webManTabsContent.style.display = 'none';
    }
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const reloadBtn = document.getElementById(
    'reload-extension-check-btn'
  ) as HTMLButtonElement;
  if (reloadBtn) {
    reloadBtn.addEventListener('click', () => {
      // Show a loading indicator
      reloadBtn.textContent = 'Checking...';
      reloadBtn.disabled = true;
      checkExtension((success) => {
        showWebManagementView();
        reloadBtn.textContent = 'I have installed it, Reload';
        reloadBtn.disabled = false;
      });
    });
  }
});

async function searchWebLogs(range?: {
  since: string;
  until: string;
}): Promise<void> {
  const webSinceDateInput = document.getElementById(
    'web_since_date'
  ) as HTMLInputElement;
  const webSinceTimeInput = document.getElementById(
    'web_since_time'
  ) as HTMLInputElement;
  const webUntilDateInput = document.getElementById(
    'web_until_date'
  ) as HTMLInputElement;
  const webUntilTimeInput = document.getElementById(
    'web_until_time'
  ) as HTMLInputElement;

  let since = '';
  let until = '';

  if (range) {
    since = range.since;
    until = range.until;
  } else {
    if (webSinceDateInput.value && webSinceTimeInput.value) {
      since = `${webSinceDateInput.value}T${webSinceTimeInput.value}`;
    }
    if (webUntilDateInput.value && webUntilTimeInput.value) {
      until = `${webUntilDateInput.value}T${webUntilTimeInput.value}`;
    }
  }
  await loadWebLogs(since, until);
}

async function loadWebLogs(since = '', until = ''): Promise<void> {
  const webLogItems = document.getElementById(
    'web-log-items'
  ) as HTMLDivElement;
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

  const res = await fetch(url);
  const data = await res.json();
  if (webLogItems) {
    webLogItems.innerHTML = '';
    if (data && data.length > 0) {
      const itemsHtml = await Promise.all(
        data.map(async (l: string[]) => {
          const urlString = l[1];
          let domain = '';
          try {
            const url = new URL(urlString);
            domain = url.hostname;
          } catch (e) {
            // Ignore invalid URLs
          }

          let title = '';
          let iconUrl = '';
          if (domain) {
            const webDetailsRes = await fetch(
              `/api/web-details?domain=${domain}`
            );
            if (webDetailsRes.ok) {
              const webDetails = await webDetailsRes.json();
              title = webDetails.title;
              iconUrl = webDetails.icon_url;
            }
          }

          const otherInfo = l[0]; // Just the timestamp

          return `<label class="list-group-item d-flex align-items-center">
                  <input class="form-check-input me-2" type="checkbox" name="web-log-domain" value="${domain}">
                  ${
                    iconUrl
                      ? `<img src="${iconUrl}" class="me-2" style="width: 24px; height: 24px;">`
                      : '<div class="me-2" style="width: 24px; height: 24px;"></div>'
                  }
                  <span class="fw-bold me-2">${title || domain}</span>
                  <span class="text-muted ms-auto">${otherInfo}</span>
                </label>`;
        })
      );
      webLogItems.innerHTML = itemsHtml.join('');
    } else {
      webLogItems.innerHTML =
        '<div class="list-group-item">Chưa có lịch sử truy cập web.</div>';
    }
  }
}

async function blockSelectedWebsites(): Promise<void> {
  const selectedDomains = Array.from(
    document.querySelectorAll('input[name="web-log-domain"]:checked')
  ).map((cb) => (cb as HTMLInputElement).value);
  if (selectedDomains.length === 0) {
    alert('Vui lòng chọn một trang web để chặn.');
    return;
  }

  const uniqueDomains = [...new Set(selectedDomains)];

  for (const domain of uniqueDomains) {
    await fetch('/api/web-blocklist/add', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ domain: domain }),
    });
  }

  alert('Các trang web đã chọn đã được thêm vào danh sách chặn.');
  // Uncheck all boxes
  (
    document.querySelectorAll(
      'input[name="web-log-domain"]:checked'
    ) as NodeListOf<HTMLInputElement>
  ).forEach((cb) => (cb.checked = false));
}

async function loadWebBlocklist(): Promise<void> {
  const webBlocklistItems = document.getElementById(
    'web-blocklist-items'
  ) as HTMLDivElement;
  const res = await fetch('/api/web-blocklist');
  const data = await res.json();
  if (webBlocklistItems) {
    webBlocklistItems.innerHTML = '';
    if (data && data.length > 0) {
      webBlocklistItems.innerHTML = data
        .map((item: { domain: string; title: string; icon_url: string }) => {
          const { domain, title, icon_url } = item;
          return `
          <div class="list-group-item d-flex justify-content-between align-items-center">
            <label class="flex-grow-1 mb-0 d-flex align-items-center">
              <input class="form-check-input me-2" type="checkbox" name="blocked-website" value="${domain}">
              ${
                icon_url
                  ? `<img src="${icon_url}" class="me-2" style="width: 24px; height: 24px;">`
                  : '<div class="me-2" style="width: 24px; height: 24px;"></div>'
              }
              <span class="fw-bold me-2">${title || domain}</span>
            </label>
            <button class="btn btn-sm btn-outline-danger" onclick="removeWebBlocklist('${domain}')">&times;</button>
          </div>
        `;
        })
        .join('');
    } else {
      webBlocklistItems.innerHTML =
        '<div class="list-group-item">Hiện không có trang web nào bị chặn.</div>';
    }
  }
}

async function removeWebBlocklist(domain: string): Promise<void> {
  if (confirm(`Bạn có chắc chắn muốn bỏ chặn ${domain} không?`)) {
    await fetch('/api/web-blocklist/remove', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ domain: domain }),
    });
    loadWebBlocklist();
  }
}

async function unblockSelectedWebsites(): Promise<void> {
  const unblockWebStatus = document.getElementById(
    'unblock-web-status'
  ) as HTMLSpanElement;
  const selectedWebsites = Array.from(
    document.querySelectorAll('input[name="blocked-website"]:checked')
  ).map((cb) => (cb as HTMLInputElement).value);
  if (selectedWebsites.length === 0) {
    alert('Vui lòng chọn các trang web để bỏ chặn.');
    return;
  }
  for (const domain of selectedWebsites) {
    await fetch('/api/web-blocklist/remove', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ domain: domain }),
    });
  }
  if (unblockWebStatus) {
    unblockWebStatus.innerText = 'Đã bỏ chặn: ' + selectedWebsites.join(', ');
    setTimeout(() => {
      unblockWebStatus.innerText = '';
    }, 3000);
  }
  loadWebBlocklist(); // Refresh the list
}

async function clearWebBlocklist(): Promise<void> {
  const unblockWebStatus = document.getElementById(
    'unblock-web-status'
  ) as HTMLSpanElement;
  if (confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn web không?')) {
    await fetch('/api/web-blocklist/clear', { method: 'POST' });
    if (unblockWebStatus) {
      unblockWebStatus.innerText = 'Đã xóa toàn bộ danh sách chặn web.';
      setTimeout(() => {
        unblockWebStatus.innerText = '';
      }, 3000);
    }
    loadWebBlocklist(); // Refresh the list
  }
}

async function saveWebBlocklist(): Promise<void> {
  const response = await fetch('/api/web-blocklist/save');
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.style.display = 'none';
  a.href = url;
  a.download = 'procguard_web_blocklist.json';
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
}

async function loadWebBlocklistFile(event: Event): Promise<void> {
  const unblockWebStatus = document.getElementById(
    'unblock-web-status'
  ) as HTMLSpanElement;
  const file = (event.target as HTMLInputElement).files?.[0];
  if (!file) {
    return;
  }
  const formData = new FormData();
  formData.append('file', file);

  await fetch('/api/web-blocklist/load', {
    method: 'POST',
    body: formData,
  });

  if (unblockWebStatus) {
    unblockWebStatus.innerText = 'Đã tải lên và hợp nhất danh sách chặn web.';
    setTimeout(() => {
      unblockWebStatus.innerText = '';
    }, 3000);
  }
  loadWebBlocklist(); // Refresh the list
}
