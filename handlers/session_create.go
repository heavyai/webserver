// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// CreateSession - CreateSession Handler
// Given a valid application/JSON body w/ valid authentication credentials for
// underlying HeavyDB instance, this endpoint will create an HeavyDB session,
// create a JWT/S/E w/ the sessionID from HeavyDB, and place a Set-Cookie
// header w/ the JWT as it's value, on the the response.
func CreateSession(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	creds := new(models.AuthCredentials)

	if err = ctx.Bind(creds); err != nil {
		ctx.NoContent(http.StatusBadRequest)
		return err
	}

	sessionID, err := util.CreateDBSession(*creds)
	if err != nil {
		config.Log.Error("Failed to create HeavyDB session: ", err)

		errStr := util.ErrorMsg(err)
		if errStr == "Authentication failure" {
			errStr = "Invalid Credentials"
		}

		ctx.String(http.StatusUnauthorized, errStr)
		return err
	}

	tokenClaims := util.GetTokenClaims(ctx)
	if tokenClaims != nil {
		immerseCtx.Claims = tokenClaims
	}

	// TODO: Either fetch session info from thrift to retrieve timeout values, or else mitigate out of sync timeouts
	//    by taking into account session creation latency to calculate session timeouts here.
	//    https://heavyai.atlassian.net/browse/FE-10542
	idleSessionExpiry := time.Now().Add(config.Session.IdleTimeout)

	var maxSessionExpiry time.Time
	if immerseCtx.Claims.MaxExpiry == 0 {
		maxSessionExpiry = time.Now().Add(config.Session.MaxTimeout)
	} else {
		maxSessionExpiry = immerseCtx.Claims.MaxExpiry.Time()
	}

	immerseCtx.Claims.Username = creds.Username
	token, err := immerseCtx.Claims.AddSessionToToken(creds.Username, sessionID, creds.DBName, idleSessionExpiry, maxSessionExpiry)
	if err != nil {
		config.Log.Error(err)
		ctx.String(http.StatusInternalServerError, "Could not create auth token")
		return err
	}

	sessionInfo, err := util.GetDBSessionInfo(immerseCtx.Claims, creds.DBName)
	if err != nil {
		config.Log.Error("Failed to retrieve HeavyDB sessionInfo: ", err)
		ctx.NoContent(http.StatusInternalServerError)
		destroyErr := util.DestroyDBSession(sessionID)
		if destroyErr != nil {
			config.Log.Error("Could not destroy HeavyDB session ID: ", sessionID, destroyErr)
		}
		return err
	}

	util.SetAuthCookie(
		ctx,
		token,
		maxSessionExpiry,
	)
	util.DeleteDocumentCookie(ctx, constants.CookieLogout)

	return ctx.JSON(http.StatusCreated, sessionInfo)
}
