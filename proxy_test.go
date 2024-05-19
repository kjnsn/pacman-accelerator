package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
)

func TestMirrorUrl(t *testing.T) {
	mUrl, _ := url.Parse("http://example.com/archlinux/$repo/os/$arch")
	p := newProxy([]url.URL{*mUrl}, http.DefaultClient)

	t.Run("empty url", func(t *testing.T) {
		r := httptest.NewRequest("", "/", nil)

		_, err := p.mirrorUrl(r)
		if err == nil {
			t.Errorf("wanted error, got nil")
		}
	})

	t.Run("replaces url params", func(t *testing.T) {
		r := httptest.NewRequest("", "/archlinux/foo/os/bar/core/packagebaz.tar.xz", nil)
		r.SetPathValue("repo", "foo")
		r.SetPathValue("arch", "bar")
		r.SetPathValue("prest", "core/packagebaz.tar.xz")

		newUrl, err := p.mirrorUrl(r)
		if err != nil {
			t.Errorf("wanted no error, got %v", err)
		}

		if newUrl.String() != "http://example.com/archlinux/foo/os/bar/core/packagebaz.tar.xz" {
			t.Errorf("wanted new url to be http://example.com/archlinux/foo/os/bar/core/packagebaz.tar.xz, got %v", newUrl)
		}
	})
}

func TestHandler(t *testing.T) {
	createProxy := func(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *proxy {
		ts := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(ts.Close)

		tsUrl, _ := url.Parse(ts.URL + "/archlinux/$repo/os/$arch")
		return newProxy([]url.URL{*tsUrl}, ts.Client())
	}

	t.Run("404 upstream", func(t *testing.T) {
		var upstreamReqPath string
		p := createProxy(t, func(w http.ResponseWriter, r *http.Request) {
			upstreamReqPath = r.URL.Path
			w.WriteHeader(http.StatusNotFound)
		})

		r := httptest.NewRequest(http.MethodGet, "/core/x86_64/package.tar.xz", nil)
		r.SetPathValue("repo", "core")
		r.SetPathValue("arch", "x86_64")
		r.SetPathValue("prest", "package.tar.xz")
		w := httptest.NewRecorder()

		p.handleRequest(w, r)

		if w.Result().StatusCode != http.StatusNotFound {
			t.Errorf("result status code: wanted 404, got %v", w.Result().StatusCode)
		}

		if upstreamReqPath != "/archlinux/core/os/x86_64/package.tar.xz" {
			t.Errorf("mirror url: wanted /archlinux/core/os/x86_64/package.tar.xz, got %v", upstreamReqPath)
		}
	})

	t.Run("streams package", func(t *testing.T) {
		packageBytes := []byte{0x01, 0x02, 0x03}
		p := createProxy(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("ETag", "foobar")
			w.Header().Set("Content-Type", "plain/text")
			w.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
			w.Write(packageBytes)
		})

		r := httptest.NewRequest(http.MethodGet, "/core/x86_64/package.tar.xz", nil)
		r.SetPathValue("repo", "core")
		r.SetPathValue("arch", "x86_64")
		r.SetPathValue("prest", "package.tar.xz")
		w := httptest.NewRecorder()

		p.handleRequest(w, r)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("result status code: wanted 200, got %v", w.Result().StatusCode)
		}

		b, err := io.ReadAll(w.Result().Body)
		if err != nil {
			t.Fatalf("failed reading response bytes, error %v", err)
		}

		if slices.Compare(b, packageBytes) != 0 {
			t.Error("response bytes do not equal packageBytes")
		}

		if w.Result().Header.Get("ETag") != "foobar" {
			t.Errorf("response header ETag: wanted foobar, got %v", w.Result().Header.Get("ETag"))
		}

		if w.Result().Header.Get("Content-Type") != "plain/text" {
			t.Errorf("response header Content-Type: wanted plain/text, got %v", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("Last-Modified") != "Wed, 21 Oct 2015 07:28:00 GMT" {
			t.Errorf("response header Last-Modified: wanted Wed, 21 Oct 2015 07:28:00 GMT, got %v",
				w.Result().Header.Get("Last-Modified"))
		}
	})
}

func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
}
