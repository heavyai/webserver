// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"net/http/httputil"

	"github.com/heavyai/webserver/internal/config"

	"github.com/labstack/echo/v4"
)

// TranscribeProxyHandler - Transcribe Proxy Handler
// Proxies requests to whisper server inference endpoint for speech to text
func TranscribeProxyHandler(ctx echo.Context) error {
	proxy := httputil.NewSingleHostReverseProxy(config.Paths.TranscriptionURL)

	proxy.Director = func(req *http.Request) {
		req.URL.Host = config.Paths.TranscriptionURL.Host
		req.URL.Scheme = config.Paths.TranscriptionURL.Scheme
		req.URL.Path = "/inference"
		req.Host = config.Paths.TranscriptionURL.Host
	}

	proxy.ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}
