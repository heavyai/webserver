// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
)

// RedirectPort - Redirect Port Middleware
func RedirectPort(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		request := ctx.Request()
		requestHost, _, err := net.SplitHostPort(request.Host)
		if err != nil {
			requestHost = request.Host
		}
		redirectURL := url.URL{Scheme: "https", Host: requestHost + ":" + strconv.Itoa(config.HTTP.Port), Path: request.URL.Path, RawQuery: request.URL.RawQuery}

		return ctx.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
	}
}
