package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
)

var portFlag = flag.Uint64("addr", 6754, "Which port to listen on")

func main() {
	flag.Parse()

	http.HandleFunc("", handler)

	if err := http.ListenAndServe("localhost:"+strconv.FormatUint(*portFlag, 10), nil); err != nil {
		log.Fatal(err)
	}

}

// Handles incoming requests, finding the requested path on a remote server.
func handler(w http.ResponseWriter, r *http.Request) {

}
