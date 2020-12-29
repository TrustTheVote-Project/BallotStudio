(function(){
  // TODO: copy in type_seq from data/type_seq.json
  var type_seq = {
  "ElectionResults.BallotMeasureContest": "bmcont",
  "ElectionResults.BallotMeasureSelection": "bmsel",
  "ElectionResults.Candidate": "candidate",
  "ElectionResults.CandidateContest": "ccont",
  "ElectionResults.CandidateSelection": "csel",
  "ElectionResults.Header": "header",
  "ElectionResults.Office": "office",
  "ElectionResults.Party": "party",
  "ElectionResults.Person": "person",
  "ElectionResults.ReportingUnit": "gpunit"
  };
  var unki = 0;
  var seqForAttype = function(attype) {
    var seq = type_seq[attype];
    if (seq == undefined) {
      unki += 1;
      seq = 'unk' + unki + '_';
      console.log('no seq for attype=' + attype + ', starting ' + seq);
    }
    return seq;
  };
  var electionid = null;
  var urls = null;
  (function() {
    var eidd = document.getElementById('electionid');
    if (eidd) {
      electionid = parseInt(eidd.getAttribute('data-id')) || 0;
    }
    var urlsd = document.getElementById('urls');
    if (urlsd) {
      urls = JSON.parse(decodeURIComponent(urlsd.getAttribute("data-urls")));
    }
  })();

    var addClass = function(elem, classname) {
	var cl = elem.classList;
	if (cl) {
	    if (cl.contains(classname)) {return;}
	    cl.add(classname);
	}
    };
    var rmClass = function(elem, classname) {
	if (!hasClass(elem, classname)) {
	    return;
	}
	var cl = elem.classList;
	if (!cl) {
	    return;
	}
	cl.remove(classname);
    };
    var hasClass = function(elem, classname) {
	var cl = elem.classList;
	if (cl) {
	    return cl.contains(classname);
	}
	if (elem.className == classname) {
	    return true;
	}
	return false;
    };
    var firstChildOfClass = function(elem, classname) {
	if (hasClass(elem, classname)){
	    return elem;
	}
	for (var i = 0; elem.children[i]; i++) {
	    var out = firstChildOfClass(elem.children[i], classname);
	    if (out) {
		return out;
	    }
	}
	return null;
    };
  var allChildrenOfClass = function(elem, classname, out) {
    if (!out) {
      out = [];
    }
    if (hasClass(elem, classname)){
      out.push(elem);
      return out;
    }
    for (var i = 0; elem.children[i]; i++) {
      out = allChildrenOfClass(elem.children[i], classname, out);
    }
    return out;
  };
    var sequenceCounters = {};
    var usedSeqIds = {};
    var claimSeq = function(seqName, atid) {
	var sv = usedSeqIds[seqName];
	if (!sv) {
	    sv = {};
	    usedSeqIds[seqName] = sv;
	}
	sv[atid] = true;
    };
    var atidTaken = function(seqName, atid) {
	var sv = usedSeqIds[seqName];
	if (!sv) {
	    return false;
	}
	return sv[atid];
    }
    var seq = function(seqName) {
	var sv = sequenceCounters[seqName];
	if (!sv) {
	    sv = 0;
	}
	while (true) {
	    sv++;
	    sequenceCounters[seqName] = sv;
	    var atid = seqName + sv;
	    if (!atidTaken(seqName, atid)) {
		claimSeq(seqName, atid);
		return atid;
	    }
	}
    };
    var tmplForAttype = {};
    (function(){
	// gather templates and their metadata from the html
	var templates = document.getElementsByTagName("template");
	for (var i = 0, te; te = templates[i]; i++) {
	    var attype = te.getAttribute("data-attype");
	    if (attype) {
	      tmplForAttype[attype] = {
		"tmpl":te,
		"seq":seqForAttype(attype)
	      };
	    } else {
	      console.log("template without data-attype ", te);
	    }
	}
    })();
    var expandSubTmpl = function(elem) {
	var tmplname = elem.getAttribute("data-tmpl");
	if (tmplname) {
	    var tmpl = document.getElementById(tmplname);
	    var pti = document.importNode(tmpl.content, true);
	    elem.appendChild(pti);
	    return;
	}
	if (!elem.children) {
	    return;
	}
	for (var i = 0, x; x = elem.children[i]; i++) {
	    expandSubTmpl(x);
	}
    };
    // Used in loading from stored spec
    var newForAtType = function(attype) {
	var ti = tmplForAttype[attype];
	if (!ti) {
	    return null;
	}
	//var tmpl = document.getElementById(ti.tmpl);
	var pti = document.importNode(ti.tmpl.content, true);
	return pti;
    };

    // onclick handler for <button class="newrec">
    var newrecDo = function() {
	var tmplId = this.getAttribute("data-btmpl");
      var seqName = this.getAttribute("data-seq");
	var pn = this.parentNode;
	while (pn) {
	    var targetelem = firstChildOfClass(pn, "arraygroup");
	    if (targetelem) {
		var tmpl = document.getElementById(tmplId);
		var pti = document.importNode(tmpl.content, true);
		if (seqName) {
		    pushOb(pti, {"@id":seq(seqName)});
		}
		targetelem.appendChild(pti);
		expandSubTmpl(targetelem.lastElementChild);
		updateDeleteButtons();
		return;
	    }
	    pn = pn.parentNode;
	}
    };

    var gatherArray = function(elem) {
	var ob = [];
	for (var i = 0; elem.children[i]; i++) {
	    var to = gatherJson(elem.children[i]);
	    if (to === undefined || to == null) {
		continue;
	    }
	    ob.push(to)
	}
	return ob;
    }
    var whitespace = " \t\r\n";
  var extractQuoted = function(t,startpos,out) {
    // extract quoted substring
    var i = startpos;
    i++;
    if (i==t.length){return out;} // garbage, quit. TODO: raise exception
    var qstart = i;
    while (i < t.length) {
      c = t[i];
      if (c == '"') {
	if((i+1 < t.length) && (t[i+1] == '"')) {
	  // skip "" escaped "
	  i = i+1;
	} else {
	  // end quoted region, unescaped contained quotes, save
	  out.push(t.substring(qstart,i).replace("\"\"", "\""));
	  // consume extra whitespace
	  i++;
	  while (i < t.length) {
	    if (!whitespace.includes(t[i])) {
	      break;
	    }
	    i++;
	  }
	  return i;
	}
      }
      i++;
    }
  };
    // split like CSV
    var textToArray = function(t) {
	if (!t) {return null;}
	var mode=0;
	var i = 0;
	var fstart = i;
	var out = [];
	while (i < t.length) {
	    var c = t[i];
	    if (c == '"') {
		i = extractQuoted(t, i, out);
		continue;
	    }
	    if (c == ',') {
		out.push(t.substring(fstart,i));
		i++;
		fstart = i;
		continue;
	    }
	    i++;
	}
	if (i != fstart) {
	    out.push(t.substring(fstart,i));
	}
	return out;
    };
    var commaEscape = function(t) {
	if (t.includes(",")) {
	    return "\"" + t.replace("\"", "\"\"") + "\""
	}
	return t
    }
    var arrayToText = function(ta) {
	if (!ta || (!ta.length)) {
	    return null;
	}
	var out = commaEscape(ta[0]);
	for (var i = 1, t; t = ta[i]; i++) {
	    out += "," + commaEscape(t);
	}
	return out;
    };
  var gatherJson = function(elem) {
    var ob = {};
    var any = false
    for (var i = 0, te; te = elem.children[i]; i++) {
      if (te.tagName == "INPUT") {
	var fieldname = te.getAttribute("data-key");
	if (fieldname == "attype") {
	  fieldname = "@type";
	} else if (fieldname == "atid") {
	  fieldname = "@id";
	}
	var fv = te.value;
	if (te.type == "checkbox") {
	  fv = te.checked;
	} else if (te.type == "number") {
	  fv = te.valueAsNumber;
	}
	if (fv) {
	  ob[fieldname] = fv;
	  any = true;
	}
      } else if (te.tagName == "TEXTAREA") {
	var fieldname = te.getAttribute("data-key");
	var fv = te.value;
	if (fv) {
	  if (te.getAttribute('data-mode') == 'array') {
	    fv = textToArray(fv);
	  }
	  ob[fieldname] = fv;
	  any = true;
	}
      } else if (hasClass(te, "arraygroup")) {
	var fieldname = te.getAttribute("data-name");
	ob[fieldname] = gatherArray(te);
	any = true;
      } else if (hasClass(te, "erobject")) {
	var fieldname = te.getAttribute("data-name");
	var fv = gatherJson(te);
	if (fv) {
	  ob[fieldname] = fv;
	  any = true;
	}
      } else if (hasClass(te, "idreflist")) {
	var fieldname = te.getAttribute("data-key");
	var they = allChildrenOfClass(te, "idref");
	if (they.length) {
	  var ar = [];
	  for (var ri = 0, r; r = they[ri]; ri++) {
	    ar.push(r.getAttribute("data-atid"));
	  }
	  if (te.getAttribute("data-one")) {
	    if (ar.length > 0){
	      ob[fieldname] = ar[0];
	      if (ar.length > 1) {
		consol.log("idreflist "+fieldname+" one but got " + ar.length);
	      }
	    }
	  } else {
	    ob[fieldname] = ar;
	  }
	  any = true;
	}
      } else {
	var to = gatherJson(te);
	if (to != null) {
	  for (var k in to) {
	    ob[k]=to[k];
	    any = true;
	  }
	}
      }
    }
    if (any) {
      //console.log("gathered ", ob);
      return ob;
    } else {
      return null;
    }
  };
    // Recursively search election tree for typed objects, return {attype:[],...}
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
  var obcache = null;
  var obtcache = null; // { @type: [{},...], ... }, list per @type
  var obidcache = null; // { @id: {}, ... }, lookup by @id
  var obcachet = 0;
  var ensureObCaches = function() {
    var now = Date.now();
    if (!obcache || ((now - obcachet) > 5000)) {
      // TODO: obcache=null on input de-focus to invalidate the cache
      obcache = gatherJson(document.body);
      obcachet = now;
      obtcache = null;
      obidcache = null;
    }
    if (!obtcache) {
      obtcache = {};
      obidcache = {};
      obsByType(obcache, obtcache, obidcache);
    }
  };
  // search known ElectionReport objects for text
  var searchObs = function(attype, text) {
    text = text.toLowerCase();
    ensureObCaches();
    var they = obtcache[attype];
    if (!they){return null;}
    var matches = [];
    for (var i = 0, e; e = they[i]; i++) {
      for (const key in e) {
	if (key == '@id' || key == '@type') {
	  continue
	}
	var v = e[key];
	if (v instanceof String || (typeof(v) == "string")) {
	  v = v.toLowerCase();
	  if (v.includes(text)) {
	    matches.push(e);
	    break; // done with e
	  }
	} else if (v instanceof Array) {
	  var hit = false;
	  for (var vi = 0, vv; vv = v[vi]; vi++) {
	    if ((vv instanceof String) || (typeof(vv) == "string")) {
	      vv = vv.toLowerCase();
	      if (vv.includes(text)) {
		matches.push(e);
		hit = true;
		break;
	      }
	    }
	  }
	  if (hit) {
	    break; // done with e
	  }
	}
      }
    }
    return matches;
  };

    var isEmpty = function(ob) {
	for (var k in ob) {
	    return false;
	}
	return true;
    };

    var reportExtra = function(ob, obtype, obid, handled){
	var extraKeys = [];
	var hd = {};
	for (var i = 0, hv; hv = handled[i]; i++) {
	    hd[hv]=1;
	    if (hv == '@id') {
		hd['atid']=1;
	    } else if (hv == '@type') {
		hd['attype']=1;
	    }
	}
	for (var obk in ob) {
	    if (!hd[obk]) {
		extraKeys.push(obk);
	    }
	}
	if (extraKeys.length) {
	    console.log("pushOb {" + obtype + ", " + obid + "} still left with keys: "+extraKeys.join(", "));
	}
    };
    var pushOb = function(elem, ob, shouldConsume, handled){
	if (!elem.children) {
	    console.log("pushOb cannot descend ", elem)
	    return false;
	}
	var obtype = ob["@type"] || "unktype";
	var obid = ob["@id"] || "";
	//var handled = [];
	if (handled === undefined) {
	    handled = [];
	}
	for (var i = 0, te; te = elem.children[i]; i++) {
	    if (hasClass(te, "arraygroup")) {
		var fieldname = te.getAttribute("data-name");
		var av = ob[fieldname];
		if (av === undefined || av == null || (!av.length)) {
		    console.log("no data for ", av);
		    continue;
		}
		for (var j = 0, jv; jv = av[j]; j++) {
		    var je = te.children[j];
		    if (je === undefined) {
			var attype = jv["@type"];
			if (attype) {
			  je = newForAtType(attype);
			  if (je == null) {
				console.log("nothing for @type ", attype);
			  } else {
			    te.appendChild(je);
			    je = te.children[j];
			    expandSubTmpl(je);
			    var atid = jv["@id"];
			    if (atid) {
				var ti = tmplForAttype[attype];
				claimSeq(ti.seq, atid);
			    }
			  }
			} else {
			  console.log("don't know how to add value ", jv);
			    continue;
			}
		    }
		    pushOb(je, jv, true);
		}
		handled.push(fieldname);
	    } else if (hasClass(te, "erobject")) {
		var fieldname = te.getAttribute("data-name");
		var av = ob[fieldname];
		if (!av) {
		    continue;
		}
		handled.push(fieldname);
		for (var j = 0, je; je = te.children[j]; j++) {
		    pushOb(je, av, true);
		}
	    } else if (hasClass(te, "idreflist")) {
	      var fieldname = te.getAttribute("data-key");
	      var refs = ob[fieldname];
	      if (refs) {
		if (!(refs instanceof Array)) {
		  refs = [refs];
		}
		for (var ri = 0, r; r = refs[ri]; ri++) {
		  var nir = document.createElement("SPAN");
		  nir.className = "idref";
		  nir.setAttribute("data-atid", r);
		  var rob = obidcache[r];
		  if (rob) {
		    var summary = getAutocompleteSummarizer(rob["@type"])(rob);
		    nir.innerHTML = summary + delxHTML;
		  } else {
		    nir.innerHTML = r + delxHTML;
		  }
		  te.appendChild(nir);
		}
		handled.push(fieldname);
	      }
	    } else if (te.tagName == "INPUT") {
		var fieldname = te.getAttribute("data-key");
		if (fieldname == "attype") {
		    fieldname = "@type";
		} else if (fieldname == "atid") {
		    fieldname = "@id";
		}
		var ov = ob[fieldname];
	        if (te.type == "checkbox") {
		    te.checked = Boolean(ov)
		    handled.push(fieldname);
	        } else if (ov) {
		    te.value = ov;
		    handled.push(fieldname);
		}
	    } else if (te.tagName == "TEXTAREA") {
		var fieldname = te.getAttribute("data-key");
		var ov = ob[fieldname];
	      if (ov) {
		// if (te.getAttribute('data-mode') == 'array') {
		    if (ov instanceof Array) {
		      ov = arrayToText(ov);
		    } else if (te.getAttribute('data-mode') != 'array') {
		      // nothing. this is fine.
		    } else {
			console.log("TODO: unpack " + fieldname + "=" +JSON.stringify(ov)+ " into TEXTAREA")
		    }
		    te.value = ov;
		    handled.push(fieldname);
		}
	    } else if (hasClass(te, "showjson")) {
		setJsonShow(te, ob);
	    } else {
		pushOb(te, ob, false, handled);
	    }
	}
	if (shouldConsume) {
	    reportExtra(ob, obtype, obid, handled);// TODO: make sure this is clean
	}
	return false;
    };

    var setJsonShow = function(te, ob) {
	var hides = te.getAttribute("data-hide");
	if (hides) {
	    var hidep = hides.split(",");
	    var hided = {};
	    for (var i = 0, hi; hi = hidep[i]; i++) {
		hided[hi] = 1;
	    }
	    var nob = {};
	    for (var k in ob) {
		if (hided[k]) {
		    continue;
		}
		nob[k] = ob[k];
	    }
	    ob = nob;
	}
	te.innerHTML = JSON.stringify(ob);
    };

    var parentWithClass = function(pn, classname) {
	while (pn) {
	    if (hasClass(pn, classname)) {
		return pn;
	    }
	    pn = pn.parentNode;
	}
	return null;
    };
    var deleterec = function() {
	var pn = parentWithClass(this, "recform");
	if (pn) {
	    pn.remove();
	    // TODO: un-claim @id
	    // TODO: cleanup @id sequence, {a1,a234,a1337} -> (a1,a2,a3)
	    return;
	}
	this.remove();
    };
    var doEditMode = function() {
	var pn = parentWithClass(this, "sectionedshow");
	if (pn) {
	    var show = firstChildOfClass(pn, "show");
	    var edit = firstChildOfClass(pn, "edit");
	    if (show && edit) {
		addClass(show,"hidden");
		rmClass(edit,"hidden");
	    }
	}
    };
    var doShowMode = function() {
	var pn = parentWithClass(this, "sectionedshow");
	if (pn) {
	    var show = firstChildOfClass(pn, "show");
	    var edit = firstChildOfClass(pn, "edit");
	    if (show && edit) {
		// make sure things are up to date
		var xo = gatherJson(edit);
		pushOb(show,xo);
		// swap visibility
		rmClass(show,"hidden");
		addClass(edit,"hidden");
	    }
	}
    }
  var delxClick = function() {
    var idref = parentWithClass(this, "idref");
    idref.parentNode.removeChild(idref);
  };
    var setOnclickForClass = function(classname, fn) {
	var they = document.getElementsByClassName(classname);
	for (var i = 0, db; db = they[i]; i++) {
	    db.onclick = fn;
	}
    }

  var autocompleteSummarizers = {
    "ElectionResults.Person": function(rec) {
      if (rec.FullName) {return rec.FullName;}
      var out = rec.Prefix || "";
      if (rec.FirstName) {out += " " + rec.FirstName;}
      if (rec.Nickname) {out += " (" + rec.Nickname + ")";}
      if (rec.MiddleName) {out += " " + rec.MiddleName;}
      if (rec.LastName) {out += " " + rec.LastName;}
      if (rec.Suffix) {out += ", " + rec.Suffix;}
      return out;
    },
    "ElectionResults.Party": function(rec) {
      if (rec.Name) {return rec.Name;}
      return JSON.stringify(rec);
    },
    "ElectionResults.ReportingUnit": function(rec) {
      if (rec.Name) {return rec.Name;}
      return JSON.stringify(rec);
    },
    "ElectionResults.BallotMeasureContest": function(rec) {
      if (rec.Name) {return rec.Name;}
      return JSON.stringify(rec);
    },
    "ElectionResults.CandidateContest": function(rec) {
      if (rec.Name) {return rec.Name;}
      return JSON.stringify(rec);
    },
    "ElectionResults.Candidate": function(rec) {
      if (rec.BallotName) {return rec.BallotName;}
      // TODO: could fall through to referred person?
      return JSON.stringify(rec);
    },
    "ElectionResults.Office": function(rec) {
      if (rec.Name) {return rec.Name;}
      return JSON.stringify(rec);
    }
  };
  var getAutocompleteSummarizer = function(attype) {
    var f = autocompleteSummarizers[attype];
    if (f) {return f;}
    return JSON.stringify;
  };
  var summarize = function(rec) {
    var attype = rec['@type'];
    var sf = autocompleteSummarizers[attype] || JSON.stringify;
    return sf(rec);
  };

  var closeAutocompleteLists = function(elem) {
    // close everything except what was passed in
    var x = document.getElementsByClassName("autocomplete-items");
    for (var i = 0; i < x.length; i++) {
      if (elem != x[i]) {// TODO: if elem == current active autocomplete field, don't close, was " && elem != inp"
	x[i].parentNode.removeChild(x[i]);
      }
    }
  };
  // clicking outside an autocomplete should close them all.
  // TODO: keep the active one if we click in that
  document.addEventListener("click", function (e) {
    closeAutocompleteLists(e.target);
  });
  var delxHTML = "<img class=\"delx\" src=\""+urls.staticroot+"/delx.svg\" height=\"14\" width=\"30\">";
  var autocompleteItemClickListener = function() {
    var container = parentWithClass(this, "acgroup");
    var idreflist = firstChildOfClass(container, "idreflist");
    var nir = document.createElement("SPAN");
    nir.className = "idref";
    nir.setAttribute("data-atid", this.getAttribute("data-atid"));
    nir.innerHTML = this.innerHTML + delxHTML;
    var action = this.rootinput.getAttribute("data-action");
    if (action == "append") {
      idreflist.appendChild(nir);
    } else {
      idreflist.innerHTML = "";
      idreflist.appendChild(nir);
    }
    var delxes = allChildrenOfClass(idreflist, "delx");
    for (var di = 0, dx; dx = delxes[di]; di++) {
      dx.onclick = delxClick;
    }
    this.rootinput.value = null;
    closeAutocompleteLists();
  };
  var search_acattype = function(acattype, val) {
    // might be a comma sep list of @type-s
    if (!acattype.includes(",")) {
      return searchObs(acattype, val);
    }
    var they = acattype.split(",");
    var matches = [];
    for (var i = 0, t; t=they[i]; i++) {
      var tm = searchObs(t, val);
      if(tm){
	matches = matches.concat(tm);
      }
    }
    return matches;
  };
  var autocompleteInputListener = function(rootinput, e, ctx) {
    var val = rootinput.value;
    closeAutocompleteLists();
    if (!val){return false;}
    var attype = rootinput.getAttribute("data-acattype");
    if (!attype){return false;}
    var items = search_acattype(attype, val);
    if (!items || !items.length){return false;}
    ctx.currentFocus = -1;
    var itemsdiv = document.createElement("DIV");
    itemsdiv.setAttribute("id", rootinput.id + "autocomplete-list");
    itemsdiv.setAttribute("class", "autocomplete-items");
    rootinput.parentNode.appendChild(itemsdiv);
    for (var i = 0, ie; ie=items[i]; i++) {
      var adiv = document.createElement("DIV");
      adiv.innerHTML = summarize(ie);
      adiv.setAttribute("data-atid", ie["@id"]);
      adiv.rootinput = rootinput;
      adiv.addEventListener("click", autocompleteItemClickListener);
      itemsdiv.appendChild(adiv);
    }
  };
  var setActive = function(rootinput, ctx, newActive) {
    var itemsdiv = document.getElementById(rootinput.id + "autocomplete-list");
    if (!itemsdiv){return;}
    var items = itemsdiv.getElementsByTagName("div");
    if (!items){return;}
    if (newActive < 0 || newActive >= items.length) {
      return;
    }
    for (var i = 0, ie; ie = items[i]; i++) {
      if (i == newActive) {
	ie.classList.add("autocomplete-active");
	ctx.currentFocus = newActive;
      } else {
	ie.classList.remove("autocomplete-active");
      }
    }
  };
  var autocompleteKeydownListener = function(rootinput, e, ctx) {
    if (e.keyCode == 13){// enter
      e.preventDefault();
      var itemsdiv = document.getElementById(rootinput.id + "autocomplete-list");
      if (itemsdiv) {
	if (ctx.currentFocus > -1) {
	  var items = itemsdiv.getElementsByTagName("div");
	  items[ctx.currentFocus].click();
	}
      } else {
	// TODO: accept what's typed in as literal though it may be wrong?
	// TODO: show warning "no known @type ..."
      }
    } else if (e.keyCode == 40) {// down arrow
      setActive(rootinput, ctx, ctx.currentFocus + 1);
    } else if (e.keyCode == 38) {// up arrow
      setActive(rootinput, ctx, ctx.currentFocus - 1);
    }
  };
  var idcounter = 1;
  var autocompleteElementClousureContext = function(elem) {
    var ctx={};
    if (!elem.id){elem.id = "i" + idcounter;idcounter++;}
    elem.addEventListener("input", function(e){
      return autocompleteInputListener(this, e, ctx);
    });
    elem.addEventListener("keydown", function(e){
      return autocompleteKeydownListener(this, e, ctx);
    });
  };
  var enableAutocompletes = function() {
    var they = document.getElementsByClassName("acsearch");
    for (var i = 0, elem; elem = they[i]; i++) {
      autocompleteElementClousureContext(elem);
    }
  };

  var updateDeleteButtons = function() {
    setOnclickForClass("deleterec",deleterec);
    setOnclickForClass("newrec",newrecDo);
    setOnclickForClass("sectionedit",doEditMode);
    setOnclickForClass("sectionshow",doShowMode);
    setOnclickForClass("delx",delxClick);
    enableAutocompletes();
  };
  updateDeleteButtons();

    var drawResultHandler = function() {
	if (this.readyState == 4 && this.status == 200) {
	    var html = document.getElementById("debugdiv").innerHTML;
	    var ob = JSON.parse(this.responseText);
	    if (ob && ob.item) {
		html += "<p style=\"font-size:200%\"><a href=\"/item?i=" + ob.item + "\">PDF</a></p>";
	    }
	    html += "<pre>" + this.responseText + "</pre>";
	    document.getElementById("debugdiv").innerHTML = html;
	}
    };

    var debugbutton = document.getElementById("debugbutton");
    debugbutton.onclick = function() {
	var js = gatherJson(document.body);
	document.getElementById("debugdiv").innerHTML = "<pre>" + JSON.stringify(js) + "</pre>";
    };

  var demobutton = document.getElementById("demobutton");
  demobutton.onclick = function() {
    GET('/static/demoelection.json', loadElectionHandler);
  };

    var saveResultHandler = function(buttonelem, http) {
	var dbt = null;
	if (http.readyState >= 2) {
	    var dbt = firstChildOfClass(buttonelem.parentNode, "debugtext");
	    if (http.readyState == 2 || http.readyState == 3) {
		if (dbt) {
		    dbt.innerHTML = "<span style=\"background-color:#ffa;font-weight:bolt;font-size:120%;\">saving...</span>";
		}
	    } else if (http.readyState == 4) {
		// done
		if (http.status == 200 && !electionid) {
		  var response = JSON.parse(http.responseText);
		  if ((!urls) || (urls.edit != response.edit)) {
		    window.location = window.location.protocol + '//' + window.location.host + response.edit;
		    return;
		  }
		  electionid = response.itemid;
		  urls = response;
		}
		if (dbt) {
		    if (http.status == 200) {
			dbt.innerHTML = "saved <a href=\"/edit/" + electionid + "\">election " + electionid + "</a> at " + Date();
		    } else {
			var msg = "error: " + http.status + " " + http.statusText;
			dbt.innerHTML = msg;
		    }
		}
	    }
	}
    };
    var savebuttonclick = function() {
	var js = gatherJson(document.body);
	var eid = electionid || 0;
	var savebutton = this;
	POSTjson(urls.post, js, function(){saveResultHandler(savebutton, this);});
    };
    setOnclickForClass("savebutton",savebuttonclick);

    var loadElectionHandler = function() {
      if (this.readyState == 4 && this.status == 200) {
	obcache = JSON.parse(this.responseText);
	obcachet = Date.now();
	obtcache = null;
	ensureObCaches();
	pushOb(document.body, obcache);
	updateDeleteButtons();
	var they = document.getElementsByClassName("sectionshow");
	for (var i = 0, db; db = they[i]; i++) {
	  db.onclick();
	}
      }
    };
    var GET = function(url, handler) {
	var http = new XMLHttpRequest();
	http.onreadystatechange = handler;
	http.timeout = 9000;
	http.open("GET",url,true);
	http.send();
    };
    //GET("/static/demo.json", loadElectionHandler);
    var POSTjson = function(url, ob, handler) {
	var data = JSON.stringify(ob);
	POST(url, data, 'application/json', handler);
    };
    var POST = function(url, data, contentType, handler) {
	var http = new XMLHttpRequest();
	http.timeout = 9000;
	http.onreadystatechange = handler;
	http.open("POST",url,true);
	http.setRequestHeader('Content-Type', contentType);
	http.send(data);
    };
    //pushOb(document.body, savedObj);
    (function(){
      if (urls && urls.url) {
	GET(urls.url, loadElectionHandler);
      }
    })();
})();
