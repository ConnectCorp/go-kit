package utils

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultTokenLifetime is the default lifetime for a token (one year).
	DefaultTokenLifetime = time.Hour * 24 * 365
)

const (
	// ErrorInvalidKeyID is returned when an invalid key ID is provided.
	ErrorInvalidKeyID = "invalid key ID: %v"
	// ErrorInvalidKeyMaterial is returned when invalid key material is provided.
	ErrorInvalidKeyMaterial = "invalid key material"
	// ErrorInvalidIssuer is returned when an invalid issuer is provided.
	ErrorInvalidIssuer = "invalid issuer: %v"
	// ErrorInvalidAudience is returned when an invalid audience is provided.
	ErrorInvalidAudience = "invalid audience: %v"
	// ErrorUnableToSignToken is returned when signing a token fails.
	ErrorUnableToSignToken = "unable to sign token"
	// ErrorInvalidSubject is returned when an invalid subject is provided.
	ErrorInvalidSubject = "invalid subject"
	// ErrorInvalidToken is returned when validation fails due to an invalid token.
	ErrorInvalidToken = "invalid token"
	// ErrorInvalidTokenHeader  is returned when validation fails due to an invalid token header.
	ErrorInvalidTokenHeader = "invalid token header"
	// ErrorExpiredToken is returned when validation fails due to an expired token.
	ErrorExpiredToken = "expired token"
)

const (
	// TokenUserRole describes a user token.
	TokenUserRole = "user"
	// TokenSystemRole describes a system token.
	TokenSystemRole = "system"
)

const (
	currentTokenVersion = "v1"
	defaultSystemUserID = 0
	keyIDHeader         = "kid"
	tokenVersionHeader  = "v"
	subjectHeader       = "sub"
	roleHeader          = "role"
	issuerHeader        = "iss"
	audienceHeader      = "aud"
	expirationHeader    = "exp"
	issuedAtHeader      = "iat"
)

// TokenIssuer describes the capability of issuing tokens.
type TokenIssuer interface {
	IssueUserToken(sub int64) (string, error)
	IssueSystemToken() (string, error)
}

type tokenIssuer struct {
	keyID      string
	privateKey []byte
	issuer     string
	audience   string
	lifetime   time.Duration
}

// NewTokenIssuer initializes a new default TokenIssuer.
func NewTokenIssuer(keyID string, privateKey []byte, issuer, audience string, lifetime time.Duration) (TokenIssuer, error) {
	if err := validateParams(keyID, privateKey, issuer, audience); err != nil {
		return nil, err
	}
	return &tokenIssuer{
		keyID:      keyID,
		privateKey: privateKey,
		issuer:     issuer,
		audience:   audience,
		lifetime:   lifetime,
	}, nil
}

func (ti *tokenIssuer) IssueUserToken(sub int64) (string, error) {
	return ti.issueToken(sub, TokenUserRole)
}

func (ti *tokenIssuer) IssueSystemToken() (string, error) {
	return ti.issueToken(defaultSystemUserID, TokenSystemRole)
}

func (ti *tokenIssuer) issueToken(sub int64, role string) (string, error) {
	if role == TokenUserRole && sub <= 0 {
		return "", xerror.New(ErrorInvalidSubject, sub)
	}

	issuedAt := time.Now()
	t := jwt.New(jwt.SigningMethodRS256)

	t.Header[keyIDHeader] = ti.keyID
	t.Claims[tokenVersionHeader] = currentTokenVersion
	t.Claims[subjectHeader] = fmt.Sprintf("%v", sub)
	t.Claims[roleHeader] = role
	t.Claims[issuerHeader] = ti.issuer
	t.Claims[audienceHeader] = ti.audience
	t.Claims[issuedAtHeader] = issuedAt.Unix()
	t.Claims[expirationHeader] = issuedAt.Add(ti.lifetime).Unix()

	s, err := t.SignedString(ti.privateKey)
	if err != nil {
		return "", xerror.Wrap(err, ErrorUnableToSignToken, t)
	}
	return s, nil
}

// TokenVerifier describes the capability of verifying tokens.
type TokenVerifier interface {
	VerifyToken(t string) (int64, string, error)
}

type tokenVerifier struct {
	keyID     string
	publicKey []byte
	issuer    string
	audience  string
	jwtParser *jwt.Parser
}

