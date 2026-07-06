// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
)

// AppIndexSessionExchange - AppIndexSessionExchange handler
func AppIndexSessionExchange(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		DBName := util.URLPathExtractDBContext(ctx.Request().URL.Path)
		if DBName == "" {
			return next(ctx)
		}

		claims := util.GetTokenClaims(ctx)
		if claims == nil || len(claims.SessionsInfoMap) == 0 {
			return next(ctx)
		}

		if claims.SessionsInfoMap[DBName] != nil && !claims.SessionsInfoMap[DBName].IsExpired() {
			return next(ctx)
		}

		singleSessionInfo := claims.SessionsInfoMap.GetSingleSessionInfo()
		if util.InaccessibleDB(singleSessionInfo, DBName) {
			config.Log.Error("Inaccessible or nonexistent database: " + DBName)
			return ctx.Redirect(http.StatusSeeOther, "/")
		}

		newSessionID, _, err := util.ExchangeSessionID(singleSessionInfo.SessionID, DBName)
		if err != nil {
			ctx.Logger().Error(err)
			ctx.NoContent(http.StatusInternalServerError)
			return err
		}
		// TODO: Either fetch session info from thrift to retrieve timeout values, or else mitigate out of sync timeouts
		//    by taking into account session creation latency to calculate session timeouts here.
		//    https://heavyai.atlassian.net/browse/FE-10542
		idleSessionExpiry := time.Now().Add(config.Session.IdleTimeout)
		maxSessionExpiry := claims.MaxExpiry.Time()

		token, err := claims.AddSessionToToken(singleSessionInfo.Username, newSessionID, DBName, idleSessionExpiry, maxSessionExpiry)
		if err != nil {
			ctx.Logger().Error(err)
			ctx.String(http.StatusInternalServerError, "Could not create auth token")
			return err
		}

		util.SetAuthCookie(
			ctx,
			token,
			claims.MaxExpiry.Time(),
		)
		return next(ctx)
	}
}
