package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var portFlag = flag.Uint64("addr", 6754, "Which port to listen on")
var mirrorlistFlag = flag.String("mirrorlist", "/etc/pacman.d/mirrorlist", "The path to the mirrorlist to use")
var forceHTTP2 = flag.Bool("force2", true, "Only use mirrors that support http2")

func main() {
	flag.Parse()

	mirrorlist, err := getMirrorlist(*mirrorlistFlag)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	mirror, err := findHTTP2Mirror(ctx, mirrorlist, http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}
	if mirror == nil && *forceHTTP2 {
		log.Fatal("Could not find any mirror that supports http2, and -force2 flag is enabled")
	} else {
		log.Printf("Found mirror that supports http2: %v\n", mirror)
		mirrorlist = []url.URL{*mirror}
	}

	initProxyHandler(mirrorlist, http.DefaultClient)

	log.Printf("Listening on port %v\n", *portFlag)
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
