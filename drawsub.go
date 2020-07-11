package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"strings"

	_ "github.com/spakin/netpbm"
)

// go get github.com/spakin/netpbm

type DrawBothOb struct {
	Pdf         []byte
	BubblesJson []byte
}

type DrawBothResponse struct {
	PdfB64  []byte                 `json:"pdfb64"`
	Bubbles map[string]interface{} `json:"bubbles"`
}

func draw(backendUrl string, electionjson string) (both *DrawBothOb, err error) {
	baseurl, err := url.Parse(backendUrl)
	if err != nil {
		return nil, fmt.Errorf("bad url, %v", err)
	}
	newpath := path.Join(baseurl.Path, "/draw")
	nurl := baseurl
	nurl.Path = newpath
	nurl.RawQuery = "both=1"
	drawurl := nurl.String()
	postbody := strings.NewReader(electionjson)
	resp, err := http.DefaultClient.Post(drawurl, "application/json", postbody)
	if err != nil {
		return nil, fmt.Errorf("draw POST, %v", err)
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		if len(body) > 50 {
			body = body[:50]
		}
		return nil, fmt.Errorf("draw POST %d %#v", resp.StatusCode, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	//dec := json.NewDecoder(resp.Body)
	var dbr DrawBothResponse
	//err = dec.Decode(&dbr)
	err = json.Unmarshal(body, &dbr)
	if err != nil {
		dbb := body
		if len(dbb) > 50 {
			dbb = dbb[:50]
		}
		return nil, fmt.Errorf("draw POST bad response, %v, %#v", err, string(dbb))
	}
	// json.Unmarshal helpfully converts "base64 content" unpacked into []byte
	/*
		pdf := make([]byte, base64.StdEncoding.DecodedLen(len(dbr.PdfB64)))
		actual, err := base64.StdEncoding.Decode(pdf, dbr.PdfB64)
		if err != nil {
			dp := dbr.PdfB64
			if len(dp) > 50 {
				dp = dp[:50]
			}
			return nil, fmt.Errorf("draw POST bad response b64, %v, %#v", err, string(dp))
		}
		if actual < len(pdf) {
			pdf = pdf[:actual]
		}
	*/
	bj, err := json.Marshal(dbr.Bubbles)
	if err != nil {
		return nil, fmt.Errorf("draw POST bad response bj, %v", err)
	}
	return &DrawBothOb{Pdf: dbr.PdfB64, BubblesJson: bj}, nil
}

type asyncReadResult struct {
	data []byte
	err error
}

func (arr asyncReadResult) String() string {
	xd := string(arr.data)
	if len(xd) > 50 {
		xd = xd[:50]
	}
	return fmt.Sprintf("err=%v msg=%v", arr.err, xd)
}

func asyncReadAll(fin io.ReadCloser, result chan asyncReadResult) {
	data, err := ioutil.ReadAll(fin)
	result <- asyncReadResult{data,err}
	close(result)
}

// uses subprocess `pdftoppm`
func pdftopng(pdf []byte) (pngbytes []byte, err error) {
	if len(pdf) == 0 {
		return nil, fmt.Errorf("pdftopng but empty pdf")
	}
	// TODO: handle multi page!
	cmd := exec.Command("pdftoppm", "-png", "-singlefile")
	if err != nil {
		return nil, fmt.Errorf("could not cmd pdftoppm, %v", err)
	}
	cmd.Stdin = bytes.NewReader(pdf)
	stdoutget, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not start pdftoppm stdout, %v", err)
	}
	stderrget, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("could not start pdftoppm stderr, %v", err)
	}
	//log.Print("starting pdftoppm")
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("could not start pdftoppm, %v", err)
	}
	result := make(chan asyncReadResult, 1)
	go asyncReadAll(stdoutget, result)
	stderr := make(chan asyncReadResult, 1)
	go asyncReadAll(stderrget, stderr)
	err = cmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("pdftoppm wait, %v", err)
	}
	stderrx := <-stderr
	ppm, ok := <-result
	if (!ok) || (ppm.err != nil) {
		return nil, fmt.Errorf("got no ppm, wrote %d bytes, %v, stderr(%s)", len(pdf), cmd.ProcessState, stderrx)
	}
	im, format, err := image.Decode(bytes.NewReader(ppm.data))
	if err != nil {
		return nil, fmt.Errorf("ppm decode as %s, %v", format, err)
	}
	pngw := bytes.Buffer{}
	err = png.Encode(&pngw, im)
	if err != nil {
		return nil, fmt.Errorf("png encode, %v", err)
	}
	pngbytes = pngw.Bytes()
	return
}
