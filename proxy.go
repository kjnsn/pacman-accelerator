package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// proxy handles incoming requests and proxies them to upstream mirrors.
type proxy struct {
	httpClient *http.Client
	mirrorlist []url.URL
}

// Initialises a proxy with the given mirrorlist and attaches a http handler to the default mux.
func initProxyHandler(mirrorlist []url.URL, client *http.Client) {
	p := newProxy(mirrorlist, client)
	http.HandleFunc("/{repo}/{arch}/{prest...}", p.handleRequest)
}

func newProxy(mirrorlist []url.URL, client *http.Client) *proxy {
	return &proxy{
		client,
		mirrorlist,
	}
}

func (p *proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	mirrorUrl, err := p.mirrorUrl(r)
	if err != nil {
		sendError(err, w)
		return
	}

	mReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, mirrorUrl.String(), http.NoBody)
	if err != nil {
		sendError(err, w)
		return
	}

	startTime := time.Now()

	mRes, err := p.httpClient.Do(mReq)
	if err != nil {
		sendError(err, w)
		return
	}
	defer mRes.Body.Close()

	if mRes.StatusCode == http.StatusOK {
		// Copy relevant headers.
		copyHeader(mRes.Header, w.Header(), "Content-Length")
		copyHeader(mRes.Header, w.Header(), "Content-Type")
		copyHeader(mRes.Header, w.Header(), "ETag")
		copyHeader(mRes.Header, w.Header(), "Last-Modified")

		// Copy the data from the mirror back to the client.
		if _, err := io.Copy(w, mRes.Body); err != nil {
			log.Printf("Error while copying content: %v\n", err)
			return
		}

		if err := http.NewResponseController(w).Flush(); err != nil {
			log.Printf("Error while flushing the response: %v\n", err)
		}
	} else {
		w.WriteHeader(mRes.StatusCode)
	}

	elapsed := time.Since(startTime).String()
	log.Printf("Proxied request for %v to %v, took %v\n", r.URL.Path, mirrorUrl.String(), elapsed)
}

// Determines the url for a request to an upstream mirror.
func (p *proxy) mirrorUrl(r *http.Request) (*url.URL, error) {
	if len(p.mirrorlist) == 0 {
		return nil, errors.New("empty mirrorlist")
	}

	repo, arch, prest := r.PathValue("repo"), r.PathValue("arch"), r.PathValue("prest")
	if repo == "" {
		return nil, errors.New("$repo section of request is empty")
	}
	if arch == "" {
		return nil, errors.New("$arch section of request is empty")
	}
	if prest == "" {
		return nil, errors.New("no package specified in URL")
	}

	return url.Parse(
		strings.ReplaceAll(
			strings.ReplaceAll(p.mirrorlist[0].String(), "$repo", repo),
			"$arch", arch) + "/" + prest)
}

func sendError(err error, w http.ResponseWriter) {
	log.Printf("Error fetching from mirror: %v\n", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// Copies the header specified by key if it exists in src to dest.
func copyHeader(src http.Header, dest http.Header, key string) {
	if src.Get(key) != "" {
		dest.Set(key, src.Get(key))
	}
}
