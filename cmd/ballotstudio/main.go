package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"           // driver="postgres"
	_ "github.com/mattn/go-sqlite3" // driver="sqlite3"

	"github.com/brianolson/ballotstudio/data"
	"github.com/brianolson/ballotstudio/draw"
	"github.com/brianolson/login/login"
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
	if code >= 500 || true {
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
	edb electionAppDB
	udb login.UserDB

	drawBackend string

	cache Cache

	//scantemplate *template.Template
	//home         *template.Template
	templates *TemplateSet
	archiver  ImageArchiver

	authmods []*login.OauthCallbackHandler
}

var pdfPathRe *regexp.Regexp
var bubblesPathRe *regexp.Regexp
var pngPathRe *regexp.Regexp
var pngPagePathRe *regexp.Regexp
var scanPathRe *regexp.Regexp
var docPathRe *regexp.Regexp

func init() {
	pdfPathRe = regexp.MustCompile(`^/election/(\d+)\.pdf$`)
	bubblesPathRe = regexp.MustCompile(`^/election/(\d+)_bubbles\.json$`)
	pngPathRe = regexp.MustCompile(`^/election/(\d+)\.png$`)
	pngPagePathRe = regexp.MustCompile(`^/election/(\d+)\.(\d+)\.png$`)
	scanPathRe = regexp.MustCompile(`^/election/(\d+)/scan$`)
	docPathRe = regexp.MustCompile(`^/election/(\d+)$`)
}

