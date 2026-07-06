// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/models"
)

// ThriftAuthorization - Thrift Authorization Middleware
func ThriftAuthorization(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		if SkipLegacyClients(ctx) {
			return next(ctx)
		}

		thriftCtx := ctx.(*models.ThriftReqContext)

		if !thriftCtx.HasSessionID() {
			return next(thriftCtx)
		}

		// We have to duplicate the cookie validation logic a bit here, and change
		// how we respond to Thrift proxy requests because the Thrift JS
		// implementation does not allow for introspection of responses that are
		// anything other than 200 response status:
		// https://github.com/omnisci/mapd-connector/blob/master/thrift/browser/thrift.js#L342-L345
		var sID heavy.TSessionId
		if !thriftCtx.IsWhitelistedMethod() {
			authCookie, err := ctx.Cookie(config.Other.CookieHeavyDBAuth)
			if err == nil && authCookie.Value != "" {
				token := models.AuthToken(authCookie.Value)
				claims, sessionInfoRef, tokenValidationErr := token.Validate(ctx)
				if claims != nil {
					// We need to add the claims to the immerseCtx if found, just in case the token is valid, but the
					//  DB context isn't, so that the claims are available to later middleware and handlers in the flow
					//  for a given proxy request.
					thriftCtx.Claims = claims
					if tokenValidationErr == nil {
						sID = sessionInfoRef.SessionID
					}
				}
			} else {
				sID = "INVALID_SESSION"
			}
		}

		if err = rewriteBodyThriftSessionID(sID, thriftCtx); err != nil {
			return err
		}

		return next(thriftCtx)
	}
}

func rewriteBodyThriftSessionID(sessionID heavy.TSessionId, ctx *models.ThriftReqContext) (err error) {
	if err = ctx.SetSessionID(string(sessionID)); err != nil {
		config.Log.Error("Could not update session id: ", err)
		return err
	}
	return
}
