// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// ProtectRoute - Protect Route Middleware
func ProtectRoute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		immerseCtx := ctx.(*models.ImmerseReqContext)
		token, err := util.GetTokenFromAuthCookie(ctx)
		if err != nil {
			ctx.String(http.StatusUnauthorized, err.Error())
			return err
		}
		_, _, tokenValidationError := token.Validate(ctx)
		if tokenValidationError != nil {
			util.DeleteHTTPOnlyCookie(ctx, config.Other.CookieHeavyDBAuth)
			util.SetLogoutCookie(ctx)
			if config.Enable.CustomIntegrationAuth {
				util.DeleteHTTPOnlyCookie(ctx, constants.CookieModifiedServersJSON)
			}
			ctx.String(http.StatusUnauthorized, tokenValidationError.Error())
			return tokenValidationError
		}
		return next(immerseCtx)
	}
}

// AuthorizeRoute - Authorize Route Middleware
func AuthorizeRoute(authRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			immerseCtx := ctx.(*models.ImmerseReqContext)
			sessionInfo, err := util.GetDBSessionInfo(immerseCtx.Claims, immerseCtx.DBName)
			if err != nil {
				config.Log.Error(err)
				return ctx.NoContent(http.StatusInternalServerError)
			}

			singleSessionInfo := *immerseCtx.Claims.SessionsInfoMap.GetSingleSessionInfo()
			roles, err := util.GetDBRoles(singleSessionInfo, "")
			if err != nil {
				config.Log.Error(err)
				return ctx.NoContent(http.StatusInternalServerError)
			}

			permittedRole := false
			for _, role := range roles {
				if role == authRole {
					permittedRole = true
				}
			}

			if sessionInfo.IsSuper || permittedRole {
				return next(ctx)
			}

			return ctx.NoContent(http.StatusUnauthorized)
		}
	}
}
