<script lang="ts">
  import { onMount } from 'svelte';
import { Search, GetAppBlocklist, AddAppBlocklist, RemoveAppBlocklist, ClearAppBlocklist } from '../../wailsjs/go/main/App';

  let currentSubView = 'search-view';
  let searchQuery = '';
  let searchResults: string[][] = [];
  let blocklist: string[] = [];
  let blockStatus = '';
  let unblockStatus = '';

  onMount(() => {
    loadBlocklist();
  });

  function showSubView(view: string) {
    currentSubView = view;
  }

  async function handleSearch(range?: { since: string; until: string }) {
    let since = range?.since || '';
    let until = range?.until || '';
    searchResults = await Search(searchQuery, since, until);
  }

  async function handleBlock() {
    const selectedApps = searchResults
      .filter((_, i) => {
        const checkbox = document.getElementById(`search-result-${i}`) as HTMLInputElement;
        return checkbox.checked;
      })
      .map((result) => result[1]);

    if (selectedApps.length === 0) {
      alert('Vui lòng chọn một ứng dụng từ kết quả tìm kiếm để chặn.');
      return;
    }

    await AddAppBlocklist(selectedApps);
    blockStatus = 'Đã chặn: ' + selectedApps.join(', ');
    setTimeout(() => (blockStatus = ''), 3000);
  }

  async function loadBlocklist() {
    blocklist = await GetAppBlocklist();
  }

  async function handleUnblock() {
    const selectedApps = blocklist.filter((_, i) => {
      const checkbox = document.getElementById(`blocked-app-${i}`) as HTMLInputElement;
      return checkbox.checked;
    });

    if (selectedApps.length === 0) {
      alert('Vui lòng chọn các ứng dụng để bỏ chặn.');
      return;
    }

    await RemoveAppBlocklist(selectedApps);
    unblockStatus = 'Đã bỏ chặn: ' + selectedApps.join(', ');
    setTimeout(() => (unblockStatus = ''), 3000);
    loadBlocklist();
  }

  async function handleClearBlocklist() {
    if (confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn không?')) {
      await ClearAppBlocklist();
      unblockStatus = 'Đã xóa toàn bộ danh sách chặn.';
      setTimeout(() => (unblockStatus = ''), 3000);
      loadBlocklist();
    }
  }

  // Functions for save and load blocklist are not implemented as they require file system access
  // which is handled differently in Wails. We can add this later if needed.

</script>

<div id="app-management-view">
  <div class="sub-nav">
    <button on:click={() => showSubView('search-view')}>Tìm kiếm</button>
    <button on:click={() => showSubView('blocklist-view')}>Xem ứng dụng bị chặn</button>
  </div>

  {#if currentSubView === 'search-view'}
    <div id="search-view">
      <h2>Bảng điều khiển ProcGuard</h2>
      <div id="search-container">
        <input bind:value={searchQuery} placeholder="nhập tên ứng dụng" />
        <button on:click={() => handleSearch()}>Tìm kiếm Log</button>
        <button on:click={handleBlock}>Chặn mục đã chọn</button>
        <span class="status-message">{blockStatus}</span>
      </div>
      <div id="time-filter-container">
        <span>Lọc theo thời gian:</span>
        <button on:click={() => handleSearch({ since: '1 hour ago', until: 'now' })}>1 giờ qua</button>
        <button on:click={() => handleSearch({ since: '24 hours ago', until: 'now' })}>24 giờ qua</button>
        <button on:click={() => handleSearch({ since: '7 days ago', until: 'now' })}>7 ngày qua</button>
        <br />
        <span>Từ: </span><input type="date" id="since_date" />
        <input type="time" id="since_time" step="60" />
        <span>Đến: </span><input type="date" id="until_date" />
        <input type="time" id="until_time" step="60" />
        <button on:click={() => handleSearch()}>Xác nhận</button>
      </div>
      <h3>Kết quả tìm kiếm</h3>
      <div id="results">
        {#each searchResults as result, i}
          <div class="result-item">
            <input type="checkbox" id={`search-result-${i}`} />
            {result.join(' | ')}
          </div>
        {/each}
      </div>
    </div>
  {/if}

  {#if currentSubView === 'blocklist-view'}
    <div id="blocklist-view">
      <h2>Các ứng dụng bị chặn</h2>
      <div id="blocklist-controls">
        <button on:click={handleUnblock}>Bỏ chặn mục đã chọn</button>
        <button on:click={handleClearBlocklist} style="background-color: #d9534f">Xóa toàn bộ danh sách chặn</button>
        <span class="status-message">{unblockStatus}</span>
      </div>
      <div id="blocklist-items">
        {#each blocklist as app, i}
          <div>
            <input type="checkbox" id={`blocked-app-${i}`} />
            {app}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  /* Add component-specific styles here */
</style>
