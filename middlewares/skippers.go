// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/models"
)

type routingException struct {
	Enabled bool
	Path    string
}

// SkipCompress returns false if compression is enabled, and true otherwise
func SkipCompress(ctx echo.Context) bool {
	if config.Enable.Compress {
		return false
	}

	return true
}

// SkipEventStream returns false if not an event stream request
// Echo reverse proxy does not support event stream responses properly
func SkipEventStream(ctx echo.Context) bool {
	if ctx.Request().Header.Get(echo.HeaderAccept) == "text/event-stream" {
		return true
	}

	return false
}

// SkipNonThrift returns false if this is a Thrift request, and true otherwise
func SkipNonThrift(ctx echo.Context) bool {
	// Check the Content-Type header(s) to see if this is a Thrift request
	for _, ct := range ctx.Request().Header["Content-Type"] {
		// JavaScript uses vnd.apache.thrift, everything else uses x-thrift
		if strings.HasPrefix(ct, "application/vnd.apache.thrift") ||
			strings.HasPrefix(ct, "application/x-thrift") {
			return false
		}
	}

	return true
}

// SkipThriftNonConnect - Skips Non-Connect Thrift Methods util
func SkipThriftNonConnect(ctx *models.ThriftReqContext) bool {
	return ctx.GetThriftMethod() != "connect"
}

// SkipLegacyClients - Echo-style Skipper function to support previous authentication method by skipping new authentication middlewares where applied
func SkipLegacyClients(ctx echo.Context) bool {
	if config.Enable.LegacyAuth {
		return !ctx.(*models.ThriftReqContext).UsesTokenAuth
	}

	return false
}

// SkipRedirectSAML - Skip Redirect SAML
func SkipRedirectSAML(ctx echo.Context) bool {
	// Skips non-GET requests
	isNotGet := ctx.Request().Method != http.MethodGet

	// Skips Routing Exceptions
	routingExceptions := []routingException{
		{true, "/session/validate"},
		{true, "/fonts"},
		{true, "/geojson"},
		{config.Enable.SAML, "/saml-error.html"},
		{config.Enable.SAML, "/saml-post"},
		{config.Enable.Jupyter, config.Jupyter.JupyterPrefix},
		{config.Enable.Development || config.Enable.BrowserLogs, "/logs"},
	}
	if config.Enable.ReverseProxies {
		for _, proxy := range config.Other.Proxies {
			routingExceptions = append(routingExceptions, routingException{true, proxy.Path})
		}
	}

	return isNotGet || containsRoutingException(ctx.Request().URL.Path, routingExceptions)
}

func containsRoutingException(path string, routingExceptions []routingException) bool {
	for _, exception := range routingExceptions {
		if exception.Enabled && strings.HasPrefix(path, exception.Path) {
			return true
		}
	}

	return false
}
