// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/models"
)

// RedirectSAML - Redirect SAML Middleware
func RedirectSAML(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if !config.Enable.SAML || SkipRedirectSAML(ctx) {
			return next(ctx)
		}

		SAMLUrlWithRelayState, err := buildSAMLUrlWithRelayQuery(*ctx.Request())
		if err != nil {
			ctx.Error(errors.New("Could not set SAML redirect"))
			return err
		}

		authCookie, err := ctx.Cookie(config.Other.CookieHeavyDBAuth)
		if err != nil || authCookie.Value == "" {
			config.Log.Info("No cookie found, redirecting to SAML provider for authentication")
			return ctx.Redirect(http.StatusFound, SAMLUrlWithRelayState)
		}

		token := models.AuthToken(authCookie.Value)
		_, sessionInfoRef, tokenValidationError := token.Validate(ctx)
		if tokenValidationError != nil {
			config.Log.Info("Could not validate session, redirecting to SAML provider: ", tokenValidationError)
			if sessionInfoRef != nil {
				config.Log.Info("DBSessionInfo returned: ", *sessionInfoRef)
			}
			return ctx.Redirect(http.StatusFound, SAMLUrlWithRelayState)
		}

		return next(ctx)
	}
}

func buildSAMLUrlWithRelayQuery(request http.Request) (SAMLUrlWithRelayQuery string, err error) {
	baseQuery := url.Values{}
	SAMLurl := config.Other.SAMLurl

	if err == nil {
		baseQuery, err = url.ParseQuery(SAMLurl.RawQuery)
		if config.Other.ReturnURLFlag == true {
			baseQuery.Add("returnurl", request.URL.String()) // Lack of camelCase is intentional here due to Oracle Access Manager specs
		} else {
			baseQuery.Add("RelayState", request.URL.String())
		}
	}

	var resultantURLBuf strings.Builder
	resultantURLBuf.WriteString(strings.Split(SAMLurl.String(), "?")[0])
	resultantURLBuf.WriteByte('?')
	resultantURLBuf.WriteString(baseQuery.Encode())
	if SAMLurl.Fragment != "" {
		resultantURLBuf.WriteByte('#')
		resultantURLBuf.WriteString(SAMLurl.EscapedFragment())
	}
	SAMLUrlWithRelayQuery = resultantURLBuf.String()
	return
}
