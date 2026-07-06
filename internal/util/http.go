// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net/http"

	"github.com/heavyai/webserver/internal/config"

	"github.com/labstack/echo/v4"
)

// GetSchemeForRequest - Get Scheme For Request util
func GetSchemeForRequest(req *http.Request) string {
	// Pretty much the only simple way to determine the request's scheme reliably in Go land:
	// https://stackoverflow.com/questions/40826664/get-scheme-of-the-current-request-url#comment92621257_40826664
	if req.TLS != nil {
		return "https"
	}
	return "http"
}

// ReqIsSameOrigin - Req Is Same Origin util
func ReqIsSameOrigin(req *http.Request) bool {
	reqOriginHeader := req.Header.Get(echo.HeaderOrigin)
	reqScheme := GetSchemeForRequest(req)
	reqURLOriginStr := reqScheme + "://" + req.Host
	return reqURLOriginStr == reqOriginHeader
}

// GetAllowedOrigin - GetAllowedOrigin util
func GetAllowedOrigin(req http.Request) (allowedOrigin string) {
	if config.Enable.UltraSecureMode {
		allowedOrigin = config.HTTP.SecureACAOUri
	} else {
		allowedOrigin = req.Header.Get(echo.HeaderOrigin)
	}
	return
}

// SetSecureResponseHeaders - SetSecureResponseHeaders util
func SetSecureResponseHeaders(resp echo.Response) {
	resp.Header().Set(
		echo.HeaderXFrameOptions,
		"SAMEORIGIN",
	)
	resp.Header().Set(
		echo.HeaderStrictTransportSecurity,
		"max-age=63072000;", // 2 years
	)
	return
}
