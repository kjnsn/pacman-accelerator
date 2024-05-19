package main

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestMirrorUrl_EmptyError(t *testing.T) {
	mUrl, _ := url.Parse("http://example.com/archlinux/$repo/os/$arch")
	p := newProxy([]url.URL{*mUrl})
	r := httptest.NewRequest("", "/", nil)

	_, err := p.mirrorUrl(r)
	if err == nil {
		t.Errorf("wanted error, got nil")
	}
}

func TestMirrorUrl_ReplacesValues(t *testing.T) {
	mUrl, _ := url.Parse("http://example.com/archlinux/$repo/os/$arch")
	p := newProxy([]url.URL{*mUrl})
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
}
