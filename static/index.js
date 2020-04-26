(function(){
    var hasClass = function(elem, classname) {
	var cl = elem.classList;
	if (cl) {
	for (var i = 0; cl[i]; i++) {
	    if (classname == cl[i]) {
		return true;
	    }
	}
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
    var tmplForAttype = {
	"ElectionResults.Party":{"tmpl":"partytmpl", "seq":"party"},
	"ElectionResults.Person":{"tmpl":"persontmpl", "seq":"pers"},
	"ElectionResults.Office":{"tmpl":"officetmpl", "seq":"office"},
    };
    // Used in loading from stored spec
    var newForAtType = function(attype) {
	var ti = tmplForAttype[attype];
	if (!ti) {
	    return null;
	}
	var tmpl = document.getElementById(ti.tmpl);
	var pti = document.importNode(tmpl.content, true);
	return pti;
    };
    // Used when a [new thing] button is pressed by a user
    var instantiate = function(tmplId, sectionId, seqName) {
	var tmpl = document.getElementById(tmplId);
	var pti = document.importNode(tmpl.content, true);
	var atidelem = firstChildOfClass(pti, "atid");
	var atid = seq(seqName);
	atidelem.value = atid;
	var section = document.getElementById(sectionId);
	section.appendChild(pti);
    };
    document.getElementById("newparty").onclick = function(){instantiate("partytmpl", "parties", "party");};
    document.getElementById("newperson").onclick = function(){instantiate("persontmpl", "people", "pers");};
    document.getElementById("newoffice").onclick = function(){instantiate("officetmpl", "offices", "office");};

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
    var gatherJson = function(elem) {
	var ob = {};
	var any = false
	for (var i = 0, te; te = elem.children[i]; i++) {
	    if (te.tagName == "INPUT") {
		var fieldname = te.className;
		if (fieldname == "attype") {
		    fieldname = "@type";
		} else if (fieldname == "atid") {
		    fieldname = "@id";
		}
		ob[fieldname] = te.value;
		any = true;
	    } else if (hasClass(te, "arraygroup")) {
		var fieldname = te.getAttribute("data-name");
		ob[fieldname] = gatherArray(te);
		any = true;
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

    var pushOb = function(elem, ob){
	if (!elem.children) {
	    console.log("pushOb cannot descend ", elem)
	    return false;
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
		    pushOb(je, jv);
		}
		delete ob[fieldname];
		if (isEmpty(ob)) {
		    return true;
		}
	    } else if (te.tagName == "INPUT") {
		var fieldname = te.className;
		if (fieldname == "attype") {
		    fieldname = "@type";
		} else if (fieldname == "atid") {
		    fieldname = "@id";
		}
		var ov = ob[fieldname];
		if (ov) {
		    te.value = ov;
		    delete ob[fieldname];
		    if (isEmpty(ob)) {
			return true;
		    }
		}
	    } else {
		var consumed = pushOb(te, ob);
		if (isEmpty(ob)) {
		    return true;
		}
	    }
	}
	return false;
    };

    var debugbutton = document.getElementById("debugbutton");
    debugbutton.onclick = function() {
	var js = gatherJson(document.body);
	document.getElementById("debugdiv").innerHTML = JSON.stringify(js);
    };

    var savedObj = {"Party":[{"@id":"party1","@type":"ElectionResults.Party","Name":"Stupid","Abbreviation":"","Color":"","IsRecognizedParty":"on","LogoUri":"","Slogan":""},{"@id":"party2","@type":"ElectionResults.Party","Name":"Evil","Abbreviation":"","Color":"","IsRecognizedParty":"on","LogoUri":"","Slogan":""}],"Person":[{"@id":"pers1","@type":"ElectionResults.Person","FullName":"SOMEGUY","Prefix":"","FirstName":"","MiddleName":"","LastName":"","Suffix":"","Nickname":"","Title":"","Profession":"","DateOfBirth":""}],"Office":[]};
    pushOb(document.body, savedObj);
})();
