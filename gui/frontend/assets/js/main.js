const topLevelViews = ['welcome-view', 'app-management-view', 'web-management-view', 'settings-view'];
const appSubViews = ['search-view', 'blocklist-view'];
const webSubViews = ['web-log-view', 'web-blocklist-view'];

function showTopLevelView(viewName) {
    topLevelViews.forEach(id => {
        const el = document.getElementById(id);
        if (el) el.style.display = 'none';
    });
    const el = document.getElementById(viewName);
    if (el) el.style.display = 'block';

    // Load content for the default subview if applicable
    if (viewName === 'app-management-view') {
      showSubView('search-view', 'app-management-view');
    } else if (viewName === 'web-management-view') {
      showSubView('web-log-view', 'web-management-view');
    } else if (viewName === 'settings-view') {
      loadAutostartStatus();
    }
}

function showSubView(viewName, parentView) {
  let subviews;
  if (parentView === 'app-management-view') {
    subviews = appSubViews;
  } else if (parentView === 'web-management-view') {
    subviews = webSubViews;
  }

  subviews.forEach(id => {
    const el = document.getElementById(id);
    if (el) el.style.display = 'none';
  });
  const el = document.getElementById(viewName);
  if (el) el.style.display = 'block';

  if (viewName === 'blocklist-view') {
      loadBlocklist();
  } else if (viewName === 'web-blocklist-view') {
      loadWebBlocklist();
  } else if (viewName === 'web-log-view') {
      loadWebLogs();
  }
}

document.addEventListener('DOMContentLoaded', () => {
    // Set defaults and show the welcome view by default
    setDefaults();
    showTopLevelView('welcome-view');
    checkExtension();
});
