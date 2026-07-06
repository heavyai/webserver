// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"

	"github.com/heavyai/webserver/internal/config"

	"github.com/heavyai/webserver/internal/util"

	"github.com/labstack/echo/v4"
)

func setResponseACAOHeaderFromRequest(req http.Request, resp echo.Response) {
	resp.Header().Set(
		echo.HeaderAccessControlAllowOrigin,
		util.GetAllowedOrigin(req),
	)
}

// ACAOHeaderOverwrite - Backend services (e.g. Thrift) may set the ACAO header to be "*" and the proxy middleware then takes that
// and creates a bad/ "[origin CORS middleware set earlier], *" ACAO header value.  To fix this, we hook into the
// response flow and manually set that header to the requestor's Origin header before sending it back over.
func ACAOHeaderOverwrite(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Before(func() {
			setResponseACAOHeaderFromRequest(*ctx.Request(), *ctx.Response())
			if config.Enable.UltraSecureMode {
				util.SetSecureResponseHeaders(*ctx.Response())
			}
		})
		return next(ctx)
	}
}
