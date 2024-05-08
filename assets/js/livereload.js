let conn = new WebSocket("ws://" + document.location.host + "/ws");  
conn.onclose = function (evt) {  
  console.log("Connection Closed")  
  setTimeout(function () {  
    location.replace(location.href);  
  }, 2000);  
};
