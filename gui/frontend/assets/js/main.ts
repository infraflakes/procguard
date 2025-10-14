const topLevelViews: string[] = [
  'welcome-view',
  'app-management-view',
  'web-management-view',
  'settings-view',
];
const appSubViews: string[] = ['search-view', 'blocklist-view'];
const webSubViews: string[] = ['web-log-view', 'web-blocklist-view'];

declare function loadAutostartStatus(): void;
declare function loadBlocklist(): void;
declare function loadWebBlocklist(): void;
declare function loadWebLogs(): void;

const sinceDateInput = document.getElementById(
  'since_date'
) as HTMLInputElement;
const untilDateInput = document.getElementById(
  'until_date'
) as HTMLInputElement;
const webSinceDateInput = document.getElementById(
  'web_since_date'
) as HTMLInputElement;
const webUntilDateInput = document.getElementById(
  'web_until_date'
) as HTMLInputElement;
const extensionStatus = document.getElementById(
  'extension-status'
) as HTMLSpanElement;
const installExtensionBtn = document.getElementById(
  'install-extension-btn'
) as HTMLButtonElement;
const EXTENSION_ID = 'ilaocldmkhlifnikhinkmiepekpbefoh';

function setDefaults(): void {
  const now = new Date();
  const year = now.getFullYear();
  const month = (now.getMonth() + 1).toString().padStart(2, '0');
  const day = now.getDate().toString().padStart(2, '0');
  const today = `${year}-${month}-${day}`;
  if (sinceDateInput) sinceDateInput.value = today;
  if (untilDateInput) untilDateInput.value = today;
  if (webSinceDateInput) webSinceDateInput.value = today;
  if (webUntilDateInput) webUntilDateInput.value = today;
}

function checkExtension(): void {
  if (typeof chrome !== 'undefined' && chrome.runtime) {
    chrome.runtime.sendMessage(
      EXTENSION_ID,
      { message: 'is_installed' },
      (response) => {
        if (chrome.runtime.lastError) {
          if (extensionStatus) extensionStatus.textContent = 'Chưa cài đặt';
          if (installExtensionBtn) installExtensionBtn.style.display = 'block';
        } else {
          if (response && extensionStatus) {
            extensionStatus.textContent = `Đã cài đặt (v${response.version})`;
          }
          if (installExtensionBtn) installExtensionBtn.style.display = 'none';
        }
      }
    );
  } else {
    if (extensionStatus)
      extensionStatus.textContent = 'Không phải trình duyệt dựa trên Chrome';
  }
}

function showTopLevelView(viewName: string): void {
  topLevelViews.forEach((id) => {
    const el = document.getElementById(id);
    if (el) el.style.display = 'none';
  });
  const el = document.getElementById(viewName);
  if (el) el.style.display = 'block';

  if (viewName === 'app-management-view') {
    showSubView('search-view', 'app-management-view');
  } else if (viewName === 'web-management-view') {
    showSubView('web-log-view', 'web-management-view');
  } else if (viewName === 'settings-view') {
    loadAutostartStatus();
  }
}

function showSubView(viewName: string, parentView: string): void {
  let subviews: string[];
  let tabContainerId: string;

  if (parentView === 'app-management-view') {
    subviews = appSubViews;
    tabContainerId = 'appManTabs';
  } else if (parentView === 'web-management-view') {
    subviews = webSubViews;
    tabContainerId = 'webManTabs';
  } else {
    return;
  }

  // --- Handle Button Highlighting ---
  const tabContainer = document.getElementById(tabContainerId);
  if (tabContainer) {
    const tabButtons = tabContainer.querySelectorAll('.nav-link');
    tabButtons.forEach(button => button.classList.remove('active'));
    
    const buttonId = viewName.replace('-view', '-tab');
    const activeButton = document.getElementById(buttonId);
    if (activeButton) activeButton.classList.add('active');
  }

  // --- Handle Content Visibility (using original display:none/block logic) ---
  subviews.forEach((id) => {
    const el = document.getElementById(id);
    if (el) el.style.display = 'none';
  });
  const el = document.getElementById(viewName);
  if (el) el.style.display = 'block';

  // --- Keep the original data loading logic ---
  if (viewName === 'blocklist-view') {
    loadBlocklist();
  } else if (viewName === 'web-blocklist-view') {
    loadWebBlocklist();
  } else if (viewName === 'web-log-view') {
    loadWebLogs();
  }
}

document.addEventListener('DOMContentLoaded', () => {
  setDefaults();
  showTopLevelView('welcome-view');
  checkExtension();
});
