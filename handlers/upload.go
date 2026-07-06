// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"

	"github.com/heavyai/webserver/internal/util"
)

// Upload - Upload Handler
// Given a Multi-part form request payload, will parse the form's file(s) and place the resulting file(s) in a configured directory on the local filesystem.
func Upload(ctx echo.Context) (err error) {
	var status int
	immerseCtx, sessionID, err := util.GetUploadSessionID(ctx)

	defer func() {
		if err != nil {
			immerseCtx.String(status, err.Error())
		}
	}()

	form, err := ctx.MultipartForm()

	if err != nil {
		status = http.StatusInternalServerError
		return
	}

	if config.Other.EnableUploadExtensionCheck {
		for _, fileHeaders := range form.File {
			for _, fileHeader := range fileHeaders {
				extension := filepath.Ext(fileHeader.Filename)
				if !util.IsSupportedFileUploadExtension(extension) {
					status = http.StatusBadRequest
					err = fmt.Errorf("file extension unsupported: %s", fileHeader.Filename)
					return
				}
			}
		}
	}

	if config.Enable.ReadOnly {
		status = http.StatusUnauthorized
		err = errors.New("Uploads disabled: server running in read-only mode")
		return
	}

	sessionIDSha256 := sha256.Sum256([]byte(filepath.Base(filepath.Clean(string(sessionID)))))
	obfuscatedSID := hex.EncodeToString(sessionIDSha256[:])
	uploadDir := filepath.Join(config.Paths.DataDir, constants.RelativeImportPath, obfuscatedSID)

	for _, fileHeaders := range form.File {
		for _, fileHeader := range fileHeaders {
			err = util.SaveFileToLocalDir(fileHeader, uploadDir)
			if err != nil {
				status = http.StatusInternalServerError
				return err
			}

			ctx.String(http.StatusAccepted, fileHeader.Filename)
		}
	}
	return
}
