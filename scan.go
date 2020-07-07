package main

import (
	"bytes"
	"encoding/json"
	"image"
	"io/ioutil"
	"net/http"
	"strings"

	"bolson.org/~/src/ballotscan/scan"
	"bolson.org/~/src/login/login"
)

func (sh *StudioHandler) handleElectionScanPOST(w http.ResponseWriter, r *http.Request, user *login.User, itemname string) {
	// TODO: multipart file upload in addition to plain body POST
	contenttype := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contenttype, "image") {
		texterr(w, http.StatusBadRequest, "bad content type")
		return
	}
	brc := http.MaxBytesReader(w, r.Body, 10000000)
	imbytes, err := ioutil.ReadAll(brc)
	if maybeerr(w, err, 400, "bad image, %v", err) {
		return
	}
	im, format, err := image.Decode(bytes.NewReader(imbytes))
	if maybeerr(w, err, 400, "bad image, %v", err) {
		return
	}
	bothob, err := sh.getPdf(itemname)
	if err != nil {
		he := err.(httpError)
		maybeerr(w, he.err, he.code, he.msg)
		return
	}
	var bubbles scan.BubblesJson
	err = json.Unmarshal(bothob.BubblesJson, &bubbles)
	if maybeerr(w, err, 500, "bubble json decode, %v", err) {
		return
	}
	pngbytes, err := sh.getPng(itemname)
	if err != nil {
		he := err.(httpError)
		maybeerr(w, he.err, he.code, he.msg)
		return
	}
	orig, format, err := image.Decode(bytes.NewReader(pngbytes))
	if maybeerr(w, err, 500, "orig png decode (%s), %v", format, err) {
		return
	}

	var s scan.Scanner
	s.Bj = bubbles
	s.SetOrigImage(orig)
	marked, err := s.ProcessScannedImage(im)
	if maybeerr(w, err, 500, "process err: %v", err) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	mjson, _ := json.Marshal(marked)
	w.Write(mjson)
}
