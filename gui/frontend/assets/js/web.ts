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
  webLogItems.innerHTML = '';
  if (data && data.length > 0) {
    // Data is already sorted by timestamp descending from the server
    webLogItems.innerHTML = data
      .map((l: string[]) => {
        const urlString = l[1];
        let domain = '';
        try {
          const url = new URL(urlString);
          domain = url.hostname;
        } catch (e) {
          // Ignore invalid URLs
        }
        return `<div class="result-item"><input type="checkbox" name="web-log-domain" value="${domain}"> ${l.join(' | ')}</div>`;
      })
      .join('');
  } else {
    webLogItems.innerHTML = 'Chưa có lịch sử truy cập web.';
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
  webBlocklistItems.innerHTML = '';
  if (data && data.length > 0) {
    data.forEach((domain: string) => {
      webBlocklistItems.innerHTML += `<div><input type="checkbox" name="blocked-website" value="${domain}"> ${domain} <button onclick="removeWebBlocklist('${domain}')">X</button></div>`;
    });
  } else {
    webBlocklistItems.innerHTML = 'Hiện không có trang web nào bị chặn.';
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
  unblockWebStatus.innerText = 'Đã bỏ chặn: ' + selectedWebsites.join(', ');
  setTimeout(() => {
    unblockWebStatus.innerText = '';
  }, 3000);
  loadWebBlocklist(); // Refresh the list
}

async function clearWebBlocklist(): Promise<void> {
  const unblockWebStatus = document.getElementById(
    'unblock-web-status'
  ) as HTMLSpanElement;
  if (confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn web không?')) {
    await fetch('/api/web-blocklist/clear', { method: 'POST' });
    unblockWebStatus.innerText = 'Đã xóa toàn bộ danh sách chặn web.';
    setTimeout(() => {
      unblockWebStatus.innerText = '';
    }, 3000);
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

  unblockWebStatus.innerText = 'Đã tải lên và hợp nhất danh sách chặn web.';
  setTimeout(() => {
    unblockWebStatus.innerText = '';
  }, 3000);
  loadWebBlocklist(); // Refresh the list
}
