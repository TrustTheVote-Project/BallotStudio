package draw

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"strings"
	//_ "github.com/spakin/netpbm"
)

type DrawBothOb struct {
	Pdf         []byte
	BubblesJson []byte
}

type DrawBothResponse struct {
	PdfB64  []byte                 `json:"pdfb64"`
	Bubbles map[string]interface{} `json:"bubbles"`
}

// TODO: launch dev draw server:
// FLASK_ENV=development FLASK_APP=draw/app.py "${HOME}/bsvenv/bin/flask" run -p 8081

func DrawElection(backendUrl string, electionjson string) (both *DrawBothOb, err error) {
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

type errorOrPngbytes struct {
	err      error
	pngpages [][]byte
}

func pngPageReader(reader io.Reader, out chan errorOrPngbytes) {
	var sizebytes [8]byte
	r := errorOrPngbytes{
		err:      nil,
		pngpages: make([][]byte, 0, 10),
	}
	for {
		_, err := io.ReadFull(reader, sizebytes[:])
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			out <- r
			close(out)
			return
		}
		if err != nil {
			r.err = err
			out <- r
			return
		}
		pnglen := binary.BigEndian.Uint64(sizebytes[:])
		b := binary.LittleEndian.Uint64(sizebytes[:])
		if b < pnglen {
			pnglen = b
		}
		nextpng := make([]byte, pnglen)
		_, err = io.ReadFull(reader, nextpng)
		if err != nil {
			r.err = err
			out <- r
			return
		}
		r.pngpages = append(r.pngpages, nextpng)
	}
}

// uses subprocess `pdftoppm`
func PdfToPng(pdf []byte) (pngbytes [][]byte, err error) {
	if len(pdf) == 0 {
		return nil, fmt.Errorf("pdftopng but empty pdf")
	}
	// requires poppler fork from https://github.com/brianolson/poppler
	cmd := exec.Command("pdftoppm", "-png", "-pngMultiBlock") // , "-singlefile"
	if err != nil {
		return nil, fmt.Errorf("could not cmd pdftoppm, %v", err)
	}
	reader, writer := io.Pipe()
	cmd.Stdin = bytes.NewReader(pdf)
	cmd.Stdout = writer
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	pchan := make(chan errorOrPngbytes, 1)
	go pngPageReader(reader, pchan)
	err = cmd.Run()
	if err != nil {
		se := string(stderr.Bytes())
		if len(se) > 50 {
			se = se[:50]
		}
		return nil, fmt.Errorf("pdftoppm err, %v, %v", err, se)
	}
	r := <-pchan
	return r.pngpages, r.err
}
