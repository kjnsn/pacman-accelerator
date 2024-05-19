package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var portFlag = flag.Uint64("addr", 6754, "Which port to listen on")
var mirrorlistFlag = flag.String("mirrorlist", "/etc/pacman.d/mirrorlist", "The path to the mirrorlist to use")

func main() {
	flag.Parse()

	mirrorlist, err := getMirrorlist(*mirrorlistFlag)
	if err != nil {
		log.Fatal(err)
	}

	initProxyHandler(mirrorlist, http.DefaultClient)

	if err := http.ListenAndServe("localhost:"+strconv.FormatUint(*portFlag, 10), nil); err != nil {
		log.Fatal(err)
	}
}

func getMirrorlist(path string) ([]url.URL, error) {
	urls := []url.URL{}

	f, err := os.Open(path)
	if err != nil {
		return urls, err
	}

	defer f.Close()

	b, err := io.ReadAll(f)
	if (err != nil) && err != io.EOF {
		return urls, err
	}

	return parseMirrorlist(string(b))
}