// NewTokenVerifier initializes a new default TokenVerifier.
func NewTokenVerifier(keyID string, publicKey []byte, issuer, audience string) (TokenVerifier, error) {
	if err := validateParams(keyID, publicKey, issuer, audience); err != nil {
		return nil, err
	}
	return &tokenVerifier{
		keyID:     keyID,
		publicKey: publicKey,
		issuer:    issuer,
		audience:  audience,
		jwtParser: &jwt.Parser{UseJSONNumber: true},
	}, nil
}

func (tv *tokenVerifier) VerifyToken(t string) (int64, string, error) {
	dt, err := tv.jwtParser.Parse(t, tv.keyCallback)
	if err != nil {
		return 0, "", xerror.Wrap(err, ErrorInvalidToken, t)
	}

	v, err := safeGetStringClaim(dt, tokenVersionHeader)
	if err != nil {
		return 0, "", err
	}
	if v != currentTokenVersion {
		return 0, "", xerror.New(ErrorInvalidToken, t)
	}

	sub, err := safeGetStringClaimAsInt64(dt, subjectHeader)
	if err != nil {
		return 0, "", err
	}

	role, err := safeGetStringClaim(dt, roleHeader)
	if err != nil {
		return 0, "", err
	}
	if role != TokenUserRole && role != TokenSystemRole {
		return 0, "", xerror.New(ErrorInvalidToken, t)
	}
	if role == TokenSystemRole && sub != defaultSystemUserID {
		return 0, "", xerror.New(ErrorInvalidToken, t)
	}

	iss, err := safeGetStringClaim(dt, issuerHeader)
	if err != nil {
		return 0, "", err
	}
	if iss != tv.issuer {
		return 0, "", xerror.New(ErrorInvalidToken, t)
	}

	aud, err := safeGetStringClaim(dt, audienceHeader)
	if err != nil {
		return 0, "", err
	}
	if aud != tv.audience {
		return 0, "", xerror.New(ErrorInvalidToken, t)
	}

	exp, err := safeGetJSONNumberClaimAsInt64(dt, expirationHeader)
	if err != nil {
		return 0, "", err
	}
	if exp < time.Now().Unix() {
		return 0, "", xerror.New(ErrorExpiredToken, t)
	}

	return sub, role, nil
}

func (tv *tokenVerifier) keyCallback(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok || token.Header[keyIDHeader].(string) != tv.keyID {
		return nil, xerror.New(ErrorInvalidTokenHeader, token)
	}
	return tv.publicKey, nil
}

func validateParams(keyID string, key []byte, issuer, audience string) error {
	if !regexp.MustCompile("^k[0-9]$").MatchString(keyID) {
		return xerror.New(ErrorInvalidKeyID, keyID)
	}
	if len(key) == 0 {
		return xerror.New(ErrorInvalidKeyMaterial)
	}
	if len(issuer) == 0 || !strings.HasPrefix(issuer, "https://") {
		return xerror.New(ErrorInvalidIssuer, issuer)
	}
	if len(audience) == 0 || !strings.HasPrefix(audience, "connect-") {
		return xerror.New(ErrorInvalidAudience, audience)
	}
	return nil
}

func safeGetStringClaim(t *jwt.Token, claimName string) (string, error) {
	if claimValue, ok := t.Claims[claimName]; ok {
		if claimStr, ok := claimValue.(string); ok {
			return claimStr, nil
		}
		return "", xerror.New(ErrorInvalidToken, t)
	}
	return "", xerror.New(ErrorInvalidToken, t)
}

func safeGetStringClaimAsInt64(t *jwt.Token, claimName string) (int64, error) {
	claimStr, err := safeGetStringClaim(t, claimName)
	if err != nil {
		return 0, err
	}
	claimInt64, err := strconv.ParseInt(claimStr, 10, 64)
	if err != nil {
		return 0, xerror.Wrap(err, ErrorInvalidToken, t)
	}
	return claimInt64, nil
}

func safeGetJSONNumberClaimAsInt64(t *jwt.Token, claimName string) (int64, error) {
	if claimValue, ok := t.Claims[claimName]; ok {
		if claimJSONNumber, ok := claimValue.(json.Number); ok {
			claimInt64, err := claimJSONNumber.Int64()
			if err != nil {
				return 0, xerror.Wrap(err, ErrorInvalidToken, t)
			}
			return claimInt64, nil

		}
		return 0, xerror.New(ErrorInvalidToken, t)
	}
	return 0, xerror.New(ErrorInvalidToken, t)
}
