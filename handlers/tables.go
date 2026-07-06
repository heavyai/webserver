// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// CreateOdbcFSITableHandler - util
func CreateOdbcFSITableHandler(ctx echo.Context) (err error) {
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	fsiConfig := models.OdbcFSIConfig{}
	if err = ctx.Bind(&fsiConfig); err != nil {
		config.Log.Error("Bad request body", err)
		ctx.NoContent(http.StatusBadRequest)
		return
	}

	serverName := models.GetFsiServerName(fsiConfig.GetTableName())
	if err = fsiConfig.CreateOdbcFSIServer(sessionID, serverName); err != nil {
		config.Log.Error("Could not create FSI server: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Could not create FSI server: %v", err))
		return
	}

	if err = fsiConfig.CreateOdbcFSIUserMapping(sessionID, serverName); err != nil {
		config.Log.Error("Could not create FSI user mapping: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Could not create FSI user mapping: %v", err))
		return
	}

	err = fsiConfig.CreateForeignTable(sessionID, serverName)
	if err != nil {
		config.Log.Error("Could not create foreign table: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf(`Could not create foreign table: %s`, err))
		return
	}

	return ctx.NoContent(201)
}

// CreateS3FSITableHandler - util
func CreateS3FSITableHandler(ctx echo.Context) (err error) {
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	fsiConfig := models.S3FSIConfig{}
	if err = ctx.Bind(&fsiConfig); err != nil {
		config.Log.Error("Bad request body", err)
		ctx.NoContent(http.StatusBadRequest)
		return
	}

	serverName := models.GetFsiServerName(fsiConfig.GetTableName())
	if err = fsiConfig.CreateS3FSIServer(sessionID, serverName); err != nil {
		config.Log.Error("Could not create FSI server: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Could not create FSI server: %v", err))
		return
	}

	if fsiConfig.GetAccessType() == models.S3AccessTypeSelect ||
		(fsiConfig.GetAccessID() != "" && fsiConfig.GetSecretKey() != "") {
		if err = fsiConfig.CreateS3FSIUserMapping(sessionID, serverName); err != nil {
			config.Log.Error("Could not create FSI user mapping: ", err)
			ctx.String(http.StatusBadRequest, fmt.Sprintf("Could not create FSI user mapping: %v", err))
			return
		}
	}

	err = fsiConfig.CreateForeignTable(sessionID, serverName)
	if err != nil {
		config.Log.Error("Could not create foreign table: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf(`Could not create foreign table: %s`, err))
		return
	}

	return ctx.NoContent(201)
}

// CreateServerFileFSITableHandler - handler
func CreateServerFileFSITableHandler(ctx echo.Context) (err error) {
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	fsiConfig := models.ServerFileConfig{}
	if err = ctx.Bind(&fsiConfig); err != nil {
		config.Log.Error("Bad request body", err)
		ctx.NoContent(http.StatusBadRequest)
		return
	}

	serverName := models.GetFsiServerName(fsiConfig.GetTableName())
	if err = fsiConfig.CreateS3FSIServer(sessionID, serverName); err != nil {
		config.Log.Error("Could not create FSI server: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Could not create FSI server: %v", err))
		return
	}

	err = fsiConfig.CreateForeignTable(sessionID, serverName)
	if err != nil {
		config.Log.Error("Could not create foreign table: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf(`Could not create foreign table: %s`, err))
		return
	}

	return ctx.NoContent(201)
}

// DetectColumnTypesHandler - handler
func DetectColumnTypesHandler(ctx echo.Context) (err error) {
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	dctConfig := models.DetectColumnTypesConfig{}
	if err = ctx.Bind(&dctConfig); err != nil {
		config.Log.Error("Bad request body", err)
		ctx.NoContent(http.StatusBadRequest)
		return
	}

	result, err := util.DetectColumnTypes(
		sessionID,
		dctConfig,
	)
	if err != nil {
		config.Log.Error("could not detect column types => ", err)
		return ctx.NoContent(http.StatusBadGateway)
	}

	return ctx.JSON(http.StatusOK, result)
}

// DropForeignTableHandler - util
func DropForeignTableHandler(ctx echo.Context) (err error) {
	tableName := ctx.Param(constants.TableNameRouteParamName)
	serverName := models.GetFsiServerName(tableName)
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	err = models.DropForeignTable(sessionID, tableName)
	if err != nil {
		config.Log.Error("Could not drop foreign table: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf(`Could not drop foreign table: %s`, err))
		return
	}

	err = models.DropServer(sessionID, serverName)
	if err != nil {
		config.Log.Error("Could not drop FSI server: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf(`Could not drop FSI server: %s`, err))
		return
	}

	return ctx.NoContent(200)
}

// RenameForeignTableHandler - util
func RenameForeignTableHandler(ctx echo.Context) (err error) {
	tableName := ctx.Param(constants.TableNameRouteParamName)
	newTableName := ctx.Param(constants.NewTableNameRouteParamName)
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	err = util.RenameForeignTable(sessionID, tableName, newTableName)
	if err != nil {
		config.Log.Error("Could not rename foreign table: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf(`Could not rename foreign table: %s`, err))
		return
	}
	return ctx.NoContent(202)
}

// RenameForeignTableColumnsHandler - util
func RenameForeignTableColumnsHandler(ctx echo.Context) (err error) {
	tableName := ctx.Param(constants.TableNameRouteParamName)
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}
	columnRenameMaps := new([]models.ColumnRenameMap)
	// We can't use Echo's .Bind() because it doesn't support anything but structs for some reason
	//  see: https://github.com/labstack/echo/issues/1565#issuecomment-813625456
	if err = json.NewDecoder(ctx.Request().Body).Decode(columnRenameMaps); err != nil {
		config.Log.Error("Bad request body", err)
		ctx.NoContent(http.StatusBadRequest)
		return
	}

	queryErrors := util.RenameForeignTableColumns(sessionID, tableName, columnRenameMaps)
	if len(queryErrors) > 0 {
		err = fmt.Errorf(`rename foreign table column(s) failed: %s`, queryErrors)
		config.Log.Error("Could not rename foreign table column(s): ", queryErrors)
		ctx.String(http.StatusBadRequest, fmt.Sprintf(`Could not rename foreign table column(s):  %s`, queryErrors))
		return
	}
	return ctx.NoContent(202)
}

// UpdateForeignTableRefreshScheduleHandler - util
func UpdateForeignTableRefreshScheduleHandler(ctx echo.Context) (err error) {
	tableName := ctx.Param(constants.TableNameRouteParamName)
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}
	refreshInfo := heavy.TTableRefreshInfo{}
	if err = ctx.Bind(&refreshInfo); err != nil {
		config.Log.Error("Bad request body: ", err)
		ctx.NoContent(http.StatusBadRequest)
		return
	}

	err = util.UpdateForeignTableRefreshSchedule(sessionID, tableName, refreshInfo)
	if err != nil {
		config.Log.Error("Could not update foreign table refresh schedule: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf(`Could not update foreign table refresh schedule: %s`, err))
		return
	}
	return ctx.NoContent(202)
}

// RefreshTableHandler - util
func RefreshTableHandler(ctx echo.Context) (err error) {
	tableName := ctx.Param(constants.TableNameRouteParamName)
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}
	tableDetails, err := util.RefreshForeignTable(sessionID, tableName)
	if err != nil {
		config.Log.Error("Could not refresh foreign table: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf(`Could not refresh foreign table: %s`, err))
		return
	}
	return ctx.JSON(202, tableDetails)
}
