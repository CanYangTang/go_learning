package auth

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func newTestManager() *Manager {
	return NewManager([]byte(testSecret), time.Hour)
}

func TestGenerateParseRoundTrip(t *testing.T) {
	m := newTestManager()

	token, err := m.Generate(42)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if token == "" {
		t.Fatalf("expected a non-empty token")
	}

	userID, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if userID != 42 {
		t.Fatalf("userID = %d, want 42", userID)
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	m := newTestManager()

	// Craft a token that expired a minute ago, signed with the same secret.
	claims := jwt.RegisteredClaims{
		Subject:   "42",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
	}
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to craft expired token: %v", err)
	}

	if _, err := m.Parse(expired); err == nil {
		t.Fatalf("Parse should reject an expired token")
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	signer := newTestManager()
	token, err := signer.Generate(42)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	verifier := NewManager([]byte("a-different-secret"), time.Hour)
	if _, err := verifier.Parse(token); err == nil {
		t.Fatalf("Parse should reject a token signed with a different secret")
	}
}

func TestParseRejectsAlgNone(t *testing.T) {
	m := newTestManager()

	// The classic "alg=none" forgery: an unsigned token. The keyfunc must
	// reject it because its method is not HMAC.
	claims := jwt.RegisteredClaims{Subject: "42"}
	none, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to craft alg=none token: %v", err)
	}

	if _, err := m.Parse(none); err == nil {
		t.Fatalf("Parse must reject an alg=none token")
	}
}

func TestParseRejectsMalformedToken(t *testing.T) {
	m := newTestManager()

	if _, err := m.Parse("not.a.valid.jwt"); err == nil {
		t.Fatalf("Parse should reject a malformed token")
	}
}

func TestParseRejectsNonNumericSubject(t *testing.T) {
	m := newTestManager()

	// A validly-signed token whose subject isn't a uint should fail conversion.
	claims := jwt.RegisteredClaims{
		Subject:   "not-a-number",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to craft token: %v", err)
	}

	if _, err := m.Parse(token); err == nil {
		t.Fatalf("Parse should reject a non-numeric subject")
	}
}

// Sanity check that the crafted expired-token error really is the expiry error,
// documenting the library behavior the middleware relies on.
func TestExpiredTokenErrorIsErrTokenExpired(t *testing.T) {
	claims := jwt.RegisteredClaims{
		Subject:   "1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Second)),
	}
	expired, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))

	var rc jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(expired, &rc, func(t *jwt.Token) (any, error) {
		return []byte(testSecret), nil
	})
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("err = %v, want errors.Is ErrTokenExpired", err)
	}
	// And the subject would parse fine if it weren't expired.
	if _, perr := strconv.ParseUint(rc.Subject, 10, 64); perr != nil {
		t.Fatalf("subject parse failed: %v", perr)
	}
}
