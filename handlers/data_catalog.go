// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
)

// DataCatalog - DataCatalog Handler
func DataCatalog(ctx echo.Context) error {
	if config.Enable.ReadOnly {
		return echo.NewHTTPError(http.StatusUnauthorized, "Uploads disabled: server running in read-only mode")
	}

	filename := filepath.Clean(ctx.FormValue("filename"))

	if len(filename) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Filename cannot be empty")
	}

	if strings.HasPrefix(filename, "..") || strings.HasPrefix(filename, "/") {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file path: "+filename)
	}

	absDataCatalogDir, _ := filepath.Abs(config.Paths.DataCatalogDir)
	sourceFile, err := filepath.Abs(filepath.Join(config.Paths.DataCatalogDir, filename))
	if err != nil || !strings.HasPrefix(sourceFile, absDataCatalogDir) {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not find file in data catalog: "+filename)
	}

	_, err = os.Stat(sourceFile)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not find file in data catalog: "+filename)
	}

	_, sessionID, err := util.GetUploadSessionID(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not resolve session ID for upload: ", err)
	}

	sessionIDSha256 := sha256.Sum256([]byte(filepath.Base(filepath.Clean(string(sessionID)))))
	obfuscatedSID := hex.EncodeToString(sessionIDSha256[:])
	uploadDir := filepath.Join(config.Paths.DataDir, constants.RelativeImportPath, obfuscatedSID)

	destFile := filepath.Join(uploadDir, filename)

	// Ignore errors - we don't care if this fails (most often, the file won't exist)
	os.Remove(destFile)

	err = os.MkdirAll(uploadDir, 0700)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not make upload directory: "+err.Error())
	}

	err = os.Symlink(sourceFile, destFile)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not link data catalog file: "+err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}
