// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// SessionValidationResponse -- SessionValidationResponse type
type SessionValidationResponse struct {
	Username string `json:"username"`
	DBName   string `json:"dbName"`
}

// ValidateSession -- ValidateSession handler
func ValidateSession(ctx echo.Context) error {
	authCookie, err := ctx.Cookie(config.Other.CookieHeavyDBAuth)
	if err != nil || authCookie.Value == "" {
		ctx.String(http.StatusUnauthorized, "Not Authenticated: no auth token cookie read")
		return err
	}
	token := models.AuthToken(authCookie.Value)
	claims, sessionInfoRef, tokenValidationError := token.Validate(ctx)
	if tokenValidationError != nil {
		if config.Enable.CustomIntegrationAuth {
			util.DeleteHTTPOnlyCookie(ctx, constants.CookieModifiedServersJSON)
		}
		// If we got back claims, we need to remove the SessionInfo that didn't validate and remove the Auth cookie if
		//   SessionInfoMap is now empty.  Otherwise, just delete the bad Auth cookie.
		if claims != nil {
			immerseCtx := ctx.(*models.ImmerseReqContext)
			claims.SessionsInfoMap.DeleteSessionInfo(immerseCtx.DBName)
			if immerseCtx.DBName == "" || len(claims.SessionsInfoMap) == 0 {
				util.DeleteHTTPOnlyCookie(ctx, config.Other.CookieHeavyDBAuth)
				util.SetLogoutCookie(ctx)
			}
		} else {
			util.DeleteHTTPOnlyCookie(ctx, config.Other.CookieHeavyDBAuth)
			util.SetLogoutCookie(ctx)
		}

		ctx.String(http.StatusUnauthorized, tokenValidationError.Error())
		return tokenValidationError
	}
	return ctx.JSON(http.StatusOK, SessionValidationResponse{Username: claims.Username, DBName: sessionInfoRef.DBName})
}
