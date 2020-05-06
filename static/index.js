(function(){
    var addClass = function(elem, classname) {
	var cl = elem.classList;
	if (cl) {
	    if (cl.contains(classname)) {return;}
	    // for (var i = 0; cl[i]; i++) {
	    // 	if (classname == cl[i]) {
	    // 	    return
	    // 	}
	    // }
	    cl.add(classname);
	}
    };
    var rmClass = function(elem, classname) {
	if (!hasClass(elem, classname)) {
	    return;
	}
	//var ncl = new DOMTokenList();
	var cl = elem.classList;
	if (!cl) {
	    return;
	}
	cl.remove(classname);
	// for (var i = 0, ci; ci = cl[i]; i++) {
	//     if (ci != classname) {
	// 	ncl.push(ci);
	//     }
	// }
	// elem.classList = ncl;
    };
    var hasClass = function(elem, classname) {
	var cl = elem.classList;
	if (cl) {
	    return cl.contains(classname);
	    // for (var i = 0; cl[i]; i++) {
	    // 	if (classname == cl[i]) {
	    // 	    return true;
	    // 	}
	    // }
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
		var ob = {"tmpl":te}
		var seqname = te.getAttribute("data-seq");
		if (seqname) {
		    ob.seq = seqname;
		}
		tmplForAttype[attype] = ob;
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
    var textareaSplitter = /[, \t\r\n]+/g;
    var whitespace = " \t\r\n";
    var extractQuoted = function(t,startpos,out) {
		// extract quoted substring
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
			i++;
		    }
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
	if (t.include(",")) {
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
		}
		if (fv) {
		    ob[fieldname] = fv;
		    any = true;
		}
	    } else if (te.tagName == "TEXTAREA") {
		var fieldname = te.getAttribute("data-key");
		var fv = te.value;
		if (fv) {
		    ob[fieldname] = textToArray(fv);//fv.split(textareaSplitter);
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
			    te.appendChild(je);
			    je = te.children[j];
			    expandSubTmpl(je);
			    var atid = jv["@id"];
			    if (atid) {
				var ti = tmplForAttype[attype];
				claimSeq(ti.seq, atid);
			    }
			} else {
			    console.log("don't know how to add value ", jv)
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
	    } else if (te.tagName == "INPUT") {
		var fieldname = te.getAttribute("data-key");
		if (fieldname == "attype") {
		    fieldname = "@type";
		} else if (fieldname == "atid") {
		    fieldname = "@id";
		}
		var ov = ob[fieldname];
		if (ov) {
		    te.value = ov;
		    handled.push(fieldname);
		}
	    } else if (te.tagName == "TEXTAREA") {
		var fieldname = te.getAttribute("data-key");
		var ov = ob[fieldname];
		if (ov) {
		    if (ov.length) {
			ov = arrayToText(ov);
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
		//show.style.visibility = 'hidden';
		//edit.style.visibility = 'visible';
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
		//show.style.visibility = 'visible';
		//edit.style.visibility = 'hidden';
	    }
	}
    }
    var setOnclickForClass = function(classname, fn) {
	var they = document.getElementsByClassName(classname);
	for (var i = 0, db; db = they[i]; i++) {
	    db.onclick = fn;
	    //console.log(classname + " " + i);
	}
    }
    var updateDeleteButtons = function() {
	setOnclickForClass("deleterec",deleterec);
	setOnclickForClass("newrec",newrecDo);
	setOnclickForClass("sectionedit",doEditMode);
	setOnclickForClass("sectionshow",doShowMode);
	/*
	var delbuttons = document.getElementsByClassName("deleterec");
	for (var i = 0, db; db = delbuttons[i]; i++) {
	    db.onclick = deleterec;
	}
	var newbuttons = document.getElementsByClassName("newrec");
	for (var i = 0, db; db = newbuttons[i]; i++) {
	    db.onclick = newrecDo;
	}
	var editbuttons = document.getElementsByClassName("sectionedit");
	for (var i = 0, db; db = editbuttons[i]; i++) {
	    db.onclick = doEditMode;
	}
	var showbuttons = document.getElementsByClassName("sectionshow");
	for (var i = 0, db; db = showbuttons[i]; i++) {
	    db.onclick = doShowMode;
	}*/
    };
    updateDeleteButtons();

    var drawResultHandler = function() {
	if (this.readyState == 4 && this.status == 200) {
	    var html = document.getElementById("debugdiv").innerHTML;
	    html += "<pre>" + this.responseText + "</pre>";
	    var ob = JSON.parse(this.responseText);
	    if (ob && ob.item) {
		html += "<p style=\"font-size:200%\"><a href=\"/item?i=" + ob.item + "\">PDF</a></p>";
	    }
	    document.getElementById("debugdiv").innerHTML = html;
	}
    };

    var debugbutton = document.getElementById("debugbutton");
    debugbutton.onclick = function() {
	var js = gatherJson(document.body);
	document.getElementById("debugdiv").innerHTML = "<pre>" + JSON.stringify(js) + "</pre>";
	POSTjson("/draw?bubbles=1", js, drawResultHandler);
    };

    //var savedObj = {"Party":[{"@id":"party1","@type":"ElectionResults.Party","Name":"Stupid","Abbreviation":"","Color":"","IsRecognizedParty":"on","LogoUri":"","Slogan":""},{"@id":"party2","@type":"ElectionResults.Party","Name":"Evil","Abbreviation":"","Color":"","IsRecognizedParty":"on","LogoUri":"","Slogan":""}],"Person":[{"@id":"pers1","@type":"ElectionResults.Person","FullName":"SOMEGUY","Prefix":"","FirstName":"","MiddleName":"","LastName":"","Suffix":"","Nickname":"","Title":"","Profession":"","DateOfBirth":""}],"Office":[{"@id":"office1","@type":"ElectionResults.Office","Name":"Mayor","Term":{"@type":"ElectionResults.Term","StartDate":"2021-01-20","EndDate":"2025-01-20","Type":"full-term"}}]}
    var loadElectionHandler = function() {
	if (this.readyState == 4 && this.status == 200) {
	    pushOb(document.body, JSON.parse(this.responseText));
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
	http.open("GET",url,true);
	http.send();
    };
    GET("/static/demo.json", loadElectionHandler);
    var POSTjson = function(url, ob, handler) {
	var data = JSON.stringify(ob);
	POST(url, data, 'application/json', handler);
    };
    var POST = function(url, data, contentType, handler) {
	var http = new XMLHttpRequest();
	http.onreadystatechange = handler;
	http.open("POST",url,true);
	http.setRequestHeader('Content-Type', contentType);
	http.send(data);
    };
    //pushOb(document.body, savedObj);
})();
