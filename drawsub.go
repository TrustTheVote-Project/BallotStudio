package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
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

func asyncReadAll(fin io.ReadCloser, result chan []byte) {
	data, err := ioutil.ReadAll(fin)
	if err != nil {
		log.Printf("pdftoppm read err, %v", err)
		close(result)
		return
	}
	//log.Printf("got %d bytes from pdftoppm", len(data))
	result <- data
	close(result)
}

// uses subprocess `pdftoppm`
func pdftopng(pdf []byte) (pngbytes []byte, err error) {
	cmd := exec.Command("pdftoppm")
	if err != nil {
		return nil, fmt.Errorf("could not cmd pdftoppm, %v", err)
	}
	stdinput, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("could not start pdftoppm stdin, %v", err)
	}
	stdoutget, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not start pdftoppm stdout, %v", err)
	}
	//log.Print("starting pdftoppm")
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("could not start pdftoppm, %v", err)
	}
	result := make(chan []byte, 1)
	go asyncReadAll(stdoutget, result)
	_, err = stdinput.Write(pdf)
	if err != nil {
		return nil, fmt.Errorf("could not write pdftoppm, %v", err)
	}
	//log.Printf("sent %d/%d bytes of pdf to pdftoppm", n, len(pdf))
	stdinput.Close()
	err = cmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("pdftoppm wait, %v", err)
	}
	ppm := <-result
	im, format, err := image.Decode(bytes.NewReader(ppm))
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
