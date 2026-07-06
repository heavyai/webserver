// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

type exchangeSessionReqBody struct {
	DBName string `json:"dbName"`
}

// ExchangeSession - ExchangeSession handler
func ExchangeSession(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	var sessionID heavy.TSessionId
	body := exchangeSessionReqBody{}
	if err = ctx.Bind(&body); err != nil || body.DBName == "" {
		config.Log.Error(err)
		ctx.NoContent(http.StatusBadRequest)
		return err
	}

	var thriftSessionInfo *heavy.TSessionInfo
	if immerseCtx.Claims.SessionsInfoMap[body.DBName] != nil && !immerseCtx.Claims.SessionsInfoMap[body.DBName].IsExpired() {
		thriftSessionInfo, err = util.GetThriftSessionInfo(immerseCtx.Claims.SessionsInfoMap[body.DBName].SessionID)
		if err != nil {
			ctx.NoContent(http.StatusInternalServerError)
			return err
		}
		return ctx.JSON(http.StatusOK, thriftSessionInfo)
	}

	sessionInfoRef := immerseCtx.Claims.SessionsInfoMap.GetSingleSessionInfo()
	if sessionID, thriftSessionInfo, err = util.ExchangeSessionID(sessionInfoRef.SessionID, body.DBName); err != nil {
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	// TODO: Either fetch session info from thrift to retrieve timeout values, or else mitigate out of sync timeouts
	//    by taking into account session creation latency to calculate session timeouts here.
	//    https://heavyai.atlassian.net/browse/FE-10542
	idleSessionExpiry := time.Now().Add(config.Session.IdleTimeout)
	maxSessionExpiry := immerseCtx.Claims.MaxExpiry.Time()

	token, err := immerseCtx.Claims.AddSessionToToken(sessionInfoRef.Username, sessionID, body.DBName, idleSessionExpiry, maxSessionExpiry)
	if err != nil {
		config.Log.Error(err)
		ctx.String(http.StatusInternalServerError, "Could not create auth token")
		return err
	}

	util.SetAuthCookie(
		ctx,
		token,
		maxSessionExpiry,
	)
	return ctx.JSON(http.StatusOK, *thriftSessionInfo)
}
