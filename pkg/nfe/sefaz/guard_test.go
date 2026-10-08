package sefaz_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/sefaz"
)

// Guard test URLs are chosen so that the tests stay offline even if the guard
// itself fails to activate (the exact bug under test):
//
//   - 127.0.0.2 belongs to the loopback range but is NOT on the guard's
//     allowlist, so an active guard must block it — and a broken guard would
//     only produce an instant local "connection refused", never a packet
//     towards a real SEFAZ/SVC endpoint. Real SEFAZ hostnames are therefore
//     deliberately not dialed from tests.
//   - Loopback targets (httptest servers on 127.0.0.1) are the allowlisted
//     pass-through case, which also keeps the door open for a future local
//     mock SEFAZ server.
const guardBlockedURL = "http://127.0.0.2:1/ws/NfeStatusServico4"

func newGuardTestClient(t *testing.T) *sefaz.Client {
	t.Helper()

	client, clientErr := sefaz.NewClient(config.SefazConfig{
		Environment: config.EnvironmentProduction,
		StateUF:     "MT",
		Timeout:     5 * time.Second,
	}, nil)
	if clientErr != nil {
		t.Fatalf("NewClient: unexpected error: %v", clientErr)
	}
	return client
}

func expectGuardError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected the guard to block the request, got nil error")
	}
	if !strings.Contains(err.Error(), "SEFAZ network guard") {
		t.Fatalf("expected a SEFAZ network guard error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "127.0.0.2:1") {
		t.Fatalf("expected the error to name the blocked target, got: %v", err)
	}
	if !strings.Contains(err.Error(), "NFE_SEFAZ_GUARD") {
		t.Fatalf("expected the error to mention NFE_SEFAZ_GUARD, got: %v", err)
	}
}

func TestGuard_BlocksNonLoopbackHost(t *testing.T) {
	t.Setenv("NFE_SEFAZ_GUARD", "block")

	_, err := newGuardTestClient(t).Post(guardBlockedURL, "action", []byte("<soap/>"))
	expectGuardError(t, err)
}

func TestGuard_UnknownValueFailsClosed(t *testing.T) {
	t.Setenv("NFE_SEFAZ_GUARD", "blok")

	_, err := newGuardTestClient(t).Post(guardBlockedURL, "action", []byte("<soap/>"))
	expectGuardError(t, err)
}

func TestGuard_AllowsLoopback(t *testing.T) {
	t.Setenv("NFE_SEFAZ_GUARD", "block")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	body, err := newGuardTestClient(t).Post(srv.URL, "action", []byte("<soap/>"))
	if err != nil {
		t.Fatalf("expected the guard to allow a loopback request, got: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", string(body))
	}
}

func TestGuard_AllowsLoopbackHostname(t *testing.T) {
	t.Setenv("NFE_SEFAZ_GUARD", "block")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	// Exercise the "localhost" allowlist branch (host without an IP literal).
	localhostURL := strings.Replace(srv.URL, "127.0.0.1", "localhost", 1)

	body, err := newGuardTestClient(t).Post(localhostURL, "action", []byte("<soap/>"))
	if err != nil {
		t.Fatalf("expected the guard to allow a localhost request, got: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", string(body))
	}
}

func TestGuard_InactiveDelegates(t *testing.T) {
	// Guard unset: the request must reach the base transport. Dialing the
	// closed loopback port produces a local connection error (never a guard
	// error), proving the guard passed the request through untouched.
	_, err := newGuardTestClient(t).Post(guardBlockedURL, "action", []byte("<soap/>"))
	if err == nil {
		t.Fatal("expected a dial error from the closed loopback port, got nil error")
	}
	if strings.Contains(err.Error(), "SEFAZ network guard") {
		t.Fatalf("guard is inactive but the request was blocked: %v", err)
	}
}

func TestGuard_ExplicitOff(t *testing.T) {
	for _, value := range []string{"off", "OFF", "false", "0"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("NFE_SEFAZ_GUARD", value)

			_, err := newGuardTestClient(t).Post(guardBlockedURL, "action", []byte("<soap/>"))
			if err == nil {
				t.Fatal("expected a dial error from the closed loopback port, got nil error")
			}
			if strings.Contains(err.Error(), "SEFAZ network guard") {
				t.Fatalf("guard is explicitly off but the request was blocked: %v", err)
			}
		})
	}
}