// implement http.Handler
func (sh *StudioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, _ := login.GetHttpUser(w, r, sh.udb)
	path := r.URL.Path
	if path == "/election" {
		if r.Method == "POST" {
			sh.handleElectionDocPOST(w, r, user, "", 0)
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
			sh.handleElectionDocPOST(w, r, user, m[1], electionid)
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
			he := err.(*httpError)
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
			he := err.(*httpError)
			maybeerr(w, he.err, he.code, he.msg)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(bothob.BubblesJson)
		return
	}
	// `^/election/(\d+)\.(\d+)\.png$`
	m = pngPagePathRe.FindStringSubmatch(path)
	if m != nil {
		pagenum, err := strconv.Atoi(string(m[2]))
		if maybeerr(w, err, 400, "bad page") {
			return
		}
		pngbytes, err := sh.getPng(m[1])
		if err != nil {
			he := err.(*httpError)
			maybeerr(w, he.err, he.code, he.msg)
			return
		}
		if pagenum > len(pngbytes) {
			texterr(w, 400, "bad page")
		}
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(200)
		w.Write(pngbytes[pagenum])
		return
	}
	// `^/election/(\d+)\.png$`
	m = pngPathRe.FindStringSubmatch(path)
	if m != nil {
		pngbytes, err := sh.getPng(m[1])
		if err != nil {
			he := err.(*httpError)
			maybeerr(w, he.err, he.code, he.msg)
			return
		}
		if len(pngbytes) > 1 {
			texterr(w, 400, "document has more than one page")
		}
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(200)
		w.Write(pngbytes[0])
		return
	}
	// `^/election/(\d+)/scan$`
	m = scanPathRe.FindStringSubmatch(path)
	if m != nil {
		// POST: receive image
		// GET: serve a page with image upload
		if r.Method == "POST" {
			sh.handleElectionScanPOST(w, r, user, m[1])
			return
		}
		electionid, err := strconv.ParseInt(m[1], 10, 64)
		if maybeerr(w, err, 400, "bad item") {
			return
		}
		w.Header().Set("Content-Type", "text/html")
		ec := EditContext{}
		ec.set(electionid)
		scantemplate, err := sh.templates.Lookup("scanform.html")
		if maybeerr(w, err, 500, "scanform.html: %v", err) {
			return
		}
		scantemplate.Execute(w, ec)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	home, err := sh.templates.Lookup("home.html")
	if maybeerr(w, err, 500, "home.html: %v", err) {
		return
	}
	home.Execute(w, HomeContext{user, sh.authmods})
}

type HomeContext struct {
	User     *login.User
	AuthMods []*login.OauthCallbackHandler
}

func (sh *StudioHandler) handleElectionDocPOST(w http.ResponseWriter, r *http.Request, user *login.User, itemname string, itemid int64) {
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
	ob = data.Fixup(ob)
	nbody, err := json.Marshal(ob)
	if maybeerr(w, err, 400, "re-json body") {
		return
	}
	body = nbody
	if itemid != 0 {
		older, _ := sh.edb.GetElection(itemid)
		if older != nil {
			if older.Owner != user.Guid {
				texterr(w, http.StatusUnauthorized, "nope")
				return
			}
		}
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
	sh.cache.Invalidate(itemname)
	sh.cache.Invalidate(itemname + ".png")
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
	// Allow everything to be readable? TODO: flexible ACL?
	// if user == nil {
	// 	texterr(w, http.StatusUnauthorized, "nope")
	// 	return
	// }
	er, err := sh.edb.GetElection(itemid)
	if maybeerr(w, err, 400, "no item") {
		return
	}
	// Allow everything to be readable? TODO: flexible ACL?
	// if user.Guid != er.Owner {
	// 	texterr(w, http.StatusForbidden, "nope")
	// 	return
	// }
	// TODO? remove fixup on GET after all old records have been fixed on POST?
	var ob map[string]interface{}
	err = json.Unmarshal([]byte(er.Data), &ob)
	if maybeerr(w, err, 400, "bad json") {
		return
	}
	ob = data.Fixup(ob)
	nbody, err := json.Marshal(ob)
	if maybeerr(w, err, 400, "re-json body") {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(nbody)
}

func (sh *StudioHandler) getPdf(el string) (bothob *draw.DrawBothOb, err error) {
	cr := sh.cache.Get(el)
	if cr != nil {
		bothob = cr.(*draw.DrawBothOb)
	} else {
		electionid, err := strconv.ParseInt(el, 10, 64)
		if err != nil {
			return nil, &httpError{400, "bad item", err}
		}
		er, err := sh.edb.GetElection(electionid)
		if err != nil {
			return nil, &httpError{400, "no item", err}
		}
		bothob, err = draw.DrawElection(sh.drawBackend, er.Data)
		if err != nil {
			return nil, &httpError{500, "draw fail", err}
		}
		sh.cache.Put(el, bothob, len(bothob.Pdf)+len(bothob.BubblesJson))
	}
	return
}

func (sh *StudioHandler) getPng(el string) (pngbytes [][]byte, err error) {
	pngkey := el + ".png"
	cr := sh.cache.Get(pngkey)
	if cr != nil {
		pngbytes = cr.([][]byte)
		return
	}
	var bothob *draw.DrawBothOb
	bothob, err = sh.getPdf(el)
	if err != nil {
		return nil, err
	}
	pngbytes, err = draw.PdfToPng(bothob.Pdf)
	if err != nil {
		return nil, &httpError{500, "png fail", err}
	}
	tlen := 0
	for _, page := range pngbytes {
		tlen += len(page)
	}
	sh.cache.Put(pngkey, pngbytes, tlen)
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
	edb electionAppDB
	udb login.UserDB
	ts  *TemplateSet
	//t   *template.Template
}

type EditContext struct {
	ElectionId    int64  `json:"itemid,omitepmty"`
	PDFURL        string `json:"pdf,omitepmty"`
	BubbleJSONURL string `json:"bubbles,omitepmty"`
	ScanFormURL   string `json:"scan,omitepmty"`
	PostURL       string `json:"post,omitempty"`
	EditURL       string `json:"edit,omitempty"`
	GETURL        string `json:"url,omitempty"`
	StaticRoot    string `json:"staticroot,omitempty"`
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
	ec.StaticRoot = "/static"
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

// http.Handler
// just fills out index.html template
func (edit *editHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	electionid := int64(0)
	if strings.HasPrefix(r.URL.Path, "/edit/") {
		xe, err := strconv.ParseInt(r.URL.Path[6:], 10, 64)
		if err == nil {
			electionid = xe
		}
	}
	w.Header().Set("Content-Type", "text/html")
	ec := EditContext{}
	ec.set(electionid)
	t, err := edit.ts.Lookup("index.html")
	if maybeerr(w, err, 500, "index.html: %v", err) {
		return
	}
	err = t.Execute(w, ec)
	if err != nil {
		log.Printf("editHandler template error, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func addrGetPort(listenAddr string) int {
	x := strings.LastIndex(listenAddr, ":")
	if x < 0 {
		return 80
	}
	v, err := strconv.ParseInt(listenAddr[x+1:], 10, 32)
	if err != nil {
		return 80
	}
	return int(v)
}

func sigtermHandler(c <-chan os.Signal, server *http.Server, cf func()) {
	_, ok := <-c
	if ok {
		go cf()
		server.Shutdown(context.Background())
	}
}

func exists(path string) (out string, ok bool) {
	_, err := os.Stat(path)
	if err == nil {
		return path, true
	}
	return "", false
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
	var imageArchiveDir string
	flag.StringVar(&imageArchiveDir, "im-archive-dir", "", "directory to archive uploaded scanned images to; will mkdir -p")
	var cookieKeyb64 string
	flag.StringVar(&cookieKeyb64, "cookie-key", "", "base64 of 16 bytes for encrypting cookies")
	var pidpath string
	flag.StringVar(&pidpath, "pid", "", "path to write process id to")
	var debug bool
	flag.BoolVar(&debug, "debug", false, "more logging")
	var flaskPath string
	flag.StringVar(&flaskPath, "flask", "", "path to flask for running draw/app.py")
	flag.Parse()

	if debug {
		data.DebugOut = os.Stderr
	}

	//templates, err := template.ParseGlob("gotemplates/*.html")
	templates, err := HtmlTemplateGlob("gotemplates/*.html")
	templates.Reloading = true // TODO: disable for prod
	maybefail(err, "parse templates, %v", err)
	_, err = templates.Lookup("index.html")
	maybefail(err, "no index.html, %v", err)

	if cookieKeyb64 == "" {
		ck := login.GenerateCookieKey()
		log.Printf("-cookie-key %s", base64.StdEncoding.EncodeToString(ck))
	} else {
		ck, err := base64.StdEncoding.DecodeString(cookieKeyb64)
		maybefail(err, "-cookie-key, %v", err)
		err = login.SetCookieKey(ck)
		maybefail(err, "-cookie-key, %v", err)
	}

	var db *sql.DB
	var udb login.UserDB
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
		edb = NewSqliteEDB(db)
	} else if len(postgresConnectString) > 0 {
		var err error
		db, err = sql.Open("postgres", postgresConnectString)
		maybefail(err, "error opening postgres db %#v, %v", postgresConnectString, err)
		udb = login.NewSqlUserDB(db)
		edb = NewPostgresEDB(db)
	} else {
		log.Print("warning, running with in-memory database that will disappear when shut down")
		var err error
		db, err = sql.Open("sqlite3", ":memory:")
		maybefail(err, "error opening sqlite3 memory db, %v", err)
		udb = login.NewSqlUserDB(db)
		edb = NewSqliteEDB(db)
	}
	defer db.Close()
	err = edb.Setup()
	maybefail(err, "edb setup, %v", err)
	err = udb.Setup()
	maybefail(err, "udb setup, %v", err)
	inviteToken := randomInviteToken(2)
	err = edb.MakeInviteToken(inviteToken, time.Now().Add(30*time.Minute))
	maybefail(err, "storing invite token %s, %v", inviteToken, err)
	ok, expires, err := edb.PeekInviteToken(inviteToken)
	log.Printf("token=%s ok=%v expires=%s, err=%v", inviteToken, ok, expires, err)
	log.Printf("http://localhost:%d/signup/%s", addrGetPort(listenAddr), inviteToken)
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	go gcThread(ctx, edb, 57*time.Minute)

	if len(drawBackend) == 0 {
		var drawserver draw.DrawServer
		if flaskPath == "" {
			for _, fp := range []string{"./flask", "bsvenv/bin/flask"} {
				var ok bool
				flaskPath, ok = exists(fp)
				if ok {
					break
				}
			}
		}
		drawserver.FlaskPath = flaskPath
		err = drawserver.Start()
		maybefail(err, "could not start draw server, %v", err)
		drawBackend = drawserver.BackendUrl()
		defer drawserver.Stop()
	}

	var archiver ImageArchiver
	if imageArchiveDir != "" {
		archiver, err = NewFileImageArchiver(imageArchiveDir)
		maybefail(err, "image archive dir, %v", err)
	}
	sh := StudioHandler{
		edb:         edb,
		udb:         udb,
		drawBackend: drawBackend,
		templates:   &templates,
		//scantemplate: templates.Lookup("scanform.html"),
		//home:         templates.Lookup("home.html"),
		archiver: archiver,
	}
	edith := editHandler{edb, udb, &templates}
	ih := inviteHandler{
		edb: edb,
		udb: udb,
		//signupPage: templates.Lookup("signup.html"),
		templates: &templates,
	}

	mith := makeInviteTokenHandler{
		edb:       edb,
		udb:       udb,
		templates: &templates, //.Lookup("invitetoken.html"),
	}

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
		authmods, err = login.BuildOauthMods(oc, udb, "/", "/")
		maybefail(err, "%s: oauth problems, %v", oauthConfigPath, err)
		for _, am := range authmods {
			mux.Handle(am.HandlerUrl(), am)
		}
	}
	ih.authmods = authmods
	sh.authmods = authmods
	mux.Handle("/signup/", &ih)
	log.Printf("initialized %d oauth mods", len(authmods))
	mux.HandleFunc("/logout", login.LogoutHandler)
	mux.Handle("/makeinvite", &mith)
	mux.Handle("/", &sh)
	server := http.Server{
		Addr:        listenAddr,
		Handler:     mux,
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}
	if pidpath != "" {
		pidf, err := os.Create(pidpath)
		if err != nil {
			log.Printf("could not create pidfile, %v", err)
			// meh, keep going
		} else {
			fmt.Fprintf(pidf, "%d", os.Getpid())
			pidf.Close()
		}
	}
	sigterm := make(chan os.Signal, 1)
	go sigtermHandler(sigterm, &server, cf)
	signal.Notify(sigterm, syscall.SIGTERM)
	log.Print("serving ", listenAddr)
	log.Fatal(server.ListenAndServe())
}
