// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"
	"net/url"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// ExtendContext - Extend Context Middleware
func ExtendContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		authCookie, _ := ctx.Cookie(config.Other.CookieHeavyDBAuth)
		authCookiePresent := authCookie != nil

		parsedRefererURL, err := url.Parse(ctx.Request().Header.Get(headers.Referer))
		if err != nil {
			config.Log.Error("Could not parse Referer header: ", ctx.Request().Header.Get(headers.Referer))
			ctx.NoContent(http.StatusBadRequest)
			return err
		}

		dbName := ""
		// For SAML, We only want to set the DB context from the referer header on the request is the same as the
		//  request host.
		if !config.Enable.SAML || (config.Enable.SAML && parsedRefererURL.Host == ctx.Request().Host) {
			dbName = util.URLPathExtractDBContext(parsedRefererURL.Path)
		}

		newImmerseCtx := models.ImmerseReqContext{
			Context:       ctx,
			Fingerprint:   uuid.New(),
			UsesTokenAuth: authCookiePresent,
			DBName:        dbName,
			Claims: &models.ImmerseJWTClaims{
				SessionsInfoMap: models.SessionsInfoMap{},
			},
		}
		return next(&newImmerseCtx)
	}
}

// ThriftExtend - Thrift Extend Context Middleware
func ThriftExtend(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		immerseCtx := ctx.(*models.ImmerseReqContext)
		thriftCtx, err := models.NewThriftReqContext(immerseCtx, useBinaryThrift(ctx))
		if err != nil {
			return err
		}

		if sID := thriftCtx.GetSessionID(); sID == config.Other.SubstituteSessionID {
			immerseCtx.UsesTokenAuth = true
		}
		immerseCtx.Claims.SessionsInfoMap = models.SessionsInfoMap{}

		if config.Enable.Development {
			method := thriftCtx.GetThriftMethod()
			config.Log.Printf("Thrift Method %s for request %s", method, immerseCtx.Fingerprint.String())

			immerseCtx.RequestTime = time.Now()
			recorded := false
			ctx.Response().After(func() {
				// this gets called multiple times, so surround in conditional
				if !recorded {
					responseTime := time.Now()
					duration := responseTime.Sub(immerseCtx.RequestTime)
					config.Log.Printf("Duration of %v for request %s", duration, immerseCtx.Fingerprint.String())
					recorded = true
				}
			})
		}

		return next(thriftCtx)
	}
}

// useBinaryThrift - Returns true if the content-type header is set to
// application/vnd.apache.thrift.binary
func useBinaryThrift(ctx echo.Context) bool {
	for _, ct := range ctx.Request().Header["Content-Type"] {
		if ct == "application/vnd.apache.thrift.binary" {
			return true
		}
	}
	return false
}
