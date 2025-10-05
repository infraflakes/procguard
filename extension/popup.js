document.addEventListener('DOMContentLoaded', () => {
  const statusElement = document.getElementById('status');
  // In the future, we will get the status from the background script.
  statusElement.textContent = 'Ready';
});
