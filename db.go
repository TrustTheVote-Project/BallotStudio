package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

type electionRecord struct {
	Id    int64
	Owner int64
	Data  string // json
	Meta  string // json
}

// edb for short
type electionAppDB interface {
	Setup() error
	GetElection(id int64) (*electionRecord, error)
	PutElection(electionRecord) (newid int64, err error)
	MakeInviteToken(token string, expires time.Time) error
	PeekInviteToken(token string) (ok bool, expires time.Time, err error)
	UseInviteToken(token string) (ok bool, err error)
	GCInviteTokens() (err error)
}

func NewSqliteEDB(db *sql.DB) electionAppDB {
	return &sqliteedb{db}
}

type sqliteedb struct {
	db *sql.DB
}

// implement electionAppDB
func (sdb *sqliteedb) Setup() error {
	cmds := []string{
		// use builtin ROWID
		"CREATE TABLE IF NOT EXISTS elections (data TEXT, owner bigint, meta TEXT)",

		"CREATE TABLE IF NOT EXISTS metastate (k TEXT PRIMARY KEY, v BLOB)",
		`CREATE TABLE IF NOT EXISTS invites (token TEXT PRIMARY KEY, expires bigint)`,
	}
	return dbTxCmdList(sdb.db, cmds)
}

func (sdb *sqliteedb) GetElection(id int64) (er *electionRecord, err error) {
	row := sdb.db.QueryRow(`SELECT data, owner, meta FROM elections WHERE ROWID = $1`, id)
	er = &electionRecord{Id: id}
	err = row.Scan(&er.Data, &er.Owner, &er.Meta)
	if err != nil {
		er = nil
	}
	return
}
func (sdb *sqliteedb) PutElection(er electionRecord) (newid int64, err error) {
	var result sql.Result
	if er.Id == 0 {
		result, err = sdb.db.Exec(`INSERT INTO elections (data, owner, meta) VALUES ($1, $2, $3)`, er.Data, er.Owner, er.Meta)
		if err != nil {
			err = fmt.Errorf("sqlite put election insert, %v", err)
			return
		}
		newid, err = result.LastInsertId()
		if err != nil {
			err = fmt.Errorf("sqlite put election wat, %v, %v", err, result)
			return
		}
		log.Printf("put new election as %d", newid)
		return
	}
	newid = er.Id
	_, err = sdb.db.Exec(`UPDATE elections SET data = $1, owner = $2, meta = $3 WHERE ROWID = $4`, er.Data, er.Owner, er.Meta, er.Id)
	return
}

func (sdb *sqliteedb) MakeInviteToken(token string, expires time.Time) (err error) {
	_, err = sdb.db.Exec(`INSERT INTO invites (token, expires) VALUES ($1, $2)`, token, expires.UTC().Unix())
	if err != nil {
		err = fmt.Errorf("invite put, %v", err)
	}
	return err
}
func (sdb *sqliteedb) PeekInviteToken(token string) (ok bool, expires time.Time, err error) {
	row := sdb.db.QueryRow(`SELECT expires from INVITES Where token = $1`, token)
	var expiresi int64
	err = row.Scan(&expiresi)
	if err == sql.ErrNoRows {
		return false, time.Time{}, err
	}
	et := time.Unix(expiresi, 0)
	return time.Now().UTC().Unix() < expiresi, et, nil
}
func (sdb *sqliteedb) UseInviteToken(token string) (ok bool, err error) {
	ok = false
	tx, err := sdb.db.Begin()
	if err != nil {
		err = fmt.Errorf("tx err, %v", err)
		return
	}
	defer tx.Rollback()
	row := tx.QueryRow(`SELECT expires FROM invites WHERE token = $1`, token)
	var expires time.Time
	err = row.Scan(&expires)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		err = fmt.Errorf("invite get, %v", err)
		return
	}
	if time.Now().After(expires) {
		return false, nil
	}
	_, err = tx.Exec(`DELETE FROM expires WHERE token = $1`, token)
	if err != nil {
		err = fmt.Errorf("invite del, %v", err)
		return
	}
	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("invite del commit, %v", err)
		return
	}
	return true, err
}
func (sdb *sqliteedb) GCInviteTokens() (err error) {
	_, err = sdb.db.Exec(`DELETE FROM expires WHERE expires < $1`, time.Now().UTC().Unix())
	return nil
}

