// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"strings"

	"github.com/heavyai/webserver/services"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/models"
)

// GetUploadSessionID - Get Upload SessionConfig ID utility function
func GetUploadSessionID(ctx echo.Context) (echo.Context, heavy.TSessionId, error) {
	if sID := ctx.FormValue("sessionid"); sID != "" && sID != config.Other.SubstituteSessionID {
		return ctx, heavy.TSessionId(sID), nil
	}

	immerseCtx := ctx.(*models.ImmerseReqContext)
	dbSession, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		return nil, "", err
	}

	return immerseCtx, dbSession.SessionID, nil
}

// GetDashboard - GetDashboard util method
func GetDashboard(sessionID heavy.TSessionId, dashboardID int32) (dashboard *heavy.TDashboard, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.GetDashboard(thriftClient.Ctx, sessionID, dashboardID)
}

// GetDashboards - GetDashboards util method
func GetDashboards(sessionID heavy.TSessionId) (dashboards []*heavy.TDashboard, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.GetDashboards(thriftClient.Ctx, sessionID)
}

// GetHeavyDBVersion - GetHeavyDBVersion util method
func GetHeavyDBVersion(ctx echo.Context) (version string, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.GetVersion(thriftClient.Ctx)
}

// ValidateSQL - Validates raw sql, returning an error if it's invalid
func ValidateSQL(sessionID heavy.TSessionId, sql string) (err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()

	_, err = thriftClient.Client.SqlValidate(thriftClient.Ctx, sessionID, sql)
	return
}

// CopyTo - exports data to a file using COPY TO. Options should be validated
// before calling this function
func CopyTo(sessionID heavy.TSessionId, options *models.ExportOptions) (err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()

	var sqlBuilder strings.Builder
	fmt.Fprintf(&sqlBuilder, `COPY (%s) TO '%s' WITH (file_type='%s'`, options.SQL, options.FileName, options.FileType)
	if options.FileType == "csv" {
		sqlBuilder.WriteString(",header='true'")
	}
	if options.LayerName != "" {
		fmt.Fprintf(&sqlBuilder, ",layer_name='%s'", options.LayerName)
	}
	if options.Compression != "" {
		fmt.Fprintf(&sqlBuilder, ",file_compression='%s'", options.Compression)
	}
	sqlBuilder.WriteRune(')')

	_, err = thriftClient.Client.SqlExecute(thriftClient.Ctx, sessionID, sqlBuilder.String(), true, fmt.Sprintf("export-%v", options.FileName), -1, -1)
	return
}

// SendInterrupt - SendInterrupt util method
func SendInterrupt(userSessionID heavy.TSessionId, interruptSessionID heavy.TSessionId) (err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()

	return thriftClient.Client.Interrupt(thriftClient.Ctx, userSessionID, interruptSessionID)
}

// GetTableDetails - GetTableDetails util method
func GetTableDetails(sessionID heavy.TSessionId, tableName string) (result *heavy.TTableDetails, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()

	return thriftClient.Client.GetTableDetails(
		thriftClient.Ctx,
		sessionID,
		tableName,
	)
}

// GetUserInfo - GetUserInfo util method
func GetUserInfo(sessionID heavy.TSessionId) (result []*heavy.TUserInfo, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()

	return thriftClient.Client.GetUsersInfo(thriftClient.Ctx, sessionID)
}

// PutImmerseUsersMetadata - PutImmerseUsersMetadata util method
func PutImmerseUsersMetadata(sessionID heavy.TSessionId, userMetadata []*heavy.TImmerseUserMetadata) (err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()

	return thriftClient.Client.PutImmerseUsersMetadata(thriftClient.Ctx, sessionID, userMetadata)
}

// ErrorMsg - ErrorMsg util method
func ErrorMsg(err error) string {
	errStr := err.Error()
	if strings.HasPrefix(errStr, "TDBException({ErrorMsg:") {
		errStr = errStr[23 : len(errStr)-2]
	}
	return errStr
}
