// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"net/http/httputil"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
)

// ReverseProxy - ReverseProxy Handler
// Handles path proxying that Echo reverse proxy does not handle
func ReverseProxy(proxy config.ReverseProxy) echo.HandlerFunc {
	proxyHandler := httputil.NewSingleHostReverseProxy(proxy.Target)
	proxyHandler.Director = func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", proxy.Target.Host)
		req.URL.Scheme = proxy.Target.Scheme
		req.URL.Host = proxy.Target.Host
		req.URL.Path = req.URL.Path[len(proxy.Path):]
	}

	return echo.WrapHandler(proxyHandler)
}
