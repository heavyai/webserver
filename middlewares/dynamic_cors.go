// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
)

// DynamicCORSMiddleware - Dynamic CORS Middleware
func DynamicCORSMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if util.ReqIsSameOrigin(ctx.Request()) {
			return next(ctx)
		}

		allowableOrigins := []string{util.GetAllowedOrigin(*ctx.Request())}

		if config.Enable.AllowAnyOrigin {
			allowableOrigins = []string{"*"}
		}

		dynamicCORSConfig := middleware.CORSConfig{
			AllowOrigins: allowableOrigins,
			AllowHeaders: []string{
				echo.HeaderAccept,
				echo.HeaderCookie,
				"Cache-Control", // Echo doesn't have this header constant
				echo.HeaderContentType,
				echo.HeaderXRequestedWith,
				config.Session.SessionIDHeader,
			},
			AllowMethods: constants.AllowedHTTPMethods,
		}

		// TODO: Make robust with explicit header deduplication:
		// https://heavyai.atlassian.net/browse/FE-10146
		respACACHeader := ctx.Response().Header().Get(echo.HeaderAccessControlAllowCredentials)
		if len(respACACHeader) < 1 || respACACHeader == "" {
			dynamicCORSConfig.AllowCredentials = true
		}

		CORSMiddleware := middleware.CORSWithConfig(dynamicCORSConfig)
		CORSHandler := CORSMiddleware(next)
		return CORSHandler(ctx)
	}
}
