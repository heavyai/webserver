// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// DataCatalogCache handles caching for data catalog - manifest should never be cached,
// but we want far-future caching for the big-ish images used for the card headers
func DataCatalogCache(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		responseHeader := ctx.Response().Header()

		if strings.HasSuffix(ctx.Request().RequestURI, "/manifest.json") {
			responseHeader.Del("Cache-Control")
			responseHeader.Add("Cache-Control", "no-cache, no-store, must-revalidate")
		} else {
			responseHeader.Del("Cache-Control")
			responseHeader.Add("Cache-Control", "max-age=31536000") // One year
		}

		return next(ctx)
	}
}
