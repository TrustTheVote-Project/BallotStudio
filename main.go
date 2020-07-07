package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/lib/pq"           // driver="postgres"
	_ "github.com/mattn/go-sqlite3" // driver="sqlite3"

	"bolson.org/~/src/login/login"
)

func maybefail(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	log.Printf(format, args...)
	os.Exit(1)
}

func maybeerr(w http.ResponseWriter, err error, code int, format string, args ...interface{}) bool {
	if err == nil {
		return false
	}
	msg := fmt.Sprintf(format, args...)
	if code >= 500 {
		log.Print(msg, "\t", err)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	w.Write([]byte(msg))
	return true
}

func texterr(w http.ResponseWriter, code int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if code >= 500 {
		log.Print(msg)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

// handler of /election and /election/*{,.pdf,.png,_bubbles.json,/scan}
type StudioHandler struct {
	udb login.UserDB
	edb electionAppDB

	drawBackend string

	cache Cache

	archiver ImageArchiver
}

var pdfPathRe *regexp.Regexp
var bubblesPathRe *regexp.Regexp
var pngPathRe *regexp.Regexp
var scanPathRe *regexp.Regexp
var docPathRe *regexp.Regexp

func init() {
	pdfPathRe = regexp.MustCompile(`^/election/(\d+)\.pdf$`)
	bubblesPathRe = regexp.MustCompile(`^/election/(\d+)_bubbles\.json$`)
	pngPathRe = regexp.MustCompile(`^/election/(\d+)\.png$`)
	scanPathRe = regexp.MustCompile(`^/election/(\d+)/scan$`)
	docPathRe = regexp.MustCompile(`^/election/(\d+)$`)
}

// implement http.Handler
func (sh *StudioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, _ := login.GetHttpUser(w, r, sh.udb)
	path := r.URL.Path
	if path == "/election" {
		if r.Method == "POST" {
			sh.handleElectionDocPOST(w, r, user, 0)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"nope"}`))
		return
	}
	// `^/election/(\d+)$`
	m := docPathRe.FindStringSubmatch(path)
	if m != nil {
		electionid, err := strconv.ParseInt(m[1], 10, 64)
		if maybeerr(w, err, 400, "bad item") {
			return
		}
		if r.Method == "GET" {
			sh.handleElectionDocGET(w, r, user, electionid)
		} else if r.Method == "POST" {
			sh.handleElectionDocPOST(w, r, user, electionid)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			w.Write([]byte(`{"error":"nope"}`))
		}
		return
	}
	// `^/election/(\d+)\.pdf$`
	m = pdfPathRe.FindStringSubmatch(path)
	if m != nil {
		bothob, err := sh.getPdf(m[1])
		if err != nil {
			he := err.(httpError)
			maybeerr(w, he.err, he.code, he.msg)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(200)
		w.Write(bothob.Pdf)
		return
	}
	// `^/election/(\d+)_bubbles\.json$`
	m = bubblesPathRe.FindStringSubmatch(path)
	if m != nil {
		bothob, err := sh.getPdf(m[1])
		if err != nil {
			he := err.(httpError)
			maybeerr(w, he.err, he.code, he.msg)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(bothob.BubblesJson)
		return
	}
	// `^/election/(\d+)\.png$`
	m = pngPathRe.FindStringSubmatch(path)
	if m != nil {
		pngbytes, err := sh.getPng(m[1])
		if err != nil {
			he := err.(httpError)
			maybeerr(w, he.err, he.code, he.msg)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(200)
		w.Write(pngbytes)
		return
	}
	// `^/election/(\d+)/scan$`
	m = scanPathRe.FindStringSubmatch(path)
	if m != nil {
		if r.Method == "POST" {
			sh.handleElectionScanPOST(w, r, user, m[1])
			return
		}
		// GET: serve a page with image upload
		// POST: receive image
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"TODO implement scan"}`))
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(400)
	w.Write([]byte(`nope`))
}

func (sh *StudioHandler) handleElectionDocPOST(w http.ResponseWriter, r *http.Request, user *login.User, itemid int64) {
	if user == nil {
		texterr(w, http.StatusUnauthorized, "nope")
		return
	}
	mbr := http.MaxBytesReader(w, r.Body, 1000000)
	body, err := ioutil.ReadAll(mbr)
	if err == io.EOF {
		err = nil
	}
	if maybeerr(w, err, 400, "bad body") {
		return
	}
	var ob map[string]interface{}
	err = json.Unmarshal(body, &ob)
	if maybeerr(w, err, 400, "bad json") {
		return
	}
	er := electionRecord{
		Id:    itemid,
		Owner: user.Guid,
		Data:  string(body),
	}
	newid, err := sh.edb.PutElection(er)
	if maybeerr(w, err, 500, "db put fail") {
		return
	}
	er.Id = newid
	ec := EditContext{}
	ec.set(newid)
	out, err := json.Marshal(ec)
	if maybeerr(w, err, 500, "json ret prep") {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(out)
}

func (sh *StudioHandler) handleElectionDocGET(w http.ResponseWriter, r *http.Request, user *login.User, itemid int64) {
	if user == nil {
		texterr(w, http.StatusUnauthorized, "nope")
		return
	}
	er, err := sh.edb.GetElection(itemid)
	if maybeerr(w, err, 400, "no item") {
		return
	}
	if user.Guid != er.Owner {
		texterr(w, http.StatusForbidden, "nope")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(er.Data))
}

func (sh *StudioHandler) getPdf(el string) (bothob *DrawBothOb, err error) {
	cr := sh.cache.Get(el)
	if cr != nil {
		bothob = cr.(*DrawBothOb)
	} else {
		electionid, err := strconv.ParseInt(el, 10, 64)
		if err != nil {
			return nil, &httpError{400, "bad item", err}
		}
		er, err := sh.edb.GetElection(electionid)
		if err != nil {
			return nil, &httpError{400, "no item", err}
		}
		bothob, err = draw(sh.drawBackend, er.Data)
		if err != nil {
			return nil, &httpError{500, "draw fail", err}
		}
		sh.cache.Put(el, bothob, len(bothob.Pdf)+len(bothob.BubblesJson))
	}
	return
}

func (sh *StudioHandler) getPng(el string) (pngbytes []byte, err error) {
	pngkey := el + ".png"
	cr := sh.cache.Get(pngkey)
	if cr != nil {
		pngbytes = cr.([]byte)
		return
	}
	var bothob *DrawBothOb
	bothob, err = sh.getPdf(el)
	if err != nil {
		return nil, err
	}
	pngbytes, err = pdftopng(bothob.Pdf)
	if err != nil {
		return nil, &httpError{500, "png fail", err}
	}
	sh.cache.Put(pngkey, pngbytes, len(pngbytes))
	return
}

type httpError struct {
	code int
	msg  string
	err  error
}

func (he httpError) Error() string {
	return fmt.Sprintf("%d %s, %v", he.code, he.msg, he.err)
}

type editHandler struct {
	udb login.UserDB
	t   *template.Template
}

type EditContext struct {
	ElectionId    int64  `json:"itemid,omitepmty"`
	PDFURL        string `json:"pdf,omitepmty"`
	BubbleJSONURL string `json:"bubbles,omitepmty"`
	ScanFormURL   string `json:"scan,omitepmty"`
	PostURL       string `json:"post,omitempty"`
	EditURL       string `json:"edit,omitempty"`
	GETURL        string `json:"url,omitempty"`
}

func (ec *EditContext) set(eid int64) {
	if eid == 0 {
		ec.PostURL = "/election"
	} else {
		ec.ElectionId = eid
		ec.PDFURL = fmt.Sprintf("/election/%d.pdf", eid)
		ec.BubbleJSONURL = fmt.Sprintf("/election/%d_bubbles.json", eid)
		ec.ScanFormURL = fmt.Sprintf("/election/%d/scan", eid)
		ec.PostURL = fmt.Sprintf("/election/%d", eid)
		ec.EditURL = fmt.Sprintf("/edit/%d", eid)
		ec.GETURL = fmt.Sprintf("/election/%d", eid)
	}
}

func (ec EditContext) Json() template.JS {
	x, err := json.Marshal(ec)
	if err != nil {
		return ""
	}
	//fmt.Printf("ec %s\n", string(x))
	return template.JS(string(x))
}

func (ec EditContext) JsonAttr() template.HTMLAttr {
	x, err := json.Marshal(ec)
	if err != nil {
		return template.HTMLAttr("")
	}
	//fmt.Printf("ec %s\n", string(x))
	return template.HTMLAttr(template.URLQueryEscaper(string(x)))
}

// http.HandlerFunc
// just fills out index.html template
func (edit *editHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	electionid := int64(0)
	if strings.HasPrefix(r.URL.Path, "/edit/") {
		xe, err := strconv.ParseInt(r.URL.Path[6:], 10, 64)
		if err == nil {
			electionid = xe
		}
	}
	log.Printf("GET %s", r.URL.Path)
	w.Header().Set("Content-Type", "text/html")
	ec := EditContext{}
	ec.set(electionid)
	err := edit.t.Execute(w, ec)
	if err != nil {
		log.Printf("editHandler template error, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	var listenAddr string
	flag.StringVar(&listenAddr, "http", ":8180", "interface:port to listen on, default \":8180\"")
	var oauthConfigPath string
	flag.StringVar(&oauthConfigPath, "oauth-json", "", "json file with oauth configs")
	var sqlitePath string
	flag.StringVar(&sqlitePath, "sqlite", "", "path to sqlite3 db to keep local data in")
	var postgresConnectString string
	flag.StringVar(&postgresConnectString, "postgres", "", "connection string to postgres database")
	var drawBackend string
	flag.StringVar(&drawBackend, "draw-backend", "", "url to drawing backend")
	flag.Parse()

	templates, err := template.ParseGlob("gotemplates/*.html")
	maybefail(err, "parse templates, %v", err)
	indextemplate := templates.Lookup("index.html")
	if indextemplate == nil {
		log.Print("no template index.html")
		os.Exit(1)
	}

	var udb login.UserDB
	var db *sql.DB
	var edb electionAppDB

	if len(sqlitePath) > 0 {
		if len(postgresConnectString) > 0 {
			fmt.Fprintf(os.Stderr, "error, only one of -sqlite or -postgres should be set")
			os.Exit(1)
			return
		}
		var err error
		db, err = sql.Open("sqlite3", sqlitePath)
		maybefail(err, "error opening sqlite3 db %#v, %v", sqlitePath, err)
		udb = login.NewSqlUserDB(db)
		edb = &sqliteedb{db}
	} else if len(postgresConnectString) > 0 {
		var err error
		db, err = sql.Open("postgres", postgresConnectString)
		maybefail(err, "error opening postgres db %#v, %v", postgresConnectString, err)
		udb = login.NewSqlUserDB(db)
		edb = &postgresedb{db}
	} else {
		log.Print("warning, running with in-memory database that will disappear when shut down")
		var err error
		db, err = sql.Open("sqlite3", ":memory:")
		maybefail(err, "error opening sqlite3 memory db, %v", err)
		udb = login.NewSqlUserDB(db)
		edb = &sqliteedb{db}
	}
	err = edb.Setup()
	maybefail(err, "edb setup, %v", err)
	err = udb.Setup()
	maybefail(err, "udb setup, %v", err)

	getdb := func() (login.UserDB, error) { return udb, nil }
	sh := StudioHandler{
		udb:         udb,
		edb:         edb,
		drawBackend: drawBackend,
	}
	edith := editHandler{udb, indextemplate}

	mux := http.NewServeMux()
	mux.Handle("/election", &sh)
	mux.Handle("/election/", &sh)
	mux.Handle("/edit", &edith)
	mux.Handle("/edit/", &edith)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	var authmods []*login.OauthCallbackHandler
	if len(oauthConfigPath) > 0 {
		fin, err := os.Open(oauthConfigPath)
		maybefail(err, "%s: could not open, %v", oauthConfigPath, err)
		oc, err := login.ParseConfigJSON(fin)
		maybefail(err, "%s: bad parse, %v", oauthConfigPath, err)
		authmods, err = login.BuildOauthMods(oc, mux, getdb, "/", "/")
		maybefail(err, "%s: oauth problems, %v", oauthConfigPath, err)
	}
	log.Printf("initialized %d oauth mods", len(authmods))
	mux.HandleFunc("/logout", login.LogoutHandler)
	mux.Handle("/", &sh)
	server := http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}
	log.Print("serving ", listenAddr)
	log.Fatal(server.ListenAndServe())
}
