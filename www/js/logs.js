(function() {
  window.addEventListener('load', (e) => {
    document.body.innerHTML += '<p>Page Loaded</p>';

    var url = new URL('/logs', window.location.href);
    url.protocol = url.protocol.replace('http', 'ws');
    var ws = new WebSocket(url);

    ws.addEventListener('open', () => {
      document.body.innerHTML += '<p>Connected</p>';
    });
    ws.addEventListener('error', () => {
      document.body.innerHTML += '<p>Error</p>';
    });
    ws.addEventListener('message', (e) => {
      document.body.innerHTML += '<p>' + e.data + '</p>';
    });
    ws.addEventListener('close', () => {
      document.body.innerHTML += '<p>Closed</p>';
    });
  });
})();
