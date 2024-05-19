package main

import "testing"

var validMirrorlist = `
##
## Arch Linux repository mirrorlist
## Filtered by mirror score from mirror status page
## Generated on 2024-05-18
##

Server = http://localhost:6754/$repo/$arch

## Australia
#Server = http://au.mirrors.cicku.me/archlinux/$repo/os/$arch
## Australia
#Server = http://gsl-syd.mm.fcix.net/archlinux/$repo/os/$arch
## Australia
Server = https://syd.mirror.rackspace.com/archlinux/$repo/os/$arch
## Australia
#Server = http://mirror.internode.on.net/pub/archlinux/$repo/os/$arch
## Australia
Server = ftp://gsl-syd.mm.fcix.net/archlinux/$repo/os/$arch
## Australia
Server = https://sydney.mirror.pkgbuild.com/$repo/os/$arch
## Australia
`

func TestParseMirrorlist(t *testing.T) {
	urls, err := parseMirrorlist(validMirrorlist)

	if err != nil {
		t.Errorf("wanted no err, got %v", err)
	}

	if len(urls) != 2 {
		t.Errorf("Wanted len(urls) to be 2, got %v", len(urls))
	}

	if urls[0].String() != "https://syd.mirror.rackspace.com/archlinux/$repo/os/$arch" {
		t.Errorf("Wanted urls[0] to be https://syd.mirror.rackspace.com/archlinux/$repo/os/$arch, got %v", urls[0])
	}
}
