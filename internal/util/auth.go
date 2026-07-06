// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"net/http"
	"time"

	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/models"
)

// DeleteDocumentCookie - Delete Document Cookie utility function
func DeleteDocumentCookie(ctx echo.Context, name string) {
	ctx.SetCookie(&http.Cookie{
		Name:    name,
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})
}

// DeleteHTTPOnlyCookie - Delete HTTP Only Cookie utility function
func DeleteHTTPOnlyCookie(ctx echo.Context, name string) {
	ctx.SetCookie(&http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: true,
	})
}

// SetLogoutCookie - Set Logout Cookie utility function
func SetLogoutCookie(ctx echo.Context) {
	ctx.SetCookie(&http.Cookie{
		Name:    constants.CookieLogout,
		Value:   "logout-initiated",
		Expires: time.Now().Add(config.Session.MaxTimeout),
		Path:    "/",
	})
}

// SetAuthCookie - Set Auth Cookie utility function
func SetAuthCookie(ctx echo.Context, token models.AuthToken, expiry time.Time) {
	authCookie := new(http.Cookie)
	authCookie.Name = config.Other.CookieHeavyDBAuth
	authCookie.Value = string(token)
	authCookie.Expires = expiry
	authCookie.HttpOnly = true
	authCookie.Path = "/"

	if config.Enable.CrossDomainAuth {
		authCookie.Secure = true
		authCookie.SameSite = http.SameSiteNoneMode
	}

	if config.Enable.UltraSecureMode {
		authCookie.Secure = true
		authCookie.SameSite = http.SameSiteStrictMode
	}

	ctx.SetCookie(authCookie)
}

// GetTokenFromAuthCookie - GetTokenFromAuthCookie util
func GetTokenFromAuthCookie(ctx echo.Context) (token models.AuthToken, err error) {
	authCookie, err := ctx.Cookie(config.Other.CookieHeavyDBAuth)
	if err != nil || authCookie.Value == "" {
		var errorMsg string
		if err != nil {
			errorMsg = err.Error()
		}
		return token, errors.New("Error reading AuthToken from cookie: " + errorMsg)
	}
	return models.AuthToken(authCookie.Value), err
}

// GetTokenClaims - GetTokenClaims method
func GetTokenClaims(ctx echo.Context) (claims *models.ImmerseJWTClaims) {
	token, _ := GetTokenFromAuthCookie(ctx)
	if token == "" {
		return
	}
	immerseCtx := ctx.(*models.ImmerseReqContext)
	claimsPtr, _, err := token.Validate(immerseCtx)
	if err != nil || claimsPtr == nil {
		return immerseCtx.Claims
	}
	return claimsPtr
}

// GetSessionIDFromCtx - util
func GetSessionIDFromCtx(ctx echo.Context) (sessionID heavy.TSessionId, err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	dbSession, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return
	}
	sessionID = dbSession.SessionID
	return
}
