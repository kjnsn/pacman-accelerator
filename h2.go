package main

import (
	"context"
	"net/http"
	"net/url"
)

// Finds the first http2 mirror in the list and returns it's URL.
// Responds with nil if no mirror supporting http2 could be found.
func findHTTP2Mirror(ctx context.Context, mirrors []url.URL, c *http.Client) (*url.URL, error) {
	for _, mirror := range mirrors {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, mirror.String(), nil)
		if err != nil {
			return nil, err
		}

		res, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.ProtoMajor == 2 {
			return &mirror, nil
		}
	}

	return nil, nil
}
