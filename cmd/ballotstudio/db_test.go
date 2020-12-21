package main

import (
	"database/sql"
	"flag"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // driver="sqlite3"
)

func mtfail(t *testing.T, err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	t.Fatalf(format, args...)
}

var pgConnectString string
var pgdb *sql.DB

func TestMain(m *testing.M) {
	flag.StringVar(&pgConnectString, "postgres", "", "connection string for postgres")
	flag.Parse()

	if pgConnectString != "" {
		var err error
		pgdb, err = sql.Open("postgres", pgConnectString)
		maybefail(err, "error opening postgres db, %v", err)
		defer pgdb.Close()
	}

	os.Exit(m.Run())
}

func TestSqliteDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	mtfail(t, err, "open sqlite mem, %v", err)
	defer db.Close()
	edb := NewSqliteEDB(db)
	err = edb.Setup()
	mtfail(t, err, "edb sqlite setup, %v", err)
	testEdb(t, edb)
}

func TestPostgresDB(t *testing.T) {
	if pgdb == nil {
		t.Skip("no -postgres connect string")
		return
	}
	edb := NewPostgresEDB(pgdb)
	err := edb.Setup()
	mtfail(t, err, "edb postgres setup, %v", err)
	testEdb(t, edb)
}

func testEdb(t *testing.T, edb electionAppDB) {
	// election data stuff
	er := electionRecord{
		Id:    0,
		Owner: 1,
		Data:  "helloo",
		Meta:  "wat",
	}
	newid, err := edb.PutElection(er)
	mtfail(t, err, "new er put, %v", err)
	er.Id = newid
	xe, err := edb.GetElection(newid)
	mtfail(t, err, "new er get, %v", err)
	if *xe != er {
		t.Errorf("put-get neq a=%#v b=%v", er, *xe)
	}
	xe.Data = "howdy"
	newid, err = edb.PutElection(*xe)
	mtfail(t, err, "er update, %v", err)
	if newid != xe.Id {
		t.Errorf("id change on update %d -> %d", xe.Id, newid)
	}
	e2, err := edb.GetElection(xe.Id)
	mtfail(t, err, "er get 2, %v", err)
	if *e2 != *xe {
		t.Errorf("update-get neq a=%#v b=%v", *xe, *e2)
	}

	eids, err := edb.ElectionsForUser(er.Owner)
	mtfail(t, err, "er ElectionsForUser, %v", err)
	if len(eids) != 1 {
		t.Errorf("expected 1 election list but got %d", len(eids))
	} else if eids[0] != newid {
		t.Errorf("listed election for owner %d wrong id, wanted %d, got %d", er.Owner, newid, eids[0])
	}

	// invite token stuff
	const token = "tok"
	now := time.Now()
	later := now.Add(20 * time.Minute)
	err = edb.MakeInviteToken(token, later)
	mtfail(t, err, "MakeInviteToken %v", err)
	ok, expires, err := edb.PeekInviteToken(token)
	mtfail(t, err, "PeekInviteToken %v", err)
	if !ok || expires.Before(now) {
		t.Errorf("bad peek ok=%v exp=%s", ok, expires)
	}
	ok, err = edb.UseInviteToken(token)
	mtfail(t, err, "UseInviteToken %v", err)
	if !ok {
		t.Errorf("could not use token")
	}
	// now try again!
	ok, err = edb.UseInviteToken(token)
	mtfail(t, err, "UseInviteToken 2 %v", err)
	if ok {
		t.Errorf("invite token double spend")
	}

	const t2 = "t2"
	ago := now.Add(-1 * time.Minute)
	err = edb.MakeInviteToken(token, ago)
	mtfail(t, err, "MakeInviteToken 2 %v", err)
	err = edb.GCInviteTokens()
	mtfail(t, err, "GCInviteTokens %v", err)

	ok, _, _ = edb.PeekInviteToken(t2)
	o2, _ := edb.UseInviteToken(t2)
	if ok || o2 {
		t.Errorf("t2 should be gone")
	}
}
