// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

func validateOptions(sessionID heavy.TSessionId, options *models.ExportOptions) error {
	if err := util.ValidateSQL(sessionID, options.SQL); err != nil {
		return fmt.Errorf("could not validate sql: %v", err)
	}

	if options.FileName == "" || strings.ContainsAny(options.FileName, `/\'"*?<>:|`) {
		return errors.New("invalid file name")
	}

	validFileTypes := map[string]bool{
		"csv":        true,
		"geojson":    true,
		"geojsonl":   true,
		"shapefile":  true,
		"flatgeobuf": true,
	}

	if !validFileTypes[options.FileType] {
		return errors.New("unknown file type: must be csv, geojson, geojsonl, shapefile, or flatgeobuf")
	}

	if strings.ContainsRune(options.LayerName, '\'') {
		return errors.New("invalid layer name - must not contain any single quotes")
	}

	if options.Compression != "" && options.Compression != "none" && options.Compression != "gzip" && options.Compression != "zip" {
		return errors.New("invalid compression - must be none, gzip, or zip")
	}

	return nil
}

// Export - Export arbitrary SQL to csv, geojson
func Export(ctx echo.Context) (err error) {
	options := new(models.ExportOptions)
	if err = ctx.Bind(options); err != nil {
		config.Log.Error("Export request failure: ", err)
		ctx.NoContent(http.StatusBadRequest)
		return
	}

	immerseCtx := ctx.(*models.ImmerseReqContext)
	dbSession, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return
	}

	sessionID := dbSession.SessionID
	if err = validateOptions(sessionID, options); err != nil {
		config.Log.Error("Bad export options: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Could not export: %v", err))
		return
	}

	if err = util.CopyTo(sessionID, options); err != nil {
		config.Log.Error("Could not export: ", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Could not export: %v", err))
		return
	}

	filename := options.FileName
	if options.Compression == "gzip" {
		filename = fmt.Sprintf("%s.gz", filename)
	} else if options.Compression == "zip" {
		filename = fmt.Sprintf("%s.zip", filename)
	}

	fullpath := filepath.Join(config.Paths.DataDir, constants.RelativeExportPath, string(sessionID), filename)
	needsRemoved := true
	ctx.Response().After(func() {
		// delete the file - this seems to get called multiple times, for some
		// reason, so we'll stick a bool around it.
		if needsRemoved {
			if removeErr := os.Remove(fullpath); removeErr != nil {
				config.Log.Warn("Couldn't remove export: ", removeErr)
			}
			needsRemoved = false
		}
	})

	if err = ctx.Attachment(fullpath, filename); err != nil {
		config.Log.Error("Could not send export with path ", fullpath, ": ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return
	}

	return
}
