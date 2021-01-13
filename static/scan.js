(function(){
  var electionid = null;
  var electionob = null;
  var scanresult = null;
  var urls = {};
  (function() {
    var eidd = document.getElementById('electionid');
    if (eidd) {
      electionid = eidd.getAttribute('data-id');
    }
    var urlsd = document.getElementById('urls');
    if (urlsd) {
      urls = JSON.parse(decodeURIComponent(urlsd.getAttribute("data-urls")));
    }
  })();
  var imf = document.getElementById("imf");
  imf.addEventListener('submit', function(e){
    e.preventDefault();
    var fi = imf.elements[0].files[0];
    var url = (urls && urls.scan) || ('/scan/' + electionid);
    POST(url, fi, fi.type, function(){imageuploadHandler(this);});
  });
  var imageuploadHandler = function(http) {
    var dbg = document.getElementById("dbg");
    if (http.readyState >= 2) {
      if (http.readyState == 2 || http.readyState == 3) {
	if (dbg) {
	  dbg.innerHTML = "<span style=\"background-color:#ffa;font-weight:bolt;font-size:120%;\">saving...</span>";
	}
      } else if (http.readyState == 4) {
	scanresult = JSON.parse(http.responseText);
	if (http.status == 200){
	  dbg.innerHTML = JSON.stringify(scanresult);
	}
	maybeShowResults();
      }
    }
  };
  var maybeShowResults = function() {
    if (scanresult == null){return;}
    if (electionob == null){return;}
    var bytype = {};
    var byid = {};
    obsByType(electionob, bytype, byid);
    var out = "<div class=\"hyv\">How you voted:</div>";
    for (const contestid in scanresult) {
      out += "<div class=\"contestvote\">";
      var contestob = byid[contestid];
      var contestname = (contestob && contestob.BallotTitle);
      if (contestname) {
	out += "<div class=\"contestname\">" + contestname + "</div>";
      }
      out += "<div class=\"cnl\">";
      var contestBubbles = scanresult[contestid];
      for (const cselid in contestBubbles) {
	var csel = byid[cselid];
	if (csel) {
	  if (csel.IsWriteIn) {
	    out += "<span class=\"candwi\">write-in</span>";
	  } else if (csel.CandidateIds) {
	    for (var i = 0, cid; cid = csel.CandidateIds[i]; i++) {
	      var cob = byid[cid];
	      var cname = null;
	      if (cob) {
		cname = cob.BallotName;
	      } else {
		cname = cid;
	      }
	      out += "<span class=\"candname\">" + cname + "</span>";
	    }
	  } else if (csel.Selection) {
	    out += "<span class=\"candname\">" + csel.Selection + "</span>";
	  }
	}
      }
      out += "</div></div>";
    }
    document.getElementById("results").innerHTML = out;
  };
  var loadElectionHandler = function() {
    if (this.readyState == 4 && this.status == 200) {
      electionob = JSON.parse(this.responseText);
      maybeShowResults();
    }
  };
  // TODO: common
  var POST = function(url, data, contentType, handler) {
    var http = new XMLHttpRequest();
    http.timeout = 9000;
    http.onreadystatechange = handler;
    http.open("POST",url,true);
    http.setRequestHeader('Content-Type', contentType);
    http.send(data);
  };
  var GET = function(url, handler) {
    var http = new XMLHttpRequest();
    http.onreadystatechange = handler;
    http.timeout = 9000;
    http.open("GET",url,true);
    http.send();
  };
  var obsByType = function(ob, bytype, byid, seen) {
    if (!((ob instanceof Object) || (ob instanceof Array))) {
      return;
    }
    var atid = ob['@id'];
    if (atid) {
      if (byid[atid]) {
	console.log("id collision, dup: " + atid);
      } else {
	byid[atid] = ob;
      }
    }
    var attype = ob['@type'];
    if (attype) {
      var they = bytype[attype];
      if (they) {
	they.push(ob);
      } else {
	they = [ob];
	bytype[attype] = they;
      }
    }
    for (const key in ob) {
      obsByType(ob[key], bytype, byid, seen);
    }
  };
  // TODO: wait until scan submission
  (function(){
    if (urls && urls.url) {
      GET(urls.url, loadElectionHandler);
    }
  })();
})();
