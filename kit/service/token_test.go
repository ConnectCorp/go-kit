package service

import (
	"github.com/ConnectCorp/go-kit/kit/test"
	"github.com/ConnectCorp/go-kit/kit/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
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

func TestToken(t *testing.T) {
	ti, err := utils.NewTokenIssuer(keyID, privateKey, issuer, audience, utils.DefaultRefreshTokenLifetime, utils.DefaultAccessTokenLifetime)
	assert.Nil(t, err)
	token, err := ti.IssueAccessUserToken(1)
	assert.Nil(t, err)

	req := test.MustNewRequest()
	req.Header.Set(authorizationHeader, "Bearer "+token)

	tv, err := utils.NewTokenVerifier(keyID, publicKey, issuer, audience)
	assert.Nil(t, err)

	ctx := TokenExtractor(context.Background(), req)
	assert.Equal(t, "Bearer "+token, ctxToken(ctx))
	assert.Equal(t, "", ctxToken(TokenExtractor(context.Background(), test.MustNewRequest())))

	tokenMiddleware := NewTokenMiddleware(tv)
	tokenFunc := tokenMiddleware(test.TerminationMiddleware)
	_, err = tokenFunc(ctx, req)
	assert.Equal(t, "terminated", err.Error())
	_, err = tokenFunc(ctxWithToken(context.Background(), "bad"), req)
	assert.Equal(t, "unauthorized: invalid token: token contains an invalid number of segments", err.Error())
	_, err = tokenFunc(context.Background(), req)
	assert.Equal(t, "unauthorized: missing token", err.Error())

	noTokenMiddleware := NewNoTokenMiddleware()
	noTokenFunc := noTokenMiddleware(test.TerminationMiddleware)
	_, err = noTokenFunc(ctx, req)
	assert.Equal(t, "bad request: must not authenticate", err.Error())
	_, err = noTokenFunc(context.Background(), req)
	assert.Equal(t, "terminated", err.Error())

}

func TestAuthVerifier(t *testing.T) {
	systemTokenCtx := ctxWithAuthorizedRole(context.Background(), utils.TokenAccessSystemRole)
	userTokenCtx1 := ctxWithAuthorizedSub(ctxWithAuthorizedRole(context.Background(), utils.TokenAccessUserRole), 1)
	userTokenCtx2 := ctxWithAuthorizedSub(ctxWithAuthorizedRole(context.Background(), utils.TokenAccessUserRole), 2)
	userTokenCtx3 := ctxWithAuthorizedSub(ctxWithAuthorizedRole(context.Background(), utils.TokenAccessUserRole), 3)

	assert.NotNil(t, NewContextAuthVerifier(systemTokenCtx).Verify())
	assert.NotNil(t, NewContextAuthVerifier(userTokenCtx1).Verify())

	assert.Nil(t, NewContextAuthVerifier(systemTokenCtx).AcceptAccessSystemToken().Verify())
	assert.NotNil(t, NewContextAuthVerifier(userTokenCtx1).AcceptAccessSystemToken().Verify())

	assert.NotNil(t, NewContextAuthVerifier(systemTokenCtx).AcceptAnyAccessUserToken().Verify())
	assert.Nil(t, NewContextAuthVerifier(userTokenCtx1).AcceptAnyAccessUserToken().Verify())

	assert.NotNil(t, NewContextAuthVerifier(systemTokenCtx).AcceptAccessUserTokenForSubs(1).Verify())
	assert.Nil(t, NewContextAuthVerifier(userTokenCtx1).AcceptAccessUserTokenForSubs(1).Verify())
	assert.NotNil(t, NewContextAuthVerifier(userTokenCtx2).AcceptAccessUserTokenForSubs(1).Verify())

	assert.NotNil(t, NewContextAuthVerifier(systemTokenCtx).AcceptAccessUserTokenForSubs(1, 2).Verify())
	assert.Nil(t, NewContextAuthVerifier(userTokenCtx1).AcceptAccessUserTokenForSubs(1, 2).Verify())
	assert.Nil(t, NewContextAuthVerifier(userTokenCtx2).AcceptAccessUserTokenForSubs(1, 2).Verify())
	assert.NotNil(t, NewContextAuthVerifier(userTokenCtx3).AcceptAccessUserTokenForSubs(1, 2).Verify())
}
