package data

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var DebugOut io.Writer = nil //os.Stderr

func debug(format string, args ...interface{}) {
	if DebugOut == nil {
		return
	}
	fmt.Fprintf(DebugOut, format, args...)
}

// Fixup will detect duplicate @id values and give new ids to some entries
func Fixup(er map[string]interface{}) map[string]interface{} {
	var ifc idFixupContext
	return ifc.fixup(er)
}

type msi struct {
	they map[string]interface{}
}

func (x *msi) set(k string, v interface{}) {
	if x.they == nil {
		x.they = make(map[string]interface{})
	}
	x.they[k] = v
}

func pathstr(path []string, key string) string {
	dp := strings.Join(path, ".")
	if key == "" {
		return dp
	}
	return dp + "." + key
}

//type nodefunc func(map[string]interface{}, interface{}) map[string]interface{}

type idFixupContext struct {
	// seen[@type][@id] = path
	seen map[string]map[string][]string

	// old id : new id
	remap map[string]string

	messages []string

	needsNewId []recNeedsNewId

	// {prefix:next id}
	nextid map[string]int

	// unk{i}
	unki int
}

type recNeedsNewId struct {
	path []string
	er   map[string]interface{}
}

func recordNeedsNewId(er map[string]interface{}, path []string) (out recNeedsNewId) {
	out.path = make([]string, len(path))
	copy(out.path, path)
	out.er = er
	return
}

func (ifc *idFixupContext) fixup(er map[string]interface{}) map[string]interface{} {
	path := make([]string, 0, 20)
	er = ifc.idFixupInner(er, path)
	for _, rec := range ifc.needsNewId {
		newid := ifc.newId(rec)
		ifc.msg("needs new id (%s %s) -> %s\n", rec.er["@type"], rec.er["@id"], newid)
		rec.er["@id"] = newid
	}
	for _, msg := range ifc.messages {
		debug("%s", msg)
	}
	return er
}

// func (ifc *idFixupContext) traverse(er map[string]interface{}, path []string, onNode nodefunc, nfc interface{}) map[string]interface{} {
// 	plen := len(path)
// 	var updates msi
// 	return er
// }

// idFixupInner recurses down map[string]interface{} and []interface{}.
// idFixupInner notes for later duplicate @id values and which elements need new ids.
func (ifc *idFixupContext) idFixupInner(er map[string]interface{}, path []string) map[string]interface{} {
	debug("%s fi\n", pathstr(path, ""))
	plen := len(path)
	var updates msi
	attype, hasat := er["@type"]
	atid, hasid := er["@id"]
	if hasat && hasid {
		debug("%s (%s %s)\n", pathstr(path, ""), attype, atid)
		dup := ifc.checkSetSeen(attype.(string), atid.(string), path)
		if dup {
			ifc.needsNewId = append(ifc.needsNewId, recordNeedsNewId(er, path))
		}
	} else if hasid && !hasat {
		ifc.msg("@id=%v but not @type at %#v", atid, path)
	} else if hasat && !hasid {
		_, shouldHaveId := tsmap[attype.(string)]
		if shouldHaveId {
			ifc.msg("@type=%s but no @id at %#v", attype, path)
		}
	}
	for k, iv := range er {
		switch v := iv.(type) {
		case map[string]interface{}:
			debug("%s {}\n", pathstr(path, k))
			path = append(path, k)
			nv := ifc.idFixupInner(v, path)
			path = path[:plen]
			updates.set(k, nv)
		case []interface{}:
			debug("%s []\n", pathstr(path, k))
			ap := append(path, k)
			aplen := len(ap)
			for i, av := range v {
				amv, ok := av.(map[string]interface{})
				if ok {
					ap = append(ap, strconv.Itoa(i))
					//debug("%s %v\n", pathstr(ap, ""), ap)
					nv := ifc.idFixupInner(amv, ap)
					ap = ap[:aplen]
					v[i] = nv
				} else {
					debug("%s [%d] %T %#v\n", pathstr(ap, k), i, av, av)
				}
			}
			path = path[:plen]
		default:
			debug("%s %T %#v\n", pathstr(path, k), iv, iv)
		}
	}
	for nk, nv := range updates.they {
		er[nk] = nv
	}
	return er
}

func (ifc *idFixupContext) msg(format string, args ...interface{}) {
	//debug(format, args...)
	m := fmt.Sprintf(format, args...)
	ifc.messages = append(ifc.messages, m)
}

func (ifc *idFixupContext) checkSetSeen(attype, atid string, path []string) (dup bool) {
	debug("seen (%s %s) at %s\n", attype, atid, pathstr(path, ""))
	seqPrefix, ok := tsmap[attype]
	if ok {
		if !strings.HasPrefix(atid, seqPrefix) {
			ifc.msg("bad prefix (%s %s) wanted %s*, path=%v", attype, atid, seqPrefix, path)
			dup = true
			return
		}
	}
	if ifc.seen == nil {
		ifc.seen = make(map[string]map[string][]string)
	}
	ts := ifc.seen[attype]
	if ts == nil {
		ts = make(map[string][]string)
		ifc.seen[attype] = ts
	}
	orig := ts[atid]
	if orig != nil {
		ifc.msg("dup id %s, orig path=%v dup path=%v", atid, orig, path)
		dup = true
	} else {
		cpath := make([]string, len(path))
		copy(cpath, path)
		ts[atid] = cpath
	}
	return
}

func (ifc *idFixupContext) newId(rec recNeedsNewId) (newatid string) {
	attype := rec.er["@type"].(string)
	prefix, ok := tsmap[attype]
	if !ok {
		ifc.unki++
		prefix = fmt.Sprintf("unk%d_", ifc.unki)
		debug("no seq prefix for %s, using %s", attype, prefix)
		tsmap[attype] = prefix
	}
	if ifc.nextid == nil {
		ifc.nextid = make(map[string]int)
	}
	dup := true
	for dup {
		ni, ok := ifc.nextid[prefix]
		if !ok {
			ni = 1
		}
		newatid = fmt.Sprintf("%s%d", prefix, ni)
		ifc.nextid[prefix] = ni + 1
		dup = ifc.checkSetSeen(attype, newatid, rec.path)
	}
	return newatid
}

//go:generate go run ../misc/texttosource/main.go data type_seq.json

var tsmap map[string]string

func init() {
	err := json.Unmarshal([]byte(type_seq_json), &tsmap)
	if err != nil {
		panic(err)
	}
}
