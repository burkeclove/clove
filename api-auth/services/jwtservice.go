package services

import (
	"os"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/burkeclove/auth-api/internal"
)

type JwtService struct {
	cfg        internal.Config
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewFromPEM constructs the service from PEM strings
func NewJwtService(cfg internal.Config, privateKeyPEM, publicKeyPEM string) (*JwtService, error) {
	priv, err := parseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	pub, err := parseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	normalizeConfig(&cfg)
	return &JwtService{cfg: cfg, privateKey: priv, publicKey: pub}, nil
}

// NewFromKeys constructs the service from already-parsed RSA keys.
func NewFromKeys(cfg internal.Config, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) (*JwtService, error) {
	if privateKey == nil || publicKey == nil {
		return nil, errors.New("privateKey and publicKey are required")
	}
	normalizeConfig(&cfg)
	return &JwtService{cfg: cfg, privateKey: privateKey, publicKey: publicKey}, nil
}

// Mint creates a signed JWT containing user_id and email.
// Subject ("sub") is set to userID by convention; JTI is randomized.
func (s *JwtService) Mint(ctx context.Context, userID, email string) (tokenString string, expiresAt time.Time, err error) {
	userID = strings.TrimSpace(userID)
	email = strings.TrimSpace(email)

	if userID == "" {
		return "", time.Time{}, errors.New("userID is required")
	}
	if email == "" {
		return "", time.Time{}, errors.New("email is required")
	}

	now := time.Now().UTC()
	exp := now.Add(s.cfg.AccessTTL)

	claims := &internal.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{s.cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-10 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        randomJTI(),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if s.cfg.KeyID != "" {
		tok.Header["kid"] = s.cfg.KeyID
	}

	signed, err := tok.SignedString(s.privateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, exp, nil
}

// Validate verifies signature and validates issuer/audience/time-based registered claims.
// Returns typed claims on success.
func (s *JwtService) Validate(ctx context.Context, tokenString string) (*internal.Claims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(s.cfg.Issuer),
		jwt.WithAudience(s.cfg.Audience),
		jwt.WithLeeway(s.cfg.Leeway),
	)

	claims := &internal.Claims{}
	_, err := parser.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return s.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Ensure required custom claims exist.
	if strings.TrimSpace(claims.UserID) == "" || strings.TrimSpace(claims.Email) == "" {
		return nil, errors.New("missing required claims: user_id and/or email")
	}

	return claims, nil
}

func normalizeConfig(cfg *internal.Config) {
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 15 * time.Minute
	}
	if cfg.Leeway <= 0 {
		cfg.Leeway = 30 * time.Second
	}
	// Issuer/Audience are logically required for tight validation; fail fast at runtime if missing.
	// We don’t hard-error here to keep constructor simple; you can decide policy.
}

func randomJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func parseRSAPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("no PEM block found")
	}

	// PKCS#1
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// PKCS#8
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return key, nil
}

func parseRSAPublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("no PEM block found")
	}

	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		pub, ok := pubAny.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not an RSA public key")
		}
		return pub, nil
	}

	pub, err2 := x509.ParsePKCS1PublicKey(block.Bytes)
	if err2 != nil {
		return nil, err
	}
	return pub, nil
}

func readPEMFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(b), nil
}

// signToken signs a JWT token with the service's private key
func (s *JwtService) signToken(token *jwt.Token) (string, error) {
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// GetPublicKey returns the RSA public key for external validation
func (s *JwtService) GetPublicKey() *rsa.PublicKey {
	return s.publicKey
}
