package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

const publicKeyHex = "3af8f9593b3331c27994f1eeacf111c727ff6015016b0af44ed3ca6934d40b13"

type Limits struct {
	MaxItems int
	Tier     string
}

func FreeLimits() Limits {
	return Limits{MaxItems: 5, Tier: "free"}
}

func ProLimits() Limits {
	return Limits{MaxItems: 0, Tier: "pro"}
}

func DefaultLimits() Limits {
	key := os.Getenv("STOCKYARD_LICENSE_KEY")
	if key == "" {
		log.Printf("[license] Free tier (5 items). Set STOCKYARD_LICENSE_KEY for Pro.")
		log.Printf("[license] Get a key at https://stockyard.dev/reservation/")
		return FreeLimits()
	}
	if validateLicenseKey(key, "reservation") {
		log.Printf("[license] Pro license valid — unlimited")
		return ProLimits()
	}
	log.Printf("[license] Invalid key — free tier")
	return FreeLimits()
}

func LimitReached(limit, current int) bool {
	if limit == 0 { return false }
	return current >= limit
}

func validateLicenseKey(key, product string) bool {
	if !strings.HasPrefix(key, "SY-") { return false }
	key = key[3:]
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 { return false }
	pb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil { return false }
	sb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(sb) != ed25519.SignatureSize { return false }
	pk, _ := hexDec(publicKeyHex)
	if len(pk) != ed25519.PublicKeySize { return false }
	if !ed25519.Verify(ed25519.PublicKey(pk), pb, sb) { return false }
	var p struct { P string `json:"p"`; X int64 `json:"x"` }
	if err := json.Unmarshal(pb, &p); err != nil { return false }
	if p.X > 0 && time.Now().Unix() > p.X { return false }
	if p.P != "*" && p.P != "stockyard" && p.P != product { return false }
	return true
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
