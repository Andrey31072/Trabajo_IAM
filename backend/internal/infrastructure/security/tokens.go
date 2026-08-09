package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"sena-iam-api/internal/config"
	"sena-iam-api/internal/domain"
)

type Claims struct {
	jwt.RegisteredClaims
	ActorType        string   `json:"actor_type"`
	ActorID          *string  `json:"actor_id,omitempty"`
	TrainingCenterID *string  `json:"training_center_id,omitempty"`
	Roles            []string `json:"roles"`
	Features         []string `json:"features"`
}

type TokenManager struct {
	issuer     string
	kid        string
	ttl        time.Duration
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewTokenManager(cfg config.Config) (*TokenManager, error) {
	privateKey, publicKey, err := loadKeys(cfg.JWTPrivateKeyB64, cfg.JWTPublicKeyB64)
	if err != nil {
		return nil, err
	}
	return &TokenManager{
		issuer:     cfg.JWTIssuer,
		kid:        cfg.JWTKID,
		ttl:        cfg.AccessTokenTTL,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (m *TokenManager) TTLSeconds() int {
	return int(m.ttl.Seconds())
}

func (m *TokenManager) Sign(user domain.User, auth domain.AuthUser) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
		ActorType:        user.ActorType,
		ActorID:          user.ActorID,
		TrainingCenterID: auth.TrainingCenterID,
		Roles:            auth.Roles,
		Features:         auth.Features,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = m.kid
	return token.SignedString(m.privateKey)
}

func (m *TokenManager) Verify(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodRS256 {
			return nil, errors.New("invalid signing method")
		}
		return m.publicKey, nil
	}, jwt.WithIssuer(m.issuer))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (m *TokenManager) JWKS() map[string]any {
	n := base64.RawURLEncoding.EncodeToString(m.publicKey.N.Bytes())
	eBytes := big.NewInt(int64(m.publicKey.E)).Bytes()
	e := base64.RawURLEncoding.EncodeToString(eBytes)
	return map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"kid": m.kid,
			"use": "sig",
			"alg": "RS256",
			"n":   n,
			"e":   e,
		}},
	}
}

func NewOpaqueToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func loadKeys(privateB64, publicB64 string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if strings.TrimSpace(privateB64) != "" && strings.TrimSpace(publicB64) != "" {
		privatePEM, err := base64.StdEncoding.DecodeString(privateB64)
		if err != nil {
			return nil, nil, err
		}
		publicPEM, err := base64.StdEncoding.DecodeString(publicB64)
		if err != nil {
			return nil, nil, err
		}
		privateKey, err := parsePrivateKey(privatePEM)
		if err != nil {
			return nil, nil, err
		}
		publicKey, err := parsePublicKey(publicPEM)
		if err != nil {
			return nil, nil, err
		}
		return privateKey, publicKey, nil
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

func parsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid private key pem")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func parsePublicKey(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid public key pem")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not rsa")
	}
	return rsaKey, nil
}

func JSONStringList(value any) []string {
	data, _ := json.Marshal(value)
	var out []string
	_ = json.Unmarshal(data, &out)
	return out
}