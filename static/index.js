(function(){
    var firstChildOfClass = function(elem, classname) {
	var cl = elem.classList;
	if (cl) {
	for (var i = 0; cl[i]; i++) {
	    if (classname == cl[i]) {
		return elem;
	    }
	}
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
    var seq = function(seqName) {
	var sv = sequenceCounters[seqName];
	if ((sv === undefined) || (sv == null)) {
	    sv = 0;
	}
	sv++;
	sequenceCounters[seqName] = sv;
	return seqName + sv;
    };
    var instantiate = function(tmplId, sectionId, seqName) {
	var tmpl = document.getElementById(tmplId);
	var pti = document.importNode(tmpl.content, true);
	//var atidelem = pti.getElementsByClassName("atid")[0];
	var atidelem = firstChildOfClass(pti, "atid");
	var atid = seq(seqName);
	atidelem.setAttribute("data-atid", atid);
	atidelem.innerHTML = atid;
	var section = document.getElementById(sectionId);
	section.appendChild(pti);
    };
    var newParty = function() {
	instantiate("partytmpl", "parties", "party");
	// var partytmpl = document.getElementById("partytmpl");
	// var pti = document.importNode(partytmpl.content, true);
	// var parties = document.getElementById("parties");
	// parties.appendChild(pti);
    };
    document.getElementById("newparty").onclick = newParty;
    document.getElementById("newperson").onclick = function(){instantiate("persontmpl", "people", "pers");};
})();
