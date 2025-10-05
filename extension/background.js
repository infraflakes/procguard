const hostName = 'com.nixuris.procguard';
let port;

function connect() {
  console.log('Connecting to native host...');
  port = chrome.runtime.connectNative(hostName);

  port.onMessage.addListener((msg) => {
    console.log('Received message from native host:', msg);
  });

  port.onDisconnect.addListener(() => {
    if (chrome.runtime.lastError) {
      console.log(`Disconnected due to an error: ${chrome.runtime.lastError.message}`);
    }
    console.log('Disconnected from native host. Reconnecting in 5 seconds...');
    setTimeout(connect, 5000);
  });

  // Send a ping to the host to check the connection
  port.postMessage({ type: 'ping', payload: 'hello from extension' });
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

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  // Only log when the tab is fully loaded and has a valid URL.
  if (changeInfo.status === 'complete' && tab.url && (tab.url.startsWith('http') || tab.url.startsWith('https'))) {
    if (port) {
      console.log(`Logging URL: ${tab.url}`);
      port.postMessage({ type: 'log_url', payload: tab.url });
    }
  }
});
