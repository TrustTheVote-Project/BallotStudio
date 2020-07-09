package main

import (
	"crypto/rand"
	"encoding/base32"
	"html/template"
	"log"
	mrand "math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/brianolson/login/login"
)

// q=1..5, for q*5 bytes of randomness, as base32 becoming q*8 chars of text
func randomInviteToken(q int) string {
	if q < 1 {
		q = 1
	}
	if q > 5 {
		q = 5
	}
	buf := make([]byte, q*5)
	_, err := rand.Read(buf)
	if err != nil {
		// wat? this can error?
		// get some lesser rand and carry on
		mrand.Read(buf)
	}
	return base32.StdEncoding.EncodeToString(buf)
}

type inviteHandler struct {
	edb electionAppDB
	udb login.UserDB

	authmods []*login.OauthCallbackHandler

	signupPage *template.Template
}

// GET /signup/
func (ih *inviteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/signup/") {
		log.Printf("not signup path=%#v", path)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if r.Method == "POST" {
		ih.handlePOST(w, r)
	}
	token := path[8:]
	ok, expires, err := ih.edb.PeekInviteToken(token)
	if !ok {
		log.Printf("token %#v %v %v %v", token, ok, expires, err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	now := time.Now()
	until := expires.Sub(now)
	icookie := http.Cookie{
		Name:   "i",
		Value:  token,
		Path:   "/",
		MaxAge: int(until.Seconds()),
	}
	http.SetCookie(w, &icookie)
	ih.renderSignup(w, r, ih.scm(""))
}

func (ih *inviteHandler) handlePOST(w http.ResponseWriter, r *http.Request) {
	cx, err := r.Cookie("i")
	if err != nil || cx == nil {
		log.Print("no invite cookie")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	ok, expires, err := ih.edb.PeekInviteToken(cx.Value)
	if !ok {
		log.Printf("invite token %v %v %v", ok, expires, err)
		ih.renderSignup(w, r, ih.scm("invalid invite token"))
		return
	}
	r.ParseForm()
	username := r.PostForm.Get("username")
	if username == "" {
		ih.renderSignup(w, r, ih.scm("username cannot be blank"))
		return
	}
	password := r.PostForm.Get("password")
	if password == "" {
		ih.renderSignup(w, r, ih.scm("password cannot be blank"))
		return
	}
	newuser := login.User{}
	newuser.Username = username
	newuser.SetPassword(password)
	_, err = ih.udb.PutNewUser(&newuser)
	if err != nil {
		texterr(w, 500, "error creating user, %v", err)
		return
	}
	ih.edb.UseInviteToken(cx.Value)
	// clear invite token
	icookie := http.Cookie{
		Name:   "i",
		MaxAge: -1,
	}
	http.SetCookie(w, &icookie)
	// this should set a login cookie using the same form values
	login.GetHttpUser(w, r, ih.udb)
	http.Redirect(w, r, "/", http.StatusFound)
}

type SignupContext struct {
	Message  string
	AuthMods []*login.OauthCallbackHandler
}

func (ih *inviteHandler) scm(message string) SignupContext {
	return SignupContext{message, ih.authmods}
}

func (ih *inviteHandler) renderSignup(w http.ResponseWriter, r *http.Request, ctx SignupContext) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	ih.signupPage.Execute(w, ctx)
}

/*
// wrap a handler, only get through if there's an invite cookie
func (ih *inviteHandler) requireInviteCookie(handler http.Handler) http.Handler {
	return &requireInviteWrapper{ih.edb, handler}
}

type requireInviteWrapper struct {
	edb electionAppDB
	sub http.Handler
}

func (riw *requireInviteWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: set a message cookie that will show on the home page
	cx, err := r.Cookie("i")
	if err != nil || cx == nil {
		log.Print("no invite cookie")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	ok, expires, err := riw.edb.PeekInviteToken(cx.Value)
	if !ok {
		log.Printf("invite token %v %v %v", ok, expires, err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	riw.sub.ServeHTTP(w, r)
}
*/

type makeInviteTokenHandler struct {
	edb       electionAppDB
	udb       login.UserDB
	tokenpage *template.Template
}

type tokenPageContext struct {
	Token string
}

func (ih *makeInviteTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, _ := login.GetHttpUser(w, r, ih.udb)
	if user == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	inviteToken := randomInviteToken(2)
	err := ih.edb.MakeInviteToken(inviteToken, time.Now().Add(7*24*time.Hour))
	if maybeerr(w, err, 500, "db err creating token, %v", err) {
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	ih.tokenpage.Execute(w, tokenPageContext{inviteToken})
}
