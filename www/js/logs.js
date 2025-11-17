(function() {
  function quote(str) {
    str = str.replace(/&/g, '&amp;');
    str = str.replace(/>/g, '&gt;');
    return str.replace(/</g, '&lt;');
  }

  function recolour(str) {
    str = str.replace('\033[35m', '<span style="color: rgb(170, 127, 240);">');
    str = str.replace('\033[31m', '<span style="color: rgb(239, 100, 135);">');
    str = str.replace('\033[33m', '<span style="color: rgb(253, 216, 119);">');
    str = str.replace('\033[32m', '<span style="color: rgb(94, 202, 137);">');
    return str.replace('\033[0m', '</span>');
  }

  function add(str) {
    var atBottom = Math.ceil(window.innerHeight + window.scrollY) >= document.body.offsetHeight;
    document.getElementById('logs').innerHTML += recolour(quote(str));
    if (atBottom) {
      window.scrollTo(0, document.body.scrollHeight);
    }
  }

  window.addEventListener('load', (e) => {
    var url = new URL('/logs', window.location.href);
    url.protocol = url.protocol.replace('http', 'ws');
    var ws = new WebSocket(url);
    var lastPongMessage = 0;
    var lastMessage = 0;

    ws.addEventListener('open', () => {
      var h = setInterval(() => {
        if ((Date.now() - lastMessage) > 60000) {
          clearInterval(h);
          ws.close();
        }
      }, 6000);
    });
    ws.addEventListener('message', (e) => {
      lastMessage = Date.now();

      if (e.data != 'PING') {
        add(e.data) 
      }
      if ((lastMessage - lastPongMessage) >= 20000) {
        lastPongMessage = lastMessage;
        ws.send('PONG');
      }
    });
    ws.addEventListener('error', () => { add('Unable to Connect') });
    ws.addEventListener('close', () => { add('Connection Closed') });
  });
})();
