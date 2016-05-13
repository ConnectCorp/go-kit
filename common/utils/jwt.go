package utils

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"strconv"
	"time"
)

const (
	ErrorInvalidToken      = "invalid token"
	ErrorExpiredToken      = "expired token"
	ErrorUnableToSignToken = "unable to sign token"
)

const (
	TokenUserRole   = "user"
	TokenSystemRole = "system"
)

const (
	tokenKeyID    = "k1"
	tokenLifeTime = time.Hour * 24 * 365 // One year.
	tokenVersion  = "v1"
)

var (
	jwtParser = &jwt.Parser{UseJSONNumber: true}
)

func VerifyToken(token string, jwtPublicKey []byte) (int64, string, error) {
	decodedToken, err := jwtParser.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok || token.Header["kid"].(string) != tokenKeyID {
			return nil, xerror.New(ErrorInvalidToken, token)
		}
		return jwtPublicKey, nil
	})
	if err != nil {
		return 0, "", xerror.Wrap(err, ErrorInvalidToken, token)
	}

	if decodedToken.Claims["v"].(string) != tokenVersion {
		return 0, "", xerror.Wrap(err, ErrorInvalidToken, token)
	}

	userID, err := strconv.ParseInt(decodedToken.Claims["sub"].(string), 10, 64)
	if err != nil {
		return 0, "", xerror.Wrap(err, ErrorInvalidToken, token)
	}

	role := decodedToken.Claims["role"].(string)
	if role != TokenUserRole && role != TokenSystemRole {
		return 0, "", xerror.Wrap(err, ErrorInvalidToken, token)
	}

	expV, err := decodedToken.Claims["exp"].(json.Number).Int64()
	if err != nil {
		return 0, "", xerror.Wrap(err, ErrorInvalidToken, token)
	}
	if time.Unix(expV, 0).Before(time.Now()) {
		return 0, "", xerror.New(ErrorExpiredToken, token)
	}

	return userID, role, nil
}

func IssueToken(userID int64, key []byte) (string, error) {
	return IssueCustomToken(userID, TokenUserRole, time.Now().Add(tokenLifeTime), key)
}

func IssueCustomToken(userID int64, role string, expirationTime time.Time, key []byte) (string, error) {
	t := jwt.New(jwt.SigningMethodRS256)
	t.Header["kid"] = tokenKeyID
	t.Claims["v"] = tokenVersion
	t.Claims["sub"] = fmt.Sprintf("%v", userID)
	t.Claims["role"] = role
	t.Claims["exp"] = expirationTime.Unix()
	s, err := t.SignedString(key)
	if err != nil {
		return "", xerror.Wrap(err, ErrorUnableToSignToken, t)
	}
	return s, nil
}
