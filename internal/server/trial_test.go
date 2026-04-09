package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stockyard-dev/stockyard-reservation/internal/store"
)

// Tests for the additive trial-gate middleware on reservation. Booking
// already has a richer license model than the standard-shape tools
// (TrialLimits/PaidLimits/NoLicense/ExpiredLimits with claim-based
// trial_end and tool-access), so this test only verifies the new
// behavior layered on top:
//
//   1. shouldBlockWrite returns true for tier=="none" + write methods
//   2. shouldBlockWrite returns true for tier=="expired" + write methods
//   3. shouldBlockWrite returns false for tier=="paid" + writes (allowed)
//   4. shouldBlockWrite returns false for tier=="trial" + writes (allowed)
//   5. shouldBlockWrite returns false for read methods regardless of tier
//   6. shouldBlockWrite returns false for /api/license/activate POST
//   7. End-to-end: tier=="none" → POST /api/reservations returns 402
//   8. End-to-end: tier=="paid" → POST /api/reservations returns 200
//   9. End-to-end: tier=="none" → GET /api/reservations still works (read-through)
//  10. End-to-end: tier=="none" → POST /api/license/activate is reachable
//      (returns 400 for bogus key, but the route is not blocked)
//  11. activateLicense returns 400 on missing license_key field
//  12. activateLicense returns 400 on invalid SY-... key
//  13. PersistLicense writes to dataDir/license.txt at 0600
//  14. loadLicenseFromDisk reads what PersistLicense wrote
//  15. ValidateLicenseKeyExported returns false for bogus keys
//
// We deliberately don't test the activate happy path with a real
// signed payload — the validateLicenseKey closure expects an Ed25519
// signature against the embedded public key, and manufacturing one
// in-test would either require leaking the private key or stubbing
// out the validator. The Option B trial_test.go pattern handles this
// for the standard-shape tools by injecting a test pubkey, but
// reservation's existing limits.go doesn't expose a hook for that.

func newTestDB(t *testing.T) *store.DB {
	t.Helper()
	dir, err := os.MkdirTemp("", "reservation-trial-test-*")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	db, err := store.Open(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	return db
}

func newTestServer(t *testing.T, tier string) *Server {
	t.Helper()
	dir, err := os.MkdirTemp("", "reservation-trial-srv-*")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	db := newTestDB(t)
	var lim Limits
	switch tier {
	case "none":
		lim = NoLicense()
	case "expired":
		lim = ExpiredLimits()
	case "trial":
		lim = TrialLimits("2099-12-31T00:00:00Z")
	case "paid":
		lim = PaidLimits()
	default:
		t.Fatalf("unknown test tier: %s", tier)
	}
	return New(db, lim, dir)
}

func TestShouldBlockWrite_NoneTier(t *testing.T) {
	s := newTestServer(t, "none")
	cases := []struct {
		method string
		want   bool
	}{
		{"POST", true},
		{"PUT", true},
		{"DELETE", true},
		{"PATCH", true},
		{"GET", false},
		{"HEAD", false},
		{"OPTIONS", false},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, "/api/reservations", nil)
		got := s.shouldBlockWrite(req)
		if got != c.want {
			t.Errorf("none/%s: got %v, want %v", c.method, got, c.want)
		}
	}
}

func TestShouldBlockWrite_ExpiredTier(t *testing.T) {
	s := newTestServer(t, "expired")
	req := httptest.NewRequest("POST", "/api/reservations", nil)
	if !s.shouldBlockWrite(req) {
		t.Error("expired tier should block POST writes")
	}
	req = httptest.NewRequest("GET", "/api/reservations", nil)
	if s.shouldBlockWrite(req) {
		t.Error("expired tier should NOT block GET reads")
	}
}

func TestShouldBlockWrite_PaidTier(t *testing.T) {
	s := newTestServer(t, "paid")
	for _, m := range []string{"POST", "PUT", "DELETE", "PATCH", "GET"} {
		req := httptest.NewRequest(m, "/api/reservations", nil)
		if s.shouldBlockWrite(req) {
			t.Errorf("paid tier should NOT block %s", m)
		}
	}
}

func TestShouldBlockWrite_TrialTier(t *testing.T) {
	s := newTestServer(t, "trial")
	for _, m := range []string{"POST", "PUT", "DELETE", "GET"} {
		req := httptest.NewRequest(m, "/api/reservations", nil)
		if s.shouldBlockWrite(req) {
			t.Errorf("trial tier should NOT block %s — trial users have full write access", m)
		}
	}
}

