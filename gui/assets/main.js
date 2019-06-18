(function() {
  var socket = new WebSocket('ws://localhost:18485/status');
  var progress = 0;
  var progressMax = 0;

  socket.onopen = function() {
    showMessage('Started');
  };

  socket.onclose = function(event) {
    closeWindow();
  };

  socket.onmessage = function(event) {
    var message = JSON.parse(event.data);
    if (message.type === 'close') {
      showMessage('Exiting...');
      closeWindow();
    } else if (message.type === 'progress_step') {
      progressStep();
    } else if (message.type === 'progress_max') {
      progressMax = +message.payload;
    } else if (message.type === 'title') {
      setTitle(message.payload);
    } else {
      showMessage(message.payload);
    }
  };

  socket.onerror = function(error) {
    showMessage('Error: ' + error.message);
  };

  function showMessage(message) {
    document.querySelector('div.text').innerText = message;
  }

  function closeWindow() {
    setTimeout(function () {
      window.close();
    }, 500);
  }

  function setTitle(title) {
    document.querySelector('div.app-title').innerText = title;
  }

  function progressStep() {
    progress++;
    var width = String(progress/progressMax * 100) + '%';
    document.querySelector('div.progress').style.width = width;
  }
})();
