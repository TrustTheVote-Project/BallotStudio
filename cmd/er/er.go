// Test utility for data/election.go
//
// Read election report json from stdin, do fixup, write result to stdout
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/brianolson/ballotstudio/data"
)

func maybefail(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func main() {
	//data.DebugOut = os.Stderr
	dec := json.NewDecoder(os.Stdin)
	var er map[string]interface{}
	err := dec.Decode(&er)
	maybefail(err, "json decode, %v", err)
	ner := data.Fixup(er)
	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(ner)
	maybefail(err, "json encode, %v", err)
}
