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

// DestroySession - Delete Auth Cookie and Destory Session Handler
// Requests session destroy in core and responds with a Set-Cookie header containing an expired cookie in the past, effectively deleting the authentication cookie
func DestroySession(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)

	sessionInfoRef, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Failed to find contextual session for DBName", err)
		return ctx.NoContent(http.StatusOK)
	}
	sessionID := sessionInfoRef.SessionID
	delete(immerseCtx.Claims.SessionsInfoMap, sessionInfoRef.DBName)

	if len(immerseCtx.Claims.SessionsInfoMap) == 0 {
		util.DeleteHTTPOnlyCookie(ctx, config.Other.CookieHeavyDBAuth)
		if config.Enable.CustomIntegrationAuth {
			util.DeleteHTTPOnlyCookie(ctx, constants.CookieModifiedServersJSON)
		}
	} else {
		token, err := models.NewAuthToken(models.AuthTokenConfig{
			SessionsInfoMap:  immerseCtx.Claims.SessionsInfoMap,
			Username:         immerseCtx.Claims.Username,
			MaxSessionExpiry: sessionInfoRef.MaxTimeout,
		})
		if err == nil {
			util.SetAuthCookie(
				ctx,
				token,
				sessionInfoRef.MaxTimeout,
			)
		} else {
			config.Log.Error("Could not create AuthToken: ", err)
			ctx.String(http.StatusInternalServerError, "Could not create AuthToken")
			return err
		}
	}

	err = util.DestroyDBSession(sessionID)
	if err != nil {
		config.Log.Error("Could not destroy HeavyDB session ID: ", sessionID, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}
