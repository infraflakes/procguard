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
    const itemsHtml = await Promise.all(
      data.map(async (l: string[]) => {
        const processName = l[1];
        const exePath = l[4]; // exe_path is the 5th element
        let commercialName = '';
        let icon = '';

        if (exePath) {
          const appDetailsRes = await fetch(
            `/api/app-details?path=${encodeURIComponent(exePath)}`
          );
          if (appDetailsRes.ok) {
            const appDetails = await appDetailsRes.json();
            commercialName = appDetails.commercialName;
            icon = appDetails.icon;
          }
        }

        const otherInfo = l.filter((_, i) => i !== 1 && i !== 4).join(' | ');

        return `<label class="list-group-item d-flex align-items-center">
                <input class="form-check-input me-2" type="checkbox" name="search-result-app" value="${processName}">
                ${
                  icon
                    ? `<img src="data:image/png;base64,${icon}" class="me-2" style="width: 24px; height: 24px;">`
                    : '<div class="me-2" style="width: 24px; height: 24px;"></div>'
                }
                <span class="fw-bold me-2">${
                  commercialName || processName
                }</span>
                <span class="text-muted">${otherInfo}</span>
              </label>`;
      })
    );
    results.innerHTML = itemsHtml.join('');
  } else {
    results.innerHTML =
      '<div class="list-group-item">Không tìm thấy kết quả.</div>';
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
    const itemsHtml = await Promise.all(
      data.map(async (app: { name: string; exe_path: string }) => {
        let commercialName = '';
        let icon = '';

        if (app.exe_path) {
          const appDetailsRes = await fetch(
            `/api/app-details?path=${encodeURIComponent(app.exe_path)}`
          );
          if (appDetailsRes.ok) {
            const appDetails = await appDetailsRes.json();
            commercialName = appDetails.commercialName;
            icon = appDetails.icon;
          }
        }

        return `<label class="list-group-item d-flex align-items-center">
                <input class="form-check-input me-2" type="checkbox" name="blocked-app" value="${
                  app.name
                }">
                ${
                  icon
                    ? `<img src="data:image/png;base64,${icon}" class="me-2" style="width: 24px; height: 24px;">`
                    : '<div class="me-2" style="width: 24px; height: 24px;"></div>'
                }
                <span class="fw-bold me-2">${commercialName || app.name}</span>
              </label>`;
      })
    );
    blocklistItems.innerHTML = itemsHtml.join('');
  } else {
    blocklistItems.innerHTML =
      '<div class="list-group-item">Hiện không có ứng dụng nào bị chặn.</div>';
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

async function loadAppLeaderboard(since = '', until = ''): Promise<void> {
  const container = document.getElementById('app-leaderboard-table-container');
  if (!container) {
    return;
  }

  let url = '/api/leaderboard/apps';
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

  if (data && data.length > 0) {
    const table = document.createElement('table');
    table.className = 'table table-hover';
    const thead = document.createElement('thead');
    thead.innerHTML = `
      <tr>
        <th scope="col">Rank</th>
        <th scope="col">Application</th>
        <th scope="col">Usage Count</th>
      </tr>
    `;
    table.appendChild(thead);

    const tbody = document.createElement('tbody');
    tbody.innerHTML = data
      .map(
        (item: { rank: number; name: string; icon: string; count: number }) => `
      <tr>
        <th scope="row"><span class="badge bg-primary">${item.rank}</span></th>
        <td>
          ${
            item.icon
              ? `<img src="data:image/png;base64,${item.icon}" class="me-2" style="width: 24px; height: 24px;">`
              : '<div class="me-2" style="width: 24px; height: 24px;"></div>'
          }
          <span class="fw-bold">${item.name}</span>
        </td>
        <td>${item.count}</td>
      </tr>
    `
      )
      .join('');
    table.appendChild(tbody);

    container.innerHTML = '';
    container.appendChild(table);
  } else {
    container.innerHTML =
      '<div class="list-group-item">No data for leaderboard.</div>';
  }
}
