module github.com/brianolson/ballotstudio

go 1.14

require (
	github.com/brianolson/ballotscan v0.0.0
	github.com/brianolson/login/login v0.0.0
	github.com/brianolson/cbor_go v1.0.0
	github.com/lib/pq v1.7.0
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/spakin/netpbm v1.3.0
	go.etcd.io/bbolt v1.3.5
)

replace github.com/brianolson/login/login => ../login/login

replace github.com/brianolson/httpcache => ../httpcache

replace github.com/brianolson/ballotscan => ../ballotscan
