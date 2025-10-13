<script lang="ts">
  import { onMount } from 'svelte';
import { GetWebLogs, GetWebBlocklist, AddWebBlocklist, RemoveWebBlocklist, ClearWebBlocklist } from '../../wailsjs/go/main/App';

  let currentSubView = 'web-log-view';
  let webLogs: string[][] = [];
  let webBlocklist: string[] = [];
  let unblockWebStatus = '';

  onMount(() => {
    loadWebLogs();
    loadWebBlocklist();
  });

  function showSubView(view: string) {
    currentSubView = view;
  }

  async function loadWebLogs(since = '', until = '') {
    webLogs = await GetWebLogs(since, until);
  }

  async function handleSearchWebLogs(range?: { since: string; until: string }) {
    let since = range?.since || '';
    let until = range?.until || '';
    await loadWebLogs(since, until);
  }

  async function blockSelectedWebsites() {
    const selectedDomains = webLogs
      .filter((_, i) => {
        const checkbox = document.getElementById(`web-log-domain-${i}`) as HTMLInputElement;
        return checkbox.checked;
      })
      .map((log) => {
        try {
          return new URL(log[1]).hostname;
        } catch (e) {
          return null;
        }
      })
      .filter((domain) => domain !== null) as string[];

    if (selectedDomains.length === 0) {
      alert('Vui lòng chọn một trang web để chặn.');
      return;
    }

    const uniqueDomains = [...new Set(selectedDomains)];

    for (const domain of uniqueDomains) {
      await AddWebBlocklist(domain);
    }

    alert('Các trang web đã chọn đã được thêm vào danh sách chặn.');
    // Uncheck all boxes
    webLogs.forEach((_, i) => {
      const checkbox = document.getElementById(`web-log-domain-${i}`) as HTMLInputElement;
      if (checkbox) checkbox.checked = false;
    });
  }

  async function loadWebBlocklist() {
    webBlocklist = await GetWebBlocklist();
  }

  async function handleUnblockSelectedWebsites() {
    const selectedWebsites = webBlocklist.filter((_, i) => {
      const checkbox = document.getElementById(`blocked-website-${i}`) as HTMLInputElement;
      return checkbox.checked;
    });

    if (selectedWebsites.length === 0) {
      alert('Vui lòng chọn các trang web để bỏ chặn.');
      return;
    }

    for (const domain of selectedWebsites) {
      await RemoveWebBlocklist(domain);
    }

    unblockWebStatus = 'Đã bỏ chặn: ' + selectedWebsites.join(', ');
    setTimeout(() => (unblockWebStatus = ''), 3000);
    loadWebBlocklist();
  }

  async function handleClearWebBlocklist() {
    if (confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn web không?')) {
      await ClearWebBlocklist();
      unblockWebStatus = 'Đã xóa toàn bộ danh sách chặn web.';
      setTimeout(() => (unblockWebStatus = ''), 3000);
      loadWebBlocklist();
    }
  }
</script>

<div id="web-management-view">
  <div class="sub-nav">
    <button on:click={() => showSubView('web-log-view')}>Lịch sử Web</button>
    <button on:click={() => showSubView('web-blocklist-view')}>Danh sách chặn</button>
  </div>

  {#if currentSubView === 'web-log-view'}
    <div id="web-log-view">
      <h2>Lịch sử Truy cập Web</h2>
      <div id="web-log-controls">
        <button on:click={blockSelectedWebsites}>Chặn mục đã chọn</button>
      </div>
      <div id="web-time-filter-container">
        <span>Lọc theo thời gian:</span>
        <button on:click={() => handleSearchWebLogs({ since: '1 hour ago', until: 'now' })}>1 giờ qua</button>
        <button on:click={() => handleSearchWebLogs({ since: '24 hours ago', until: 'now' })}>24 giờ qua</button>
        <button on:click={() => handleSearchWebLogs({ since: '7 days ago', until: 'now' })}>7 ngày qua</button>
        <br />
        <span>Từ: </span><input type="date" id="web_since_date" />
        <input type="time" id="web_since_time" step="60" />
        <span>Đến: </span><input type="date" id="web_until_date" />
        <input type="time" id="web_until_time" step="60" />
        <button on:click={() => handleSearchWebLogs()}>Xác nhận</button>
      </div>
      <div id="web-log-items">
        {#each webLogs as log, i}
          <div class="result-item">
            <input type="checkbox" id={`web-log-domain-${i}`} />
            {log.join(' | ')}
          </div>
        {/each}
      </div>
    </div>
  {/if}

  {#if currentSubView === 'web-blocklist-view'}
    <div id="web-blocklist-view">
      <h2>Quản lý Danh sách chặn Web</h2>
      <div id="web-blocklist-controls">
        <button on:click={handleUnblockSelectedWebsites}>Bỏ chặn mục đã chọn</button>
        <button on:click={handleClearWebBlocklist} style="background-color: #d9534f">Xóa toàn bộ danh sách chặn</button>
        <span class="status-message">{unblockWebStatus}</span>
      </div>
      <div id="web-blocklist-items">
        {#each webBlocklist as domain, i}
          <div>
            <input type="checkbox" id={`blocked-website-${i}`} />
            {domain}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  /* Add component-specific styles here */
</style>
