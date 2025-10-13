async function search(range){
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
  } else {
    if (sinceDateInput.value && sinceTimeInput.value) {
        since = `${sinceDateInput.value}T${sinceTimeInput.value}`;
    }
    if (untilDateInput.value && untilTimeInput.value) {
        until = `${untilDateInput.value}T${untilTimeInput.value}`;
    }
  }

  let url = '/api/search?q='+encodeURIComponent(q.value);
  if (since) {
    url += '&since=' + encodeURIComponent(since);
  }
  if (until) {
    url += '&until=' + encodeURIComponent(until);
  }
  const res = await fetch(url);
  const data = await res.json();
  if (data && data.length > 0) {
    results.innerHTML = data.map(l => {
        const processName = l[1];
        return `<div class="result-item"><input type="checkbox" name="search-result-app" value="${processName}"> ${l.join(' | ')}</div>`;
    }).join('');
  } else {
    results.innerHTML = "Không tìm thấy kết quả.";
  }
}

async function block(){
  const blockStatus = document.getElementById('block-status');
  const selectedApps = Array.from(document.querySelectorAll('input[name="search-result-app"]:checked')).map(cb => cb.value);
  if (selectedApps.length === 0) {
    alert("Vui lòng chọn một ứng dụng từ kết quả tìm kiếm để chặn.");
    return;
  }
  
  // Remove duplicates
  const uniqueApps = [...new Set(selectedApps)];

  await fetch('/api/block',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({names:uniqueApps})});
  blockStatus.innerText = 'Đã chặn: ' + uniqueApps.join(', ');
  setTimeout(() => { blockStatus.innerText = ''; }, 3000);
}

async function loadBlocklist() {
    const blocklistItems = document.getElementById('blocklist-items');
    const res = await fetch('/api/blocklist');
    const data = await res.json();
    blocklistItems.innerHTML = '';
    if (data && data.length > 0) {
        data.forEach(app => {
            blocklistItems.innerHTML += `<div><input type="checkbox" name="blocked-app" value="${app}"> ${app}</div>`;
        });
    } else {
        blocklistItems.innerHTML = 'Hiện không có ứng dụng nào bị chặn.';
    }
}

async function unblockSelected() {
    const unblockStatus = document.getElementById('unblock-status');
    const selectedApps = Array.from(document.querySelectorAll('input[name="blocked-app"]:checked')).map(cb => cb.value);
    if (selectedApps.length === 0) {
        alert("Vui lòng chọn các ứng dụng để bỏ chặn.");
        return;
    }
    await fetch('/api/unblock', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({names: selectedApps})
    });
    unblockStatus.innerText = 'Đã bỏ chặn: ' + selectedApps.join(', ');
    setTimeout(() => { unblockStatus.innerText = ''; }, 3000);
    loadBlocklist(); // Refresh the list
}

async function clearBlocklist() {
    const unblockStatus = document.getElementById('unblock-status');
    if (confirm("Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn không?")) {
        await fetch('/api/blocklist/clear', { method: 'POST' });
        unblockStatus.innerText = 'Đã xóa toàn bộ danh sách chặn.';
        setTimeout(() => { unblockStatus.innerText = ''; }, 3000);
        loadBlocklist(); // Refresh the list
    }
}

async function saveBlocklist() {
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

async function loadBlocklistFile(event) {
    const unblockStatus = document.getElementById('unblock-status');
    const file = event.target.files[0];
    if (!file) {
        return;
    }
    const formData = new FormData();
    formData.append('file', file);

    await fetch('/api/blocklist/load', {
        method: 'POST',
        body: formData
    });

    unblockStatus.innerText = 'Đã tải lên và hợp nhất danh sách chặn.';
    setTimeout(() => { unblockStatus.innerText = ''; }, 3000);
    loadBlocklist(); // Refresh the list
}
