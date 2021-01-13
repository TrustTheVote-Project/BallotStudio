package main

import (
	"bytes"
	"encoding/json"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/brianolson/ballotscan/scan"
	"github.com/brianolson/login/login"
)

func (sh *StudioHandler) handleElectionScanPOST(w http.ResponseWriter, r *http.Request, user *login.User, itemname string) {
	imbytes := getImage(w, r)
	if imbytes == nil {
		return
	}
	im, format, err := image.Decode(bytes.NewReader(imbytes))
	if maybeerr(w, err, 400, "bad image, %v", err) {
		return
	}
	bothob, err := sh.getPdf(r.Context(), itemname, false)
	if err != nil {
		he := err.(*httpError)
		maybeerr(w, he.err, he.code, he.msg)
		return
	}
	var bubbles scan.BubblesJson
	err = json.Unmarshal(bothob.BubblesJson, &bubbles)
	if maybeerr(w, err, 500, "bubble json decode, %v", err) {
		return
	}
	pngbytes, err := sh.getPng(r.Context(), itemname, false)
	if err != nil {
		he := err.(*httpError)
		maybeerr(w, he.err, he.code, he.msg)
		return
	}
	// TODO: detect which page was scanned (by barcode, header, match quality?)
	orig, format, err := image.Decode(bytes.NewReader(pngbytes[0]))
	if maybeerr(w, err, 500, "orig png decode (%s), %v", format, err) {
		return
	}

	if sh.archiver != nil {
		go sh.archiver.ArchiveImage(imbytes, r)
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

// get image whether it is main POST body or in a multipart section
func getImage(w http.ResponseWriter, r *http.Request) (imbytes []byte) {
	var err error
	contenttype := r.Header.Get("Content-Type")
	if isImage(contenttype) {
		brc := http.MaxBytesReader(w, r.Body, 10000000)
		imbytes, err = ioutil.ReadAll(brc)
		if maybeerr(w, err, 400, "bad image, %v", err) {
			return
		}
		return
	}

	mpreader, err := r.MultipartReader()
	if maybeerr(w, err, 400, "bad multipart, %v", err) {
		return
	}
	for true {
		part, err := mpreader.NextPart()
		if err == io.EOF {
			break
		}
		if maybeerr(w, err, 400, "bad multipart part, %v", err) {
			return
		}

		//log.Printf("got part cd=%v fn=%v form=%v", part.Header.Get("Content-Disposition"), part.FileName(), part.FormName())
		if isImage(part.Header.Get("Content-Type")) {
			imbytes, err = ioutil.ReadAll(part)
			if maybeerr(w, err, 400, "bad multipart image, %v", err) {
				return
			}
			return
		}
	}
	// got nothing; return nil,nil
	return
}

func isImage(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}
