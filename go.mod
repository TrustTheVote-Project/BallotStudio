module bolson.org/~/src/ballotstudio

go 1.14

require (
	bolson.org/~/src/ballotscan v0.0.0
	bolson.org/~/src/login/login v0.0.0
	github.com/lib/pq v1.7.0
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/spakin/netpbm v1.3.0
)

replace bolson.org/~/src/login/login => ../login/login

replace bolson.org/~/src/httpcache => ../httpcache

replace bolson.org/~/src/ballotscan => ../ballotscan
