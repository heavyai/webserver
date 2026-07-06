// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"errors"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
)

// AuthToken - AuthToken type
type AuthToken string

// AuthTokenConfig - AuthTokenConfig type
type AuthTokenConfig struct {
	SessionsInfoMap  SessionsInfoMap
	Username         string
	MaxSessionExpiry time.Time
}

// NewAuthToken - NewAuthToken helper
func NewAuthToken(tokenConfig AuthTokenConfig) (newToken AuthToken, err error) {
	maxSessionExpiry := tokenConfig.MaxSessionExpiry
	username := tokenConfig.Username
	var signerOpts = jose.SignerOptions{}
	signerOpts.WithType("JWT")
	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.PS256,
			Key:       config.Session.PrivateKey,
		},
		&signerOpts,
	)
	if err != nil {
		return newToken, err
	}

	builder := jwt.Signed(signer)

	jwtClaims := ImmerseJWTClaims{
		Claims: &jwt.Claims{
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
			Expiry:   jwt.NewNumericDate(maxSessionExpiry.UTC()),
		},
		SessionsInfoMap: tokenConfig.SessionsInfoMap,
		Username:        username,
		MaxExpiry:       *jwt.NewNumericDate(maxSessionExpiry.UTC()),
	}

	rawJWT, err := builder.Claims(jwtClaims).CompactSerialize()
	if err != nil {
		return newToken, err
	}

	publicKey := &config.Session.PrivateKey.PublicKey
	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{Algorithm: jose.RSA_OAEP, Key: publicKey},
		nil,
	)
	if err != nil {
		return newToken, err
	}

	object, err := encrypter.Encrypt([]byte(rawJWT))
	if err != nil {
		return newToken, err
	}

	jwtString, err := object.CompactSerialize()
	asserted := AuthToken(jwtString)
	newToken = asserted
	return newToken, err
}

// Validate - AuthToken Validation method
func (token AuthToken) Validate(ctx echo.Context) (claims *ImmerseJWTClaims, sessionInfoRef *DBSessionInfo, validationErr error) {
	object, err := jose.ParseEncrypted(string(token))
	if err != nil {
		return nil, nil, errors.New("Not Authenticated: Not able to parse token")
	}

	decryptedByteArray, err := object.Decrypt(config.Session.PrivateKey)
	if err != nil {
		return nil, nil, errors.New("Not Authenticated: Could not decrypt token")
	}

	decryptedJWT := string(decryptedByteArray)
	parsedJWT, err := jwt.ParseSigned(decryptedJWT)
	if err != nil {
		return nil, nil, errors.New("Not Authenticated: token signature invalid")
	}

	claims = &ImmerseJWTClaims{}
	err = parsedJWT.Claims(&config.Session.PrivateKey.PublicKey, claims)
	if err != nil {
		return nil, nil, errors.New("Not Authenticated: Could not parse token claims")
	}

	var immerseCtx *ImmerseReqContext
	var ok bool
	immerseCtx, ok = ctx.(*ImmerseReqContext)
	if !ok {
		thriftCtx := ctx.(*ThriftReqContext)
		immerseCtx = thriftCtx.ImmerseReqContext
	}

	dbSessionInfo, validationErr := claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if validationErr != nil {
		return
	}

	if dbSessionInfo.IsExpired() {
		return claims, dbSessionInfo, errors.New("Not Authenticated: Auth token expired")
	}

	claims.SessionsInfoMap.MergeMap(immerseCtx.Claims.SessionsInfoMap)
	immerseCtx.Claims = claims
	return claims, dbSessionInfo, nil
}
