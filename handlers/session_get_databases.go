// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// SessionGetDatabases - SessionGetDatabases handler
func SessionGetDatabases(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo := immerseCtx.Claims.SessionsInfoMap.GetSingleSessionInfo()
	thriftDBAccessList, err := util.GetDBAccessList(sessionInfo.SessionID)
	if err != nil {
		config.Log.Error("Could not retrieve thriftDBAccessList: ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return
	}
	DBAccessList := models.CreateDBAccessList(thriftDBAccessList)

	return ctx.JSON(http.StatusOK, DBAccessList)
}
