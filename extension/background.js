const hostName = 'com.nixuris.procguard';
let port;
let webBlocklist = [];

function connect() {
  console.log('Connecting to native host...');
  port = chrome.runtime.connectNative(hostName);

  port.onMessage.addListener((msg) => {
    console.log('Received message from native host:', msg);
    if (msg.type === 'web_blocklist') {
      webBlocklist = msg.payload || [];
      console.log('Updated web blocklist:', webBlocklist);
    }
  });

  port.onDisconnect.addListener(() => {
    if (chrome.runtime.lastError) {
      console.log(`Disconnected due to an error: ${chrome.runtime.lastError.message}`);
    }
    console.log('Disconnected from native host. Reconnecting in 5 seconds...');
    setTimeout(connect, 5000);
  });

  // Request the blocklist on connection.
  port.postMessage({ type: 'get_web_blocklist' });
}

connect();

// Listen for messages from the web GUI for installation detection.
chrome.runtime.onMessageExternal.addListener((request, sender, sendResponse) => {
  if (request.message === 'is_installed') {
    sendResponse({ status: 'installed', version: chrome.runtime.getManifest().version });
  }
  // Return true to indicate you wish to send a response asynchronously
  return true;
});

// Listen for messages from the popup.
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.type === 'add_to_web_blocklist') {
    if (port) {
      port.postMessage(request);
      // Also add it to the local blocklist immediately for a faster response.
      if (!webBlocklist.includes(request.payload)) {
        webBlocklist.push(request.payload);
      }
      sendResponse({ status: 'ok' });
    }
  }
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  // Check if the site is in the blocklist.
  if (tab.url) {
    try {
      const url = new URL(tab.url);
      const domain = url.hostname;
      if (webBlocklist.includes(domain)) {
        chrome.tabs.update(tabId, { url: chrome.runtime.getURL('blocked.html') });
        return;
      }
    } catch (e) {
      // Ignore invalid URLs
    }
  }

  // Only log when the tab is fully loaded and has a valid URL.
  if (changeInfo.status === 'complete' && tab.url && (tab.url.startsWith('http') || tab.url.startsWith('https'))) {
    if (port) {
      console.log(`Logging URL: ${tab.url}`);
      port.postMessage({ type: 'log_url', payload: tab.url });
    }
  }
});
