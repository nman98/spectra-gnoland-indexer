package main

import (
	"net/http"
	"strings"
)

// txHashNormalizer rewrites the URL for transaction endpoints where a standard
// base64 tx hash contains '/' (a valid base64 character, but a URL path
// separator). It percent-encodes those slashes so chi can route correctly.
//
// Must be registered BEFORE middleware.CleanPath so that the double-slash a
// leading '/' in the hash would create is encoded before CleanPath collapses it.
//
// Handles both lookup paths:
//
//	/v1/transactions/<44-char-hash>
//	/v1/transactions/<44-char-hash>/messages
func txHashNormalizer(next http.Handler) http.Handler {
	const prefix = "/v1/transactions/"
	const hashLen = 44
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if the client already percent-encoded the URL.
		if r.URL.RawPath == "" && strings.HasPrefix(r.URL.Path, prefix) {
			rest := r.URL.Path[len(prefix):]
			if len(rest) >= hashLen {
				hashPart := rest[:hashLen]
				if strings.Contains(hashPart, "/") {
					suffix := rest[hashLen:]
					encoded := strings.ReplaceAll(hashPart, "/", "%2F")
					r.URL.RawPath = prefix + encoded + suffix
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
