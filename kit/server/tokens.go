package server

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
)

// MustInitTokenIssuer initializes a new token issuer from conifg, or panics.
func MustInitTokenIssuer(cfg *TokenIssuerConfig) utils.TokenIssuer {
	tokenIssuer, err := utils.NewTokenIssuer(
		cfg.JWTKeyID, cfg.JWTKeyPrivate, cfg.JWTIssuer, cfg.JWTAudience,
		utils.DefaultRefreshTokenLifetime, utils.DefaultAccessTokenLifetime)
	if err != nil {
		panic(err)
	}
	return tokenIssuer
}

// MustInitTokenVerifier initializes a new token verifier from config, or panics.
func MustInitTokenVerifier(cfg *TokenVerifierConfig) utils.TokenVerifier {
	tokenVerifier, err := utils.NewTokenVerifier(cfg.JWTKeyID, cfg.JWTKeyPublic, cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		panic(err)
	}
	return tokenVerifier
}
