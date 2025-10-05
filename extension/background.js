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
