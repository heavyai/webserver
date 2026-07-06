// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
)

// Logs - Log Handler
// Returns text of select server log files
func Logs(logFilePath string) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if !config.Enable.Development {
			tokenClaims := util.GetTokenClaims(ctx)
			sessionInfo, err := util.GetDBSessionInfo(tokenClaims, "")
			if err != nil || !sessionInfo.IsSuper {
				return ctx.NoContent(http.StatusNotFound)
			}
		}

		l, err := os.ReadFile(logFilePath)
		if err != nil {
			return ctx.NoContent(http.StatusNotFound)
		}

		return ctx.String(http.StatusOK, string(l))
	}
}
