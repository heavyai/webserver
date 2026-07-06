// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"errors"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

// GetMetadataForUser -- returns metadata for the current user
func GetMetadataForUser(ctx echo.Context, userInfo []*heavy.TUserInfo) (metadata string, err error) {
	immerseCtx := ctx.(*ImmerseReqContext)
	username := immerseCtx.Claims.Username

	for _, info := range userInfo {
		if info.Username == username {
			if info.ImmerseMetadataJSON == "" {
				// Ensure return value is parseable JSON
				metadata = "{}"
				return
			}
			metadata = info.ImmerseMetadataJSON
			return
		}
	}

	err = errors.New("user metadata not found")
	return
}
