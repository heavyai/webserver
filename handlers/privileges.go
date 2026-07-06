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

// FetchSessionPrivileges - FetchSessionPrivileges handler
func FetchSessionPrivileges(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	dbName := immerseCtx.GetDBNameFromParams()

	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(dbName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context: '"+dbName+"'", err)
		ctx.NoContent(http.StatusNotFound)
		return err
	}

	thriftRoles, err := util.GetDBPrivileges(*sessionInfo)
	if err != nil {
		config.Log.Error("Could not fetch roles for dbName '"+dbName+"'", err)
		ctx.String(http.StatusNotFound, "Could not fetch roles for dbName '"+dbName+"'")
		return err
	}
	return ctx.JSON(http.StatusOK, thriftRoles)
}

// FetchTablePrivileges - FetchTablePrivileges handler
func FetchTablePrivileges(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	tableName := immerseCtx.Context.Param(constants.TableNameRouteParamName)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context: '"+immerseCtx.DBName+"'", err)
		ctx.NoContent(http.StatusNotFound)
		return err
	}

	tablePrivs, err := util.GetTablePrivileges(sessionInfo.SessionID, tableName)
	if err != nil {
		config.Log.Error("Could not fetch privileges for table '"+tableName+"'", err)
		ctx.String(http.StatusNotFound, "Could not fetch privileges for table '"+tableName+"'")
		return err
	}
	return ctx.JSON(http.StatusOK, tablePrivs)
}
