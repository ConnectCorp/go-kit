package utils

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultAccessTokenLifetime is the default lifetime for an access token.
	DefaultAccessTokenLifetime = time.Hour * 24
	// DefaultRefreshTokenLifetime is the default lifetime for a refresh token.
	DefaultRefreshTokenLifetime = time.Hour * 24 * 365
	// DefaultSinglePurposeTokenLifetime is the default lifetime for a single purpose token.
	DefaultSinglePurposeTokenLifetime = time.Hour
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
	// ErrorInvalidRole is returned when an invalid role is provided.
	ErrorInvalidRole = "invalid role"
	// ErrorInvalidToken is returned when validation fails due to an invalid token.
	ErrorInvalidToken = "invalid token"
	// ErrorInvalidTokenHeader  is returned when validation fails due to an invalid token header.
	ErrorInvalidTokenHeader = "invalid token header"
	// ErrorExpiredToken is returned when validation fails due to an expired token.
	ErrorExpiredToken = "expired token"
	// ErrorInvalidCustomClaim is returned when an invalid custom claim is provided.
	ErrorInvalidCustomClaim = "invalid custom claim: %v"
	// ErrorMissingCustomClaims is returned when some custom claims are missing.
	ErrorMissingCustomClaims = "missing custom claims: %v"
)

