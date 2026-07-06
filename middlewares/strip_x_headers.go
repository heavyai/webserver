// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
)

// StripRequestXHeaders - required comment to tell you it strips X-type custom headers
func StripRequestXHeaders(stripXHeaders []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if !config.Enable.StripXHeadersMiddleware {
				return next(ctx)
			}

			for _, xHeaderToStrip := range stripXHeaders {
				ctx.Request().Header.Del(xHeaderToStrip)
			}
			return next(ctx)
		}
	}
}
