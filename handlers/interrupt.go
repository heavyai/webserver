// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// Interrupt - Interrupt handler
func Interrupt(ctx echo.Context) (err error) {
	if !config.Enable.QueryInterrupt {
		config.Log.Error("Query Interrupt not enabled in HeavyDB", err)
		ctx.String(http.StatusInternalServerError, "Could not send interrupt")
		return
	}

	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return
	}

	var interruptSessionID heavy.TSessionId
	userSessionID := sessionInfo.SessionID
	qp := ctx.QueryParam("session-id")
	if len(qp) > 0 {
		interruptSessionID = heavy.TSessionId(qp)
	} else {
		interruptSessionID = userSessionID
	}

	err = util.SendInterrupt(userSessionID, interruptSessionID)
	if err != nil {
		config.Log.Error("Sending interrupt failed", err)
		ctx.String(http.StatusInternalServerError, "Could not send interrupt")
		return err
	}

	return ctx.NoContent(http.StatusOK)
}
