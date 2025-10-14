async function search(range?: { since: string; until: string }): Promise<void> {
  const q = document.getElementById('q') as HTMLInputElement;
  const sinceDateInput = document.getElementById(
    'since_date'
  ) as HTMLInputElement;
  const sinceTimeInput = document.getElementById(
    'since_time'
  ) as HTMLInputElement;
  const untilDateInput = document.getElementById(
    'until_date'
  ) as HTMLInputElement;
  const untilTimeInput = document.getElementById(
    'until_time'
  ) as HTMLInputElement;
  const results = document.getElementById('results') as HTMLDivElement;

  let since = '';
  let until = '';

  if (range) {
    since = range.since;
    until = range.until;
  } else {
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
  const res = await fetch(url);
  const data = await res.json();
  if (data && data.length > 0) {
    results.innerHTML = data
      .map((l: string[]) => {
        const processName = l[1];
        return `<label class="list-group-item">
                  <input class="form-check-input me-2" type="checkbox" name="search-result-app" value="${processName}">
                  ${l.join(' | ')}
                </label>`;
      })
      .join('');
  } else {
    results.innerHTML = '<div class="list-group-item">Không tìm thấy kết quả.</div>';
  }
}

async function block(): Promise<void> {
  const blockStatus = document.getElementById(
    'block-status'
  ) as HTMLSpanElement;
  const selectedApps = Array.from(
    document.querySelectorAll('input[name="search-result-app"]:checked')
  ).map((cb) => (cb as HTMLInputElement).value);
  if (selectedApps.length === 0) {
    alert('Vui lòng chọn một ứng dụng từ kết quả tìm kiếm để chặn.');
    return;
  }

  // Remove duplicates
  const uniqueApps = [...new Set(selectedApps)];

  await fetch('/api/block', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ names: uniqueApps }),
  });
  blockStatus.innerText = 'Đã chặn: ' + uniqueApps.join(', ');
  setTimeout(() => {
    blockStatus.innerText = '';
  }, 3000);
}

async function loadBlocklist(): Promise<void> {
  const blocklistItems = document.getElementById(
    'blocklist-items'
  ) as HTMLDivElement;
  const res = await fetch('/api/blocklist');
  const data = await res.json();
  if (data && data.length > 0) {
    blocklistItems.innerHTML = data.map((app: string) => {
      return `<label class="list-group-item">
                <input class="form-check-input me-2" type="checkbox" name="blocked-app" value="${app}">
                ${app}
              </label>`;
    }).join('');
  } else {
    blocklistItems.innerHTML = '<div class="list-group-item">Hiện không có ứng dụng nào bị chặn.</div>';
  }
}

async function unblockSelected(): Promise<void> {
  const unblockStatus = document.getElementById(
    'unblock-status'
  ) as HTMLSpanElement;
  const selectedApps = Array.from(
    document.querySelectorAll('input[name="blocked-app"]:checked')
  ).map((cb) => (cb as HTMLInputElement).value);
  if (selectedApps.length === 0) {
    alert('Vui lòng chọn các ứng dụng để bỏ chặn.');
    return;
  }
  await fetch('/api/unblock', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ names: selectedApps }),
  });
  unblockStatus.innerText = 'Đã bỏ chặn: ' + selectedApps.join(', ');
  setTimeout(() => {
    unblockStatus.innerText = '';
  }, 3000);
  loadBlocklist(); // Refresh the list
}

async function clearBlocklist(): Promise<void> {
  const unblockStatus = document.getElementById(
    'unblock-status'
  ) as HTMLSpanElement;
  if (confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn không?')) {
    await fetch('/api/blocklist/clear', { method: 'POST' });
    unblockStatus.innerText = 'Đã xóa toàn bộ danh sách chặn.';
    setTimeout(() => {
      unblockStatus.innerText = '';
    }, 3000);
    loadBlocklist(); // Refresh the list
  }
}

async function saveBlocklist(): Promise<void> {
  const response = await fetch('/api/blocklist/save');
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.style.display = 'none';
  a.href = url;
  a.download = 'procguard_blocklist.json';
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
}

async function loadBlocklistFile(event: Event): Promise<void> {
  const unblockStatus = document.getElementById(
    'unblock-status'
  ) as HTMLSpanElement;
  const file = (event.target as HTMLInputElement).files?.[0];
  if (!file) {
    return;
  }
  const formData = new FormData();
  formData.append('file', file);

  await fetch('/api/blocklist/load', {
    method: 'POST',
    body: formData,
  });

  unblockStatus.innerText = 'Đã tải lên và hợp nhất danh sách chặn.';
  setTimeout(() => {
    unblockStatus.innerText = '';
  }, 3000);
  loadBlocklist(); // Refresh the list
}
