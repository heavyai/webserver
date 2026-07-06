// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ThriftOnly - Thrift Only Middleware
func ThriftOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		if SkipNonThrift(ctx) {
			ctx.String(http.StatusNotAcceptable, "Thrift request not detected")
			return
		}

		return next(ctx)
	}
}
