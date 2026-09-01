// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
//
// This file uses github.com/mholt/archives, which transitively includes
// github.com/hashicorp/golang-lru/v2 (Copyright (c) 2014 HashiCorp, Inc.).
// golang-lru is licensed under the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, you can obtain
// one at https://mozilla.org/MPL/2.0/. Source is available at
// https://github.com/hashicorp/golang-lru.

package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mholt/archives"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// DashboardsExport - Dashboards Export Handler
func DashboardsExport(ctx echo.Context) (err error) {
	dashboards := new(models.Dashboards)
	if err := ctx.Bind(dashboards); err != nil {
		config.Log.Error("Selected dashboards request failure", err)
		ctx.NoContent(http.StatusBadRequest)
		return err
	}

	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context: '"+immerseCtx.DBName+"'", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	var files []archives.FileInfo

	for _, id := range dashboards.IDs {
		dashboard, err := util.GetDashboard(sessionInfo.SessionID, id)
		if err != nil {
			config.Log.Error(err)
			ctx.String(http.StatusInternalServerError, "Dashboard request failure for given dashboard ID: "+fmt.Sprint(id))
			return err
		}
		dashboardState, err := base64.StdEncoding.DecodeString(dashboard.GetDashboardState())
		if err != nil {
			config.Log.Error("Dashboard state deserialization failure", err)
			ctx.NoContent(http.StatusInternalServerError)
			return err
		}
		exportContents := []byte(dashboard.DashboardName + "\n" + dashboard.DashboardMetadata + "\n" + string(dashboardState))

		var dbVersion string
		dbVersion, err = util.GetHeavyDBVersion(ctx)
		if err != nil {
			config.Log.Error("HeavyDB Version Check Failure")
		}
		exportFileName := strings.ReplaceAll(dashboard.DashboardName, "/", "_") + "_" + dbVersion + "_" + time.Now().Format(time.RFC3339) + ".json"

		contentsCopy := make([]byte, len(exportContents))
		copy(contentsCopy, exportContents)

		file := archives.FileInfo{
			FileInfo:      &models.VirtualFileInfo{FileName: exportFileName, Data: contentsCopy},
			NameInArchive: exportFileName,
			Open: func() (fs.File, error) {
				return &models.VirtualFile{
					Reader: bytes.NewReader(contentsCopy),
					VFI:    models.VirtualFileInfo{FileName: exportFileName, Data: contentsCopy},
				}, nil
			},
		}
		files = append(files, file)
	}

	ctx.Response().Header().Set(echo.HeaderContentType, "application/zip")
	ctx.Response().WriteHeader(http.StatusOK)

	format := archives.Zip{
		SelectiveCompression: true,
	}
	err = format.Archive(context.Background(), ctx.Response(), files)
	if err != nil {
		config.Log.Error("Dashboard export failure", err)
		return err
	}

	return nil
}
