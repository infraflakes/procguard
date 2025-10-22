const hostName = 'com.nixuris.procguard';
let port;
let webBlocklist = [];

function connect() {
  try {
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
  } catch (err) {
    console.error('Error connecting to native host:', err);
    // Still try to reconnect
    console.log('Reconnecting in 5 seconds...');
    setTimeout(connect, 5000);
  }
}

connect();

// Listen for messages from the web GUI for installation detection.
chrome.runtime.onMessageExternal.addListener((request, sender, sendResponse) => {
  if (request.message === 'is_installed') {
    sendResponse({
      status: 'installed',
      version: chrome.runtime.getManifest().version,
      id: chrome.runtime.id,
    });
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
  } else if (request.type === 'log_web_metadata') {
    if (port) {
      port.postMessage(request);
    }
  }
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  // Inject extension ID into the web UI
  if (
    changeInfo.status === 'complete' &&
    tab.url &&
    tab.url.startsWith('http://127.0.0.1:58141')
  ) {
    chrome.scripting.executeScript({
      target: { tabId: tabId },
      func: () => {
        const extensionId = chrome.runtime.id;
        // Remove existing div if it's there
        const existingDiv = document.getElementById('procguard-extension-id');
        if (existingDiv) {
          existingDiv.remove();
        }
        const idDiv = document.createElement('div');
        idDiv.id = 'procguard-extension-id';
        idDiv.textContent = extensionId;
        idDiv.style.display = 'none';
        document.body.appendChild(idDiv);
      },
    });
  }

  const blockedPage = chrome.runtime.getURL('blocked.html');
  if (tab.url === blockedPage) {
    return;
  }

  // Check if the site is in the blocklist.
  if (tab.url) {
    try {
      const url = new URL(tab.url);
      const domain = url.hostname;
      if (webBlocklist.includes(domain)) {
        chrome.tabs.update(tabId, { url: blockedPage });
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

      // Inject a script to get the title and favicon
      chrome.scripting.executeScript({
        target: { tabId: tabId },
        func: () => {
          function getFaviconUrl() {
            let faviconUrl = '';
            const linkElements = document.head.querySelectorAll('link[rel*="icon"]');
            for (const link of linkElements) {
              if (link.href) {
                faviconUrl = link.href;
                break;
              }
            }
            if (!faviconUrl) {
              faviconUrl = `${window.location.origin}/favicon.ico`;
            }
            return faviconUrl;
          }

          chrome.runtime.sendMessage({
            type: 'log_web_metadata',
            payload: {
              domain: window.location.hostname,
              title: document.title,
              iconUrl: getFaviconUrl()
            }
          });
        }
      });
    }
  }
});
