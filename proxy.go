package main

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type proxy struct {
	httpClient *http.Client
	mirrorlist []url.URL
}

func newProxy(mirrorlist []url.URL) *proxy {
	return &proxy{
		&http.Client{},
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

	mRes, err := p.httpClient.Do(mReq)
	if err != nil {
		sendError(err, w)
		return
	}
	defer mRes.Body.Close()

	// Copy the data from the mirror back to the client.
	if _, err := io.Copy(w, mRes.Body); err != nil {
		sendError(err, w)
		return
	}
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
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
