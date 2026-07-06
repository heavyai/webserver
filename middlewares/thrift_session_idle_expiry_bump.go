// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// ThriftSessionIdleExpiryBump - Thrift SessionConfig Idle Expiry Bump
func ThriftSessionIdleExpiryBump(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		thriftCtx := ctx.(*models.ThriftReqContext)
		if thriftCtx.IsWhitelistedMethod() || SkipLegacyClients(ctx) {
			return next(ctx)
		}

		claims := thriftCtx.Claims

		dbSession, err := thriftCtx.Claims.SessionsInfoMap.GetDBSessionInfo(thriftCtx.DBName)
		if err != nil || dbSession == nil {
			config.Log.Error("Session does not exist for given DB context: ", thriftCtx.DBName)
			// If DB session cannot be found, just skip over timeouts bumping since there's nothing left to do
			return next(ctx)
		}

		dbSession.IdleTimeout = time.Now().Add(config.Session.IdleTimeout)

		newToken, err := models.NewAuthToken(models.AuthTokenConfig{
			SessionsInfoMap:  thriftCtx.Claims.SessionsInfoMap,
			Username:         claims.Username,
			MaxSessionExpiry: claims.MaxExpiry.Time(),
		})
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Could not create token")
			return err
		}

		util.SetAuthCookie(thriftCtx, newToken, claims.MaxExpiry.Time())
		return next(thriftCtx)
	}
}
