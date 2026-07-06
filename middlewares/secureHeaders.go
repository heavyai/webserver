// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
)

// SecureHeaders - SecureHeaders handler
func SecureHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if config.Enable.UltraSecureMode {
			util.SetSecureResponseHeaders(*ctx.Response())
		}
		return next(ctx)
	}
}
