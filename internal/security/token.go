package security

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Tokens struct{ secret []byte }

func New(secret string) (*Tokens, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET must contain at least 32 bytes")
	}
	return &Tokens{secret: []byte(secret)}, nil
}

func (t *Tokens) Issue(userID uint) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{Subject: strconv.FormatUint(uint64(userID), 10), Issuer: "devon-articles", Audience: jwt.ClaimStrings{"devon-articles-api"}, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour))}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
}

func (t *Tokens) Parse(raw string) (uint, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) { return t.secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("devon-articles"), jwt.WithAudience("devon-articles-api"), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 32)
	if err != nil || id == 0 {
		return 0, errors.New("invalid token subject")
	}
	return uint(id), nil
}