const (
	// TokenAccessUserRole describes an access user token.
	TokenAccessUserRole = "user"
	// TokenRefreshUserRole describes a refresh user token.
	TokenRefreshUserRole = "user-refresh"
	// TokenAccessSystemRole describes an access system token.
	TokenAccessSystemRole = "system"
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

// SinglePurposeTokenDescriptor describes the settings for issuing and verifying a single purpose token.
type SinglePurposeTokenDescriptor interface {
	GetRole() string
	IsSubMeaningful() bool
	GetLifetime() time.Duration
	GetAllowedCustomClaims() map[string]reflect.Type
}

type singlePurposeTokenDescriptor struct {
	role                string
	isSubMeaningful     bool
	lifetime            time.Duration
	allowedCustomClaims map[string]reflect.Type
}

// NewSinglePurposeTokenDescriptor initializes a new SinglePurposeTokenDescriptor.
func NewSinglePurposeTokenDescriptor(role string, isSubMeaningful bool, lifetime time.Duration, allowedCustomClaims map[string]reflect.Type) SinglePurposeTokenDescriptor {
	return &singlePurposeTokenDescriptor{
		role:                role,
		isSubMeaningful:     isSubMeaningful,
		lifetime:            lifetime,
		allowedCustomClaims: allowedCustomClaims,
	}
}

// GetRole implements the SinglePurposeTokenDescriptor interface.
func (s *singlePurposeTokenDescriptor) GetRole() string {
	return s.role
}

// IsSubMeaningful implements the SinglePurposeTokenDescriptor interface.
func (s *singlePurposeTokenDescriptor) IsSubMeaningful() bool {
	return s.isSubMeaningful
}

// GetLifetime implements the SinglePurposeTokenDescriptor interface.
func (s *singlePurposeTokenDescriptor) GetLifetime() time.Duration {
	return s.lifetime
}

// GetCustomClaims implements the SinglePurposeTokenDescriptor interface.
func (s *singlePurposeTokenDescriptor) GetAllowedCustomClaims() map[string]reflect.Type {
	return s.allowedCustomClaims
}

// TokenIssuer describes the capability of issuing tokens.
type TokenIssuer interface {
	IssueAccessUserToken(sub int64) (string, error)
	IssueRefreshUserToken(sub int64) (string, error)
	IssueAccessSystemToken() (string, error)
	IssueSinglePurposeToken(descriptor SinglePurposeTokenDescriptor, sub int64, customClaims map[string]interface{}) (string, error)
}

type tokenIssuer struct {
	keyID           string
	privateKey      []byte
	issuer          string
	audience        string
	refreshLifetime time.Duration
	accessLifetime  time.Duration
}

// NewTokenIssuer initializes a new default TokenIssuer.
func NewTokenIssuer(keyID string, privateKey []byte, issuer, audience string, refreshLifetime, accessLifetime time.Duration) (TokenIssuer, error) {
	if err := validateParams(keyID, privateKey, issuer, audience); err != nil {
		return nil, err
	}
	return &tokenIssuer{
		keyID:           keyID,
		privateKey:      privateKey,
		issuer:          issuer,
		audience:        audience,
		refreshLifetime: refreshLifetime,
		accessLifetime:  accessLifetime,
	}, nil
}

func (ti *tokenIssuer) IssueAccessUserToken(sub int64) (string, error) {
	return ti.issueAuthenticationToken(sub, TokenAccessUserRole)
}

func (ti *tokenIssuer) IssueRefreshUserToken(sub int64) (string, error) {
	return ti.issueAuthenticationToken(sub, TokenRefreshUserRole)
}

func (ti *tokenIssuer) IssueAccessSystemToken() (string, error) {
	return ti.issueAuthenticationToken(defaultSystemUserID, TokenAccessSystemRole)
}

func (ti *tokenIssuer) IssueSinglePurposeToken(descriptor SinglePurposeTokenDescriptor, sub int64, customClaims map[string]interface{}) (string, error) {
	if descriptor.IsSubMeaningful() && sub <= 0 {
		return "", xerror.New(ErrorInvalidSubject, sub)
	}
	if !descriptor.IsSubMeaningful() && sub != 0 {
		return "", xerror.New(ErrorInvalidSubject, sub)
	}
	if err := validateCustomClaims(descriptor.GetAllowedCustomClaims(), customClaims); err != nil {
		return "", err
	}

	return ti.issueLowLevelToken(sub, descriptor.GetRole(), descriptor.GetLifetime(), customClaims)
}

func (ti *tokenIssuer) issueAuthenticationToken(sub int64, role string) (string, error) {
	if role != TokenAccessUserRole && role != TokenAccessSystemRole && role != TokenRefreshUserRole {
		return "", xerror.New(ErrorInvalidRole, sub)
	}
	if (role == TokenAccessUserRole || role == TokenRefreshUserRole) && sub <= 0 {
		return "", xerror.New(ErrorInvalidSubject, sub)
	}
	if role == TokenAccessSystemRole && sub != 0 {
		return "", xerror.New(ErrorInvalidSubject, sub)
	}

	var lifetime time.Duration

	if role == TokenRefreshUserRole {
		lifetime = ti.refreshLifetime
	} else {
		lifetime = ti.accessLifetime
	}

	return ti.issueLowLevelToken(sub, role, lifetime, map[string]interface{}{})
}

func (ti *tokenIssuer) issueLowLevelToken(sub int64, role string, lifetime time.Duration, customClaims map[string]interface{}) (string, error) {
	issuedAt := time.Now()
	t := jwt.New(jwt.SigningMethodRS256)

	t.Header[keyIDHeader] = ti.keyID
	t.Claims[tokenVersionHeader] = currentTokenVersion
	t.Claims[subjectHeader] = fmt.Sprintf("%v", sub)
	t.Claims[roleHeader] = role
	t.Claims[issuerHeader] = ti.issuer
	t.Claims[audienceHeader] = ti.audience
	t.Claims[issuedAtHeader] = issuedAt.Unix()
	t.Claims[expirationHeader] = issuedAt.Add(lifetime).Unix()

	for claim, value := range customClaims {
		t.Claims[claim] = value
	}

	s, err := t.SignedString(ti.privateKey)
	if err != nil {
		return "", xerror.Wrap(err, ErrorUnableToSignToken, t)
	}
	return s, nil
}

// TokenVerifier describes the capability of verifying tokens.
type TokenVerifier interface {
	VerifyToken(t string) (int64, string, error)
	VerifySinglePurposeToken(t string, descriptor SinglePurposeTokenDescriptor) (int64, map[string]interface{}, error)
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
	dt, sub, role, err := tv.preVerify(t)
	if err != nil {
		return 0, "", err
	}

	if role != TokenAccessUserRole && role != TokenAccessSystemRole && role != TokenRefreshUserRole {
		return 0, "", xerror.New(ErrorInvalidToken, t)
	}
	if role == TokenAccessSystemRole && sub != defaultSystemUserID {
		return 0, "", xerror.New(ErrorInvalidToken, t)
	}

	if err := tv.postVerify(dt, t); err != nil {
		return 0, "", err
	}

	return sub, role, nil
}

func (tv *tokenVerifier) VerifySinglePurposeToken(t string, descriptor SinglePurposeTokenDescriptor) (int64, map[string]interface{}, error) {
	dt, sub, role, err := tv.preVerify(t)
	if err != nil {
		return 0, nil, err
	}

	if err := tv.postVerify(dt, t); err != nil {
		return 0, nil, err
	}

	if role != descriptor.GetRole() {
		return 0, nil, xerror.New(ErrorInvalidToken, t)
	}

	if descriptor.IsSubMeaningful() && sub <= 0 {
		return 0, nil, xerror.New(ErrorInvalidToken, t)
	}
	if !descriptor.IsSubMeaningful() && sub != 0 {
		return 0, nil, xerror.New(ErrorInvalidToken, t)
	}

	customClaims := make(map[string]interface{}, len(descriptor.GetAllowedCustomClaims()))
	for cN, _ := range descriptor.GetAllowedCustomClaims() {
		if cV, ok := dt.Claims[cN]; ok {
			customClaims[cN] = cV
		}
	}
	if err := validateCustomClaims(descriptor.GetAllowedCustomClaims(), customClaims); err != nil {
		return 0, nil, xerror.Wrap(err, ErrorInvalidToken, t)
	}

	if err := tv.postVerify(dt, t); err != nil {
		return 0, nil, err
	}

	return sub, customClaims, nil
}

func (tv *tokenVerifier) preVerify(t string) (*jwt.Token, int64, string, error) {
	dt, err := tv.jwtParser.Parse(t, tv.keyCallback)
	if err != nil {
		return nil, 0, "", xerror.Wrap(err, ErrorInvalidToken, t)
	}

	v, err := safeGetStringClaim(dt, tokenVersionHeader)
	if err != nil {
		return nil, 0, "", err
	}
	if v != currentTokenVersion {
		return nil, 0, "", xerror.New(ErrorInvalidToken, t)
	}

	sub, err := safeGetStringClaimAsInt64(dt, subjectHeader)
	if err != nil {
		return nil, 0, "", err
	}

	role, err := safeGetStringClaim(dt, roleHeader)
	if err != nil {
		return nil, 0, "", err
	}

	return dt, sub, role, nil
}

func (tv *tokenVerifier) postVerify(dt *jwt.Token, t string) error {
	iss, err := safeGetStringClaim(dt, issuerHeader)
	if err != nil {
		return err
	}
	if iss != tv.issuer {
		return xerror.New(ErrorInvalidToken, t)
	}

	aud, err := safeGetStringClaim(dt, audienceHeader)
	if err != nil {
		return err
	}
	if aud != tv.audience {
		return xerror.New(ErrorInvalidToken, t)
	}

	exp, err := safeGetJSONNumberClaimAsInt64(dt, expirationHeader)
	if err != nil {
		return err
	}
	if exp < time.Now().Unix() {
		return xerror.New(ErrorExpiredToken, t)
	}

	return nil
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

func validateCustomClaims(allowedCustomClaims map[string]reflect.Type, customClaims map[string]interface{}) error {
	remainingCustomClaims := make(map[string]bool, len(allowedCustomClaims))
	for acN, _ := range allowedCustomClaims {
		remainingCustomClaims[acN] = true
	}

	for cN, cV := range customClaims {
		if acT, ok := allowedCustomClaims[cN]; !ok || acT != reflect.TypeOf(cV) {
			return xerror.New(ErrorInvalidCustomClaim, cN)
		}
		delete(remainingCustomClaims, cN)
	}

	if len(remainingCustomClaims) != 0 {
		return xerror.New(ErrorMissingCustomClaims, remainingCustomClaims)
	}

	return nil
}
