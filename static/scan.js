(function(){
  var electionid = null;
  (function() {
    var eidd = document.getElementById('electionid');
    if (eidd) {
      electionid = eidd.getAttribute('data-id');
    }
  })();
  var imf = document.getElementById("imf");
  imf.addEventListener('submit', function(e){
    e.preventDefault();
    var fi = imf.elements[0].files[0];
    POST('/scan/' + electionid, fi, fi.type, function(){imageuploadHandler(this);});
  });
  var imageuploadHandler = function(http) {
    var dbg = document.getElementById("dbg");
    if (http.readyState >= 2) {
      if (http.readyState == 2 || http.readyState == 3) {
	if (dbg) {
	  dbg.innerHTML = "<span style=\"background-color:#ffa;font-weight:bolt;font-size:120%;\">saving...</span>";
	}
      } else if (http.readyState == 4) {
	var response = JSON.parse(http.responseText);
	if (http.status == 200){
	  dbg.innerHTML = JSON.stringify(response);
	}
      }
    }
  };
  var POST = function(url, data, contentType, handler) {
    var http = new XMLHttpRequest();
    http.timeout = 9000;
    http.onreadystatechange = handler;
    http.open("POST",url,true);
    http.setRequestHeader('Content-Type', contentType);
    http.send(data);
  };
})();
