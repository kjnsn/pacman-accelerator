package main

import (
	"net/url"
	"strings"
)

// Parses the given mirrorlist, returning a list of mirrors with supported protocols.
func parseMirrorlist(list string) ([]url.URL, error) {
	servers := []url.URL{}

	for _, line := range strings.Split(list, "\n") {
		lineTrimmed := strings.TrimSpace(line)
		if strings.HasPrefix(lineTrimmed, "#") {
			continue
		}

		fields := strings.Fields(lineTrimmed)
		if len(fields) < 3 || fields[0] != "Server" || fields[1] != "=" {
			continue
		}

		serverUrl, err := url.Parse(fields[2])
		if err != nil {
			return servers, err
		}
		servers = append(servers, *serverUrl)

	}

	return servers, nil
}
