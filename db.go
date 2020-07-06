package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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

type postgresedb struct {
	db *sql.DB
}

// implement electionAppDB
func (sdb *postgresedb) Setup() error {
	cmds := []string{
		"CREATE TABLE IF NOT EXISTS elections (id bigserial, data TEXT, owner bigint, meta TEXT)",

		"CREATE TABLE IF NOT EXISTS metastate (k TEXT PRIMARY KEY, v bytea)",
	}
	return dbTxCmdList(sdb.db, cmds)
}

func (sdb *postgresedb) GetElection(id int64) (er *electionRecord, err error) {
	return nil, errors.New("GetElection not implemented")
}
func (sdb *postgresedb) PutElection(er electionRecord) (newid int64, err error) {
	return 0, errors.New("PutElection not implemented")
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
