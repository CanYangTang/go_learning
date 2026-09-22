package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Manager signs and verifies JWTs. The signing secret and token lifetime are
// injected in main.go, so the same configured Manager is shared by the login
// handler (which signs) and the auth middleware (which verifies).
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager builds a Manager with the given signing secret and token TTL.
func NewManager(secret []byte, ttl time.Duration) *Manager {
	return &Manager{secret: secret, ttl: ttl}
}

// Generate signs a token for userID with an expiry of now + m.ttl.
func (m *Manager) Generate(userID uint) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse verifies the signature and expiry and returns the userID from the
// token's Subject. Any verification failure (bad signature, expired, malformed,
// wrong algorithm) returns an error; the middleware translates it into a 401.
func (m *Manager) Parse(tokenString string) (uint, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		// 只接受 HMAC。不写这一步，攻击者可以把 alg 改成 "none"
		// 递一个无签名的 token 进来——这是 JWT 最经典的漏洞。
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return m.secret, nil
	})
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