func NewPostgresEDB(db *sql.DB) electionAppDB {
	return &postgresedb{db}
}

type postgresedb struct {
	db *sql.DB
}

// implement electionAppDB
func (sdb *postgresedb) Setup() error {
	cmds := []string{
		"CREATE TABLE IF NOT EXISTS elections (id bigserial, data TEXT, owner bigint, meta TEXT)",

		"CREATE TABLE IF NOT EXISTS metastate (k TEXT PRIMARY KEY, v bytea)",
		`CREATE TABLE IF NOT EXISTS invites (token text PRIMARY KEY, expires timestamp without time zone)`,
	}
	return dbTxCmdList(sdb.db, cmds)
}

func (sdb *postgresedb) GetElection(id int64) (er *electionRecord, err error) {
	return nil, errors.New("GetElection not implemented")
}
func (sdb *postgresedb) PutElection(er electionRecord) (newid int64, err error) {
	return 0, errors.New("PutElection not implemented")
}

func (sdb *postgresedb) MakeInviteToken(token string, expires time.Time) error {
	_, err := sdb.db.Exec(`INSERT INTO invites (token, expires) VALUES ($1, $2)`, token, expires.UTC())
	if err != nil {
		err = fmt.Errorf("invite put, %v", err)
	}
	return err
}
func (sdb *postgresedb) PeekInviteToken(token string) (ok bool, expires time.Time, err error) {
	row := sdb.db.QueryRow(`SELECT expires from INVITES Where token = $1`, token)
	//var expires time.Time
	err = row.Scan(&expires)
	if err == sql.ErrNoRows {
		return false, time.Time{}, nil
	}
	if err != nil {
		return false, expires, err
	}
	return expires.After(time.Now()), expires, nil
}
func (sdb *postgresedb) UseInviteToken(token string) (ok bool, err error) {
	ok = false
	tx, err := sdb.db.Begin()
	if err != nil {
		err = fmt.Errorf("tx err, %v", err)
		return
	}
	defer tx.Rollback()
	row := tx.QueryRow(`SELECT expires FROM invites WHERE token = $1`, token)
	var expires int64
	err = row.Scan(&expires)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		err = fmt.Errorf("invite get, %v", err)
		return
	}
	if time.Now().UTC().Unix() > expires {
		return false, nil
	}
	_, err = tx.Exec(`DELETE FROM expires WHERE token = $1`, token)
	if err != nil {
		err = fmt.Errorf("invite del, %v", err)
		return
	}
	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("invite del commit, %v", err)
		return
	}
	return true, err
}
func (sdb *postgresedb) GCInviteTokens() (err error) {
	var result sql.Result
	result, err = sdb.db.Exec(`DELETE FROM invites WHERE expires < CURRENT_TIMESTAMP AT TIME ZONE 'UTC'`)
	if err != nil {
		err = fmt.Errorf("invite del, %v", err)
	}
	rows, xe := result.RowsAffected()
	if rows > 0 || xe != nil {
		log.Printf("invite gc deleted %d, %v", rows, xe)
	}
	return
}

func dbTxCmdList(db *sql.DB, cmds []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // nop if committed
	for _, cmd := range cmds {
		_, err := tx.Exec(cmd)
		if err != nil {
			return fmt.Errorf("sql failed %#v, %v", cmd, err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit of %d commands failed, %v", len(cmds), err)
	}
	return nil
}

func gcThread(ctx context.Context, edb electionAppDB, period time.Duration) {
	t := time.NewTicker(period)
	defer t.Stop()
	for true {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			edb.GCInviteTokens()
		}
	}
}