func TestShouldBlockWrite_ActivateAllowlisted(t *testing.T) {
	s := newTestServer(t, "none")
	req := httptest.NewRequest("POST", "/api/license/activate", nil)
	if s.shouldBlockWrite(req) {
		t.Error("/api/license/activate must be reachable in none tier (it's the only way out)")
	}
}

func TestEndToEnd_NoneTierBlocksWrite(t *testing.T) {
	s := newTestServer(t, "none")
	body := bytes.NewBufferString(`{"guest_name":"Alice","date":"2026-05-01","time":"19:00"}`)
	req := httptest.NewRequest("POST", "/api/reservations", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusPaymentRequired {
		t.Errorf("expected 402, got %d. body: %s", w.Code, w.Body.String())
	}
}

func TestEndToEnd_PaidTierAllowsWrite(t *testing.T) {
	s := newTestServer(t, "paid")
	body := bytes.NewBufferString(`{"guest_name":"Alice","date":"2026-05-01","time":"19:00"}`)
	req := httptest.NewRequest("POST", "/api/reservations", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Errorf("paid tier should allow POST, got %d. body: %s", w.Code, w.Body.String())
	}
}

func TestEndToEnd_NoneTierAllowsRead(t *testing.T) {
	s := newTestServer(t, "none")
	req := httptest.NewRequest("GET", "/api/reservations", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("none tier should allow GET reads, got %d", w.Code)
	}
}

func TestEndToEnd_ActivateReachableInNoneTier(t *testing.T) {
	s := newTestServer(t, "none")
	body := bytes.NewBufferString(`{"license_key":"SY-bogus"}`)
	req := httptest.NewRequest("POST", "/api/license/activate", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	// We expect 400 (bogus key) not 402 (blocked) — proves the route
	// is reachable even when writes are otherwise locked.
	if w.Code == http.StatusPaymentRequired {
		t.Error("/api/license/activate must be reachable in none tier — got 402, expected 400 for bogus key")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bogus key, got %d. body: %s", w.Code, w.Body.String())
	}
}

func TestActivate_MissingKey(t *testing.T) {
	s := newTestServer(t, "none")
	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest("POST", "/api/license/activate", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing license_key, got %d", w.Code)
	}
}

func TestActivate_InvalidKey(t *testing.T) {
	s := newTestServer(t, "none")
	body := bytes.NewBufferString(`{"license_key":"not-a-real-key"}`)
	req := httptest.NewRequest("POST", "/api/license/activate", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid key, got %d", w.Code)
	}
}

func TestPersistLicense_WritesFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "reservation-persist-test-*")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	defer os.RemoveAll(dir)
	if err := PersistLicense(dir, "SY-test-key-payload.signature"); err != nil {
		t.Fatalf("PersistLicense: %v", err)
	}
	path := filepath.Join(dir, licenseFilename)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// File mode should be 0600 (owner read/write only).
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 perms, got %o", info.Mode().Perm())
	}
}

func TestLoadLicenseFromDisk_RoundTrip(t *testing.T) {
	dir, err := os.MkdirTemp("", "reservation-load-test-*")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	defer os.RemoveAll(dir)
	want := "SY-test-key-payload.signature"
	if err := PersistLicense(dir, want); err != nil {
		t.Fatalf("PersistLicense: %v", err)
	}
	got := loadLicenseFromDisk(dir)
	if got != want {
		t.Errorf("round trip: got %q, want %q", got, want)
	}
}

func TestLoadLicenseFromDisk_MissingReturnsEmpty(t *testing.T) {
	dir, err := os.MkdirTemp("", "reservation-missing-test-*")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	defer os.RemoveAll(dir)
	got := loadLicenseFromDisk(dir)
	if got != "" {
		t.Errorf("expected empty string for missing file, got %q", got)
	}
}

func TestValidateLicenseKeyExported_BogusKey(t *testing.T) {
	if ValidateLicenseKeyExported("not-a-key") {
		t.Error("expected false for non-SY key")
	}
	if ValidateLicenseKeyExported("SY-fake.fake") {
		t.Error("expected false for malformed SY key")
	}
	if ValidateLicenseKeyExported("") {
		t.Error("expected false for empty string")
	}
}
