(function() {
  function quote(str) {
    str = str.replace(/&/g, '&amp;');
    str = str.replace(/>/g, '&gt;');
    return str.replace(/</g, '&lt;');
  }

  function ansiToRGB(str) {
    str = str.replace('\033[31m', '<span style="color: rgb(239, 100, 135);">'); // Red
    str = str.replace('\033[32m', '<span style="color: rgb(94, 202, 137);">'); // Green
    str = str.replace('\033[33m', '<span style="color: rgb(253, 216, 119);">'); // Yellow
    str = str.replace('\033[34m', '<span style="color: rgb(101, 174, 247);">'); // Blue
    str = str.replace('\033[35m', '<span style="color: rgb(170, 127, 240);">'); // Magenta
    str = str.replace('\033[36m', '<span style="color: rgb(67, 193, 190);">'); // Cyan
    return str.replace('\033[0m', '</span>');
  }

  function add(str) {
    var atBottom = Math.ceil(window.innerHeight + window.scrollY) >= document.body.offsetHeight;
    document.querySelector('pre').innerHTML += ansiToRGB(quote(str));
    if (atBottom) {
      window.scrollTo(0, document.body.scrollHeight);
    }
  }

  window.addEventListener('load', (e) => {
    var lastPongMessage = 0;
    var lastMessage = 0;

    function connect() {
      var url = new URL('/logs', window.location.href);
      url.protocol = url.protocol.replace('http', 'ws');
      var ws = new WebSocket(url);
      var statusCode = 0;

      ws.addEventListener('open', () => {
        var h = setInterval(() => {
          if ((Date.now() - lastMessage) > 60000) {
            clearInterval(h);
            ws.close();
          }
        }, 6000);
      });

      ws.addEventListener('message', (e) => {
        if (statusCode === 0) {
          statusCode = parseInt(e.data.split(' ')[0]);

          if (statusCode === 401) {
            mtoken.show();
            ws.close();
          }
        }
        else {
          lastMessage = Date.now();

          if (e.data != 'PING') {
            add(e.data) 
          }
          if ((lastMessage - lastPongMessage) >= 20000) {
            lastPongMessage = lastMessage;
            ws.send('PONG');
          }
        }
      });

      ws.addEventListener('error', (e) => {
        add('Unable to Connect');
        ws.close();
      });

      ws.addEventListener('close', () => {
        if (statusCode === 200) {
          add('Connection Closed')
        }
      });
    }

    var mtoken = new bootstrap.Modal(document.getElementById('mtoken'), { keyboard: false });
    document.getElementById('mtoken').addEventListener('shown.bs.modal', (e) => {
      document.getElementById('token').focus();
    });

    document.getElementById('ftoken').addEventListener('submit', (e) => {
      document.cookie = 'Authentication-Token=' + document.getElementById('token').value + '; max-age=86400; path=/';
      document.getElementById('token').value = '';
      e.preventDefault();
      mtoken.hide();
      connect();
    });

    connect();
  });
})();
