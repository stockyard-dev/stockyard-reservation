package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const publicKeyHex = "3af8f9593b3331c27994f1eeacf111c727ff6015016b0af44ed3ca6934d40b13"

type Limits struct {
	MaxItems     int
	Tier         string
	TrialEnd     string
	TrialExpired bool
}

func TrialLimits(trialEnd string) Limits {
	return Limits{MaxItems: 0, Tier: "trial", TrialEnd: trialEnd}
}

func PaidLimits() Limits {
	return Limits{MaxItems: 0, Tier: "paid"}
}

func NoLicense() Limits {
	return Limits{MaxItems: 0, Tier: "none"}
}

func ExpiredLimits() Limits {
	return Limits{MaxItems: 0, Tier: "expired", TrialExpired: true}
}

func DefaultLimits() Limits {
	key := os.Getenv("STOCKYARD_LICENSE_KEY")
	if key == "" {
		log.Printf("[license] No license key. Start a trial at https://stockyard.dev/for/")
		return NoLicense()
	}
	claims := validateLicenseKey(key, "reservation")
	if claims == nil {
		log.Printf("[license] Invalid license key")
		return NoLicense()
	}

	// Check tool access
	if claims.Tier != "individual" && claims.Tier != "*" {
		found := false
		for _, t := range claims.Tools {
			if t == "reservation" || t == "*" { found = true; break }
		}
		if !found {
			log.Printf("[license] Tool reservation not in licensed tools")
			return NoLicense()
		}
	}

	// Check trial
	if claims.TrialEnd != "" {
		te, _ := time.Parse(time.RFC3339, claims.TrialEnd)
		if time.Now().After(te) {
			// Trial expired — check if subscription is still active via expiry
			if claims.Exp > 0 && time.Now().Unix() <= claims.Exp {
				log.Printf("[license] Paid subscription active — unlimited")
				return PaidLimits()
			}
			log.Printf("[license] Trial expired")
			return ExpiredLimits()
		}
		days := int(time.Until(te).Hours() / 24)
		log.Printf("[license] Trial active — %d days remaining", days)
		return TrialLimits(claims.TrialEnd)
	}

	log.Printf("[license] Paid license valid — unlimited")
	return PaidLimits()
}

type licenseClaims struct {
	P        string   `json:"p"`
	Tier     string   `json:"tier"`
	Tools    []string `json:"tools"`
	TrialEnd string   `json:"trial_end"`
	Exp      int64    `json:"x"`
}

func validateLicenseKey(key, product string) *licenseClaims {
	if !strings.HasPrefix(key, "SY-") { return nil }
	key = key[3:]
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 { return nil }
	pb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil { return nil }
	sb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(sb) != ed25519.SignatureSize { return nil }
	pk, _ := hexDec(publicKeyHex)
	if len(pk) != ed25519.PublicKeySize { return nil }
	if !ed25519.Verify(ed25519.PublicKey(pk), pb, sb) { return nil }
	var claims licenseClaims
	if err := json.Unmarshal(pb, &claims); err != nil { return nil }
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp { return nil }
	if claims.P != "" && claims.P != "*" && claims.P != "stockyard" && claims.P != product { return nil }
	return &claims
}

func (s *Server) tierHandler(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"tier": s.limits.Tier}
	if s.limits.TrialEnd != "" {
		te, _ := time.Parse(time.RFC3339, s.limits.TrialEnd)
		m["trial_end"] = s.limits.TrialEnd
		m["days_remaining"] = int(time.Until(te).Hours() / 24)
	}
	if s.limits.TrialExpired {
		m["expired"] = true
		m["message"] = "Trial ended. Subscribe at stockyard.dev to continue."
	}
	if s.limits.Tier == "none" {
		m["message"] = "No license. Start a trial at stockyard.dev"
	}
	m["upgrade_url"] = "https://stockyard.dev/reservation/"
	wj(w, 200, m)
}

func hexDec(s string) ([]byte, error) {
	if len(s)%2 != 0 { return nil, os.ErrInvalid }
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		h, l := hv(s[i]), hv(s[i+1])
		if h == 255 || l == 255 { return nil, os.ErrInvalid }
		b[i/2] = h<<4 | l
	}
	return b, nil
}

func hv(c byte) byte {
	switch {
	case c >= '0' && c <= '9': return c - '0'
	case c >= 'a' && c <= 'f': return c - 'a' + 10
	case c >= 'A' && c <= 'F': return c - 'A' + 10
	}
	return 255
}
