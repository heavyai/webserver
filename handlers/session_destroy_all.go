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

// DestroyAllSessions - Delete all sessions
func DestroyAllSessions(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)

	util.DeleteHTTPOnlyCookie(ctx, config.Other.CookieHeavyDBAuth)
	if config.Enable.CustomIntegrationAuth {
		util.DeleteHTTPOnlyCookie(ctx, constants.CookieModifiedServersJSON)
	}

	for dbName := range immerseCtx.Claims.SessionsInfoMap {
		sessionID := immerseCtx.Claims.SessionsInfoMap[dbName].SessionID
		err = util.DestroyDBSession(sessionID)
		if err != nil {
			config.Log.Error("Could not destroy DB session ID: ", sessionID, err)
		}
	}
	util.SetLogoutCookie(ctx)

	return ctx.NoContent(http.StatusNoContent)
}
