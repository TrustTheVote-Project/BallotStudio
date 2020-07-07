package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	cbor "github.com/brianolson/cbor_go"
	"go.etcd.io/bbolt"
)

var BytesPerImageArchiveFile uint64 = 10000000

type ImageArchiver interface {
	ArchiveImage(imbytes []byte, r *http.Request)
}

// Archive images to files in archivedir.
// Will `mkdir -p archivedir`
func NewFileImageArchiver(archivedir string) (archie ImageArchiver, err error) {
	err = os.MkdirAll(archivedir, 0755)
	if err != nil {
		return
	}
	out := &fileImageArchiver{path: archivedir}
	err = out.ensureDupDB()
	if err != nil {
		return nil, err
	}
	return out, nil
}

type fileImageArchiver struct {
	path string

	dupdb *bbolt.DB

	fname string
	fpath string
	fout  io.WriteCloser
	lock  sync.Mutex

	foutBytesWritten uint64
}

type ArchiveImageMeta struct {
	Header     http.Header `cbor:"h"`
	RemoteAddr string      `cbor:"a"`
	Timestamp  int64       `cbor:"t"` // Java-time milliseconds since 1970
}

// TODO: more accessible format that is a tar of foo.jpg and foo.jpg_meta.json ?
type ArchiveImageRecord struct {
	Meta  ArchiveImageMeta `cbor:"m"`
	Image []byte           `cbor:"i"`
}

func JavaTime() int64 {
	now := time.Now()
	return (now.Unix() * 1000) + int64(now.Nanosecond()/1000000)
}

func (fia *fileImageArchiver) ArchiveImage(imbytes []byte, r *http.Request) {
	fia.lock.Lock()
	if fia.isDup(imbytes) {
		fia.lock.Unlock()
		return
	}
	fia.lock.Unlock()
	rec := ArchiveImageRecord{
		Meta: ArchiveImageMeta{
			Header:     r.Header,
			RemoteAddr: r.RemoteAddr,
			Timestamp:  JavaTime(),
		},
		Image: imbytes,
	}
	recbytes, err := cbor.Dumps(rec)
	if err != nil {
		log.Printf("ArchiveImage cbor dumps %s", err.Error())
		return
	}
	fia.lock.Lock()
	defer fia.lock.Unlock()
	if fia.fout == nil || fia.foutBytesWritten > BytesPerImageArchiveFile {
		err = fia.newFout()
		if err != nil {
			fia.fout = nil
			log.Printf("%s: ArchiveImage new %s", fia.fpath, err.Error())
			return
		}
	}
	_, err = fia.fout.Write(recbytes)
	if err != nil {
		fia.fout.Close()
		fia.fout = nil
		log.Printf("%s: ArchiveImage write %s", fia.fpath, err.Error())
	}
}

var imhashes = []byte("imh")
var trueByte = []byte("t")

// Checks image bytes against dup database, returns true if already seen.
// Records image bytes hash in dup database so that next time it will have been seen.
func (fia *fileImageArchiver) isDup(imbytes []byte) bool {
	hasher := fnv.New64a()
	hasher.Write(imbytes)
	var imhash [8]byte
	hasher.Sum(imhash[:0])

	var hit bool
	fia.dupdb.Update(func(tx *bbolt.Tx) error {
		bu := tx.Bucket(imhashes)
		val := bu.Get(imhash[:])
		hit = val != nil
		if !hit {
			bu.Put(imhash[:], trueByte)
		}
		return nil
	})
	return hit
}

func (fia *fileImageArchiver) newFout() (err error) {
	if fia.fout != nil {
		fia.fout.Close()
		fia.fout = nil
	}
	fia.fname = fmt.Sprintf("ima_%d_%d.cbor", JavaTime(), rand.Int31())
	fia.fpath = filepath.Join(fia.path, fia.fname)
	fia.fout, err = os.Create(fia.fpath)
	fia.foutBytesWritten = 0
	return
}

func (fia *fileImageArchiver) ensureDupDB() (err error) {
	dupdbpath := filepath.Join(fia.path, "dupdb")
	db, err := bbolt.Open(dupdbpath, 0600, nil)
	if err != nil {
		return fmt.Errorf("%s: %v", dupdbpath, err)
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(imhashes)
		return err
	})
	if err != nil {
		return fmt.Errorf("%s: %v", dupdbpath, err)
	}
	fia.dupdb = db
	return nil
}
