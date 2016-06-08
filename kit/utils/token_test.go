package utils

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"testing"
	"time"
	"strings"
)

const (
	keyID    = "k1"
	issuer   = "https://test-issuer"
	audience = "connect-test"
)

var privateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAl2vpHuMDO8uycGDXv/J7bJudjaiow27bwHU9SH18cClFbAbl
yEGDPeFHKAbxTmIRVN+NMSGEdYlYVG7k308uDXJ0/24tHPKVcbG7XXCLNIBQssJa
06sA6xx76JhHG1uwBFtS7lOXTMANStnwfk9yFDmc27qU0TA3Cu/aYIRgmOP/EDE4
mL73U0CDHfv5O/gv3GRdVyn5Mdf0VFIFHhTlTgjSdzoVIhv4QD5ctv4YiFutuF23
CVlMVhvJi5upen6XgvQKNM8supiY1pnyxOLk3XGsM/X36Xy6eu+/Izvqp26bMW1A
rTTN73a9dWR0bmJTPLkp6P4WPoIcDHF9QyX81wIDAQABAoIBAAfoo/lwA/g4kG5c
MIifAqFOk3EVsRLcFHA972B85WX6Udztk0zdVxyBSrUlf68HYj5bmsVJKeD1tn5a
eGvNd+tN4hyBRavwY6vXi6C6wxqP5SchDZtmoBqnlzUz1urv5AamOnOmPA3PLiKN
tYjzX1L9G1tCqIkwin9wvagy4dS6Z0NfcRCpRv1oyy3E6JBIM+KF7dUOFgf/En6s
hMruckbX6nXT9pcZKaoRL4XI58X/dVf9TMq2nsA6riGxwbJOJ2WiSiRIEJPEmBA6
SQCDAX6lMZFHqfzRR6ogUkHzITwvq7nixu4hwchAepn7SI6ZRoiirFT50RvzFRh9
CV20XWkCgYEAx7+vKDIQ/NcS5pBNobx1Ay1cmGOROBiCrdT4kqe2ARUPQMYSXDHR
TtJbH3rd1oMXLq9obwTMQphS/UdPiF3yieD2WmRK0C8ZU9yhNnwiMmzvU0xlvAZV
McZyHVupJ9CbYa6MdAuYzqfdIGe1InkanoWU0RkPy/yhkPzrXCTyXFMCgYEAwhA2
WH09csSW3oygoXMKTekBT5ws5zwaFMhEUmtb4mftbkL14dfauL5e+DUjgb9SwVaw
kTwDUXl4csrGYPnbRrZIMU4SK6VcbkXbBwNGHmKMDcgikw3pzDVaFC9s1GQIQYv2
8024bFH54HpcVvpNrddvXKcV2L8O6JP2xXJu7O0CgYAO53SQUTwHQZz9ayL/wGoS
tJ3GGRfK0blecxehCbaA2itrL9xK2MS/Vt7JuIc47EschqYKMpdzGJ6Im3uJt0jT
lN+M2xLh+cGwCjRVNmnuzUYGNxsYLnjI3/+/xQkYGW6emUGNnxflw4yyUEqpqdOc
pGb4OyB8nfsIMHb3RyJ2VQKBgBNnfZehhjhokdFU7GbYUupxZvEn45GHf/AeCj7X
f0uHKsWAqodXhwY7+tEEtzUtBUBRw7vx7T8DT1jjD6z4rsVGSrerX8O/eBuKnpj3
6dX18p0aKuLbXEpP917XUyF1kyHCtgGj/tHN7JdWhM8pngTI6tiv2E5g5EO7L8yU
YaUNAoGAUoXwPbgwJZMTzFjeQFRiC64VaarxIxkBHRodGLRrmdRSVDg/ULKWyDeZ
83exUsNiMXRYRFMystBjxcjHssCZWgCngr4O8MH2sVHdsa4kOEvIiI3X/v4f1WS9
7s88gcPeoKnZAFRDn+wo0ezxqF2DGrH024R6FWyLRwXh7BA+qN8=
-----END RSA PRIVATE KEY-----`)

var otherPrivateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAtSuEQoCZz+zrNcQHi8iU4tG5szsjLIVggJOHfXZHIdgR3Nnn
mPX/oH274iVPaDkWnfPq8MjndqzQD9/iFUbuNtjZdXsDL/aglRINuyI35mUHpv/9
Tr+XgsVuF5935I7iOUDa3fsOl5bcxfriTEPRhY8c1Bh7ip4hrDscshQhj4whEbPm
1TeQe0V7Hk7xhyQRZR8eO4YN9iz8JZtmVFUbQHtnt01qBsoKQnhH5dMqeMDtWmAZ
iILTTtKlzZwdgIzuOnC9JKxDjnDtPWtTLtmWpBImoYREoKhaIupe8ONOv5i+PVzp
ng3JzzqN4VCXJxbisafIGu5Psn1qF8Q9dhmO+QIDAQABAoIBAAOlb5EDE5VquEUp
v9khXVW/UNR5oTiZKpsL3RM6WS4mMh3WyOF8OfvZ6/keNR3s4DannRfhgz8RNVLM
d7lj5xF1hdLmeebyOMMnTN1yT9u0NtY2mruGAZ+hJ5kUHY6dDZKHaqBDGEUWxTS6
2ukcCFB+096062+bxSO5QNeYriA8chBL/pheRYOFBTDUSe4qON3LA6lNs06VzJSE
0/Ft0OwCmMGgj+M/QsofaXZYMcqCzDfHTLuhg069Fw7KtX81gLy+b46loR4EFuEq
B09/GdvZ/yiOyIMU+Nit7uTL96GaOKOmLgQkILWu9rzDoVxq3ngg3vTZXW6HLg0l
B5qyO60CgYEA6GTSfEF3jZGYbE/2/4N4dWOaYnd7O7RsQTws3EaJ0vdWBQW/Sugk
P+7fMAg4nTLaZgPQC8cA8CmRiiAOig3LAej75MjjY9JkNLs/4CUc5qu1FnSLXocq
cFWwV+pKWs6brkczE6vbNt2pBaq3woCbwD96qDuy1ARznYAzTkagQ48CgYEAx5Kq
5iPJvRUMN2Y4Wx8BxxcV0mmHHLsJSU8/ew8zDO3TI2DA301kw6c1zTCLlLScDVcL
ZLhWbIKIYJn+3cyenBzTToqF9bw0I/uvh7QQqGj2Mgh+zg8cfu184ELKjlLWMrpr
gHxb8YAXXrtH31DYH7qy1WWr/63raMTFxklRoPcCgYBxs8qsUteso1zBOcqur2OD
g+0oWi8oQhlpPYjxaW3Lk4o5wNscSkJaKYR3mr4gY54ppZnn+UEDQENeIlsavq7h
y11bTdK7p1ex2R/iiiX+0moyh2kdIeLovXQfP5mLnmTbOyjJah9CU+d7x1BLUONj
h2t63mKbi2YJ3Iy9sp59DwKBgDtLyJs4ZuhXKJoNNRFd1RliMomh8RMIP2oYsbPO
gEyHHQSV6rhuNlIrjEC6+73jK7qK8keqvYLgBcUt/BvKgBXCOsZLQiIRGSzXyv92
8LwY841KGOMAemb8CO5Y6fX/hsTrvqUeTfMjK85ptqETVCOZRSlCXChLdHZcgKa5
ghdhAoGAB33pjv4i8gM8yanGz/rDm2xoR6tE1ZNdn46Q7KKCBwt6r6PN/Pvu4N8+
AwWo0m+/+8oUOFvWgP92Ropdfr0yI0sPHZPLq6o5cdoYoY9Ww65+7+ldIYl5clgz
MMc6OXlBAEMZ0B6o/2pvDLXeGlaPMdnkVXjwNJCOO1LeDZkvdLc=
-----END RSA PRIVATE KEY-----`)

var publicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAl2vpHuMDO8uycGDXv/J7
bJudjaiow27bwHU9SH18cClFbAblyEGDPeFHKAbxTmIRVN+NMSGEdYlYVG7k308u
DXJ0/24tHPKVcbG7XXCLNIBQssJa06sA6xx76JhHG1uwBFtS7lOXTMANStnwfk9y
FDmc27qU0TA3Cu/aYIRgmOP/EDE4mL73U0CDHfv5O/gv3GRdVyn5Mdf0VFIFHhTl
TgjSdzoVIhv4QD5ctv4YiFutuF23CVlMVhvJi5upen6XgvQKNM8supiY1pnyxOLk
3XGsM/X36Xy6eu+/Izvqp26bMW1ArTTN73a9dWR0bmJTPLkp6P4WPoIcDHF9QyX8
1wIDAQAB
-----END PUBLIC KEY-----`)

func TestIssue(t *testing.T) {
	ti, err := NewTokenIssuer(keyID, privateKey, issuer, audience, DefaultTokenLifetime)
	assert.Nil(t, err)

	tv, err := NewTokenVerifier(keyID, publicKey, issuer, audience)
	assert.Nil(t, err)

	_, err = ti.IssueUserToken(0)
	assert.Equal(t, "invalid subject", err.Error())

	ut, err := ti.IssueUserToken(1)
	assert.Nil(t, err)

	sub, role, err := tv.VerifyToken(ut)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, sub)
	assert.Equal(t, tokenUserRole, role)

	st, err := ti.IssueSystemToken()
	assert.Nil(t, err)

	sub, role, err = tv.VerifyToken(st)
	assert.Nil(t, err)
	assert.EqualValues(t, defaultSystemUserID, sub)
	assert.Equal(t, tokenSystemRole, role)
}

func TestVerify(t *testing.T) {
	tv, err := NewTokenVerifier(keyID, publicKey, issuer, audience)
	assert.Nil(t, err)

	goodToken, err := issueTestToken(time.Now(), keyID, currentTokenVersion, "1", tokenUserRole, issuer, audience, DefaultTokenLifetime, privateKey)
	assert.Nil(t, err)
	sub, role, err := tv.VerifyToken(goodToken)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, sub)
	assert.Equal(t, tokenUserRole, role)

	_, _, err = tv.VerifyToken("")
	assert.Equal(t, "invalid token: token contains an invalid number of segments", err.Error())

	_, _, err = tv.VerifyToken("bad")
	assert.Equal(t, "invalid token: token contains an invalid number of segments", err.Error())

	badKeyID, err := issueTestToken(time.Now(), "k2", currentTokenVersion, "1", tokenUserRole, issuer, audience, DefaultTokenLifetime, privateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(badKeyID)
	assert.Equal(t, "invalid token: invalid token header", err.Error())

	badKey, err := issueTestToken(time.Now(), keyID, currentTokenVersion, "1", tokenUserRole, issuer, audience, DefaultTokenLifetime, otherPrivateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(badKey)
	assert.Equal(t, "invalid token: crypto/rsa: verification error", err.Error())

	badTokenVersion, err := issueTestToken(time.Now(), keyID, "v2", "1", tokenUserRole, issuer, audience, DefaultTokenLifetime, privateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(badTokenVersion)
	assert.Equal(t, "invalid token", err.Error())

	badSub, err := issueTestToken(time.Now(), keyID, currentTokenVersion, "a", tokenUserRole, issuer, audience, DefaultTokenLifetime, privateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(badSub)
	assert.Equal(t, "invalid token: strconv.ParseInt: parsing \"a\": invalid syntax", err.Error())

	badSub, err = issueTestToken(time.Now(), keyID, currentTokenVersion, "1.1", tokenUserRole, issuer, audience, DefaultTokenLifetime, privateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(badSub)
	assert.Equal(t, "invalid token: strconv.ParseInt: parsing \"1.1\": invalid syntax", err.Error())

	badRole, err := issueTestToken(time.Now(), keyID, currentTokenVersion, "1", "bad", issuer, audience, DefaultTokenLifetime, privateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(badRole)
	assert.Equal(t, "invalid token", err.Error())

	badIssuer, err := issueTestToken(time.Now(), keyID, currentTokenVersion, "1", tokenUserRole, "bad", audience, DefaultTokenLifetime, privateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(badIssuer)
	assert.Equal(t, "invalid token", err.Error())

	badAudience, err := issueTestToken(time.Now(), keyID, currentTokenVersion, "1", tokenUserRole, issuer, "bad", DefaultTokenLifetime, privateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(badAudience)
	assert.Equal(t, "invalid token", err.Error())

	badSubForSystem, err := issueTestToken(time.Now(), keyID, currentTokenVersion, "1", tokenSystemRole, issuer, audience, DefaultTokenLifetime, privateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(badSubForSystem)
	assert.Equal(t, "invalid token", err.Error())

	expired, err := issueTestToken(time.Now().Add(-time.Hour), keyID, currentTokenVersion, "1", tokenUserRole, issuer, audience, time.Minute, privateKey)
	assert.Nil(t, err)
	_, _, err = tv.VerifyToken(expired)
	assert.True(t, strings.HasPrefix(err.Error(),  "invalid token: token is expired"))
}

func TestIssuerInit(t *testing.T) {
	_, err := NewTokenIssuer("bad", privateKey, issuer, audience, DefaultTokenLifetime)
	assert.Equal(t, "invalid key ID: bad", err.Error())

	_, err = NewTokenIssuer(keyID, []byte{}, issuer, audience, DefaultTokenLifetime)
	assert.Equal(t, "invalid key material", err.Error())

	_, err = NewTokenIssuer(keyID, privateKey, "bad", audience, DefaultTokenLifetime)
	assert.Equal(t, "invalid issuer: bad", err.Error())

	_, err = NewTokenIssuer(keyID, privateKey, issuer, "bad", DefaultTokenLifetime)
	assert.Equal(t, "invalid audience: bad", err.Error())
}

func TestVerifierInit(t *testing.T) {
	_, err := NewTokenVerifier("bad", publicKey, issuer, audience)
	assert.Equal(t, "invalid key ID: bad", err.Error())

	_, err = NewTokenVerifier(keyID, []byte{}, issuer, audience)
	assert.Equal(t, "invalid key material", err.Error())

	_, err = NewTokenVerifier(keyID, publicKey, "bad", audience)
	assert.Equal(t, "invalid issuer: bad", err.Error())

	_, err = NewTokenVerifier(keyID, publicKey, issuer, "bad")
	assert.Equal(t, "invalid audience: bad", err.Error())
}

func issueTestToken(issuedAt time.Time, keyID, tokenVersion, sub, role, issuer, audience string, lifetime time.Duration, privateKey []byte) (string, error) {
	t := jwt.New(jwt.SigningMethodRS256)

	t.Header[keyIDHeader] = keyID
	t.Claims[tokenVersionHeader] = tokenVersion
	t.Claims[subjectHeader] = sub
	t.Claims[roleHeader] = role
	t.Claims[issuerHeader] = issuer
	t.Claims[audienceHeader] = audience
	t.Claims[issuedAtHeader] = issuedAt.Unix()
	t.Claims[expirationHeader] = issuedAt.Add(lifetime).Unix()

	s, err := t.SignedString(privateKey)
	if err != nil {
		return "", xerror.Wrap(err, ErrorUnableToSignToken, t)
	}
	return s, nil
}
