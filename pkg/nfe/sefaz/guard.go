package sefaz

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

// guardEnvVar is the environment variable that activates the SEFAZ network
// guard. When set to any non-empty value other than the explicit off values,
// the guard refuses outbound requests to any non-loopback host before any
// network I/O — making accidental real SEFAZ/SVC connections (production or
// homologação) impossible in test environments, while keeping loopback targets
// (e.g. a local mock SEFAZ server) working.
const guardEnvVar = "NFE_SEFAZ_GUARD"

// guardOffValues are the only values that explicitly disable the guard.
// Any other non-empty value activates it, so an accidental typo (e.g. "blok")
// fails closed instead of silently turning the guard off.
var guardOffValues = map[string]bool{
	"off":   true,
	"false": true,
	"0":     true,
}

var guardLogOnce sync.Once

// guardEnabled reports whether the SEFAZ network guard is active.
func guardEnabled() bool {
	value := os.Getenv(guardEnvVar)
	if value == "" {
		return false
	}
	if guardOffValues[strings.ToLower(strings.TrimSpace(value))] {
		return false
	}

	guardLogOnce.Do(func() {
		log.Printf("SEFAZ network guard: ACTIVE (only loopback allowed)")
	})

	return true
}

// isLoopbackHost reports whether host (port stripped, lowercased) is a loopback
// target the guard allows: localhost, 127.0.0.1, or ::1.
func isLoopbackHost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return false
}

// guardHost extracts the lowercased host (port stripped) from a request URL.
func guardHost(u *url.URL) (string, error) {
	if u == nil || u.Host == "" {
		return "", fmt.Errorf("SEFAZ network guard: request has no target host")
	}

	host := u.Host
	if hostname, _, splitErr := net.SplitHostPort(host); splitErr == nil {
		host = hostname
	}

	return strings.ToLower(strings.Trim(host, "[]")), nil
}

// guardTransport wraps the SEFAZ HTTP transport and, when the guard is active,
// refuses requests to non-loopback hosts before any dial. It is always
// installed by NewClient; activation is decided per request via
// NFE_SEFAZ_GUARD, so construction has a single path and production behavior
// (env unset) is an untouched pass-through.
type guardTransport struct {
	base http.RoundTripper
}

func (t *guardTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !guardEnabled() {
		return t.base.RoundTrip(req)
	}

	host, hostErr := guardHost(req.URL)
	if hostErr != nil {
		return nil, hostErr
	}

	if !isLoopbackHost(host) {
		return nil, fmt.Errorf(
			"SEFAZ network guard: request to %s blocked (%s=%s; only loopback is allowed)",
			req.URL.String(), guardEnvVar, os.Getenv(guardEnvVar),
		)
	}

	return t.base.RoundTrip(req)
}
