// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
)

// QueryParamAuth - Query Parameter Authentication Middleware
func QueryParamAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		qs := ctx.QueryString()
		if qs == "" {
			return next(ctx)
		}

		util.DeleteHTTPOnlyCookie(ctx, config.Other.CookieHeavyDBAuth)
		util.SetLogoutCookie(ctx)

		session, sessionErr := config.Session.SessionStore.Get(ctx.Request(), constants.CookieModifiedServersJSON)
		if sessionErr != nil {
			util.DeleteHTTPOnlyCookie(ctx, constants.CookieModifiedServersJSON)
			session, _ = config.Session.SessionStore.Get(ctx.Request(), constants.CookieModifiedServersJSON)
		}

		for _, key := range config.Other.ServersJSONParams {
			if len(ctx.QueryParam(key)) > 0 {
				session.Values[key] = ctx.QueryParam(key)
			}
		}

		session.Save(ctx.Request(), ctx.Response())

		return next(ctx)
	}
}
