// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/models"
)

// EventProxy - Event Proxy Middleware
func EventProxy(targetURL *url.URL) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			proxyHTTPWithFlushInterval(targetURL, 100*time.Millisecond).ServeHTTP(ctx.Response(), ctx.Request())

			return
		}
	}
}

// ReverseProxy - Reverse Proxy Middleware returns middleware that proxies requests
func ReverseProxy(targetURL *url.URL) echo.MiddlewareFunc {
	return middleware.ProxyWithConfig(middleware.ProxyConfig{
		Skipper:  SkipEventStream,
		Balancer: SingleTargetBalancer(targetURL),
	})
}

// ThriftProxy - Thrift Proxy Middleware
// @Summary Thrift Proxy
// @Description Proxies Thrift requests to HeavyDB
// @Tags Immerse
// @Accept application/vnd.apache.thrift
// @Accept application/x-thrift
// @Success 200
// @Router / [post]
func ThriftProxy() echo.MiddlewareFunc {
	jsonProxy := middleware.ProxyWithConfig(middleware.ProxyConfig{
		Skipper:  SkipNonThrift,
		Balancer: SingleTargetBalancer(config.Paths.BackendURL),
	})
	binaryProxy := middleware.ProxyWithConfig(middleware.ProxyConfig{
		Skipper:  SkipNonThrift,
		Balancer: SingleTargetBalancer(config.Paths.BinaryBackendURL),
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			thriftCtx := ctx.(*models.ThriftReqContext)
			if thriftCtx.IsBinary() {
				return binaryProxy(next)(ctx)
			}
			return jsonProxy(next)(ctx)
		}
	}
}

// SingleTargetBalancer Proxy Balancer for a Single Target
func SingleTargetBalancer(url *url.URL) middleware.ProxyBalancer {
	targetURL := []*middleware.ProxyTarget{
		{
			URL: url,
		},
	}
	return middleware.NewRoundRobinBalancer(targetURL)
}

// proxyHTTPWithFlushInterval - Reverse Proxy for special cases like event stream
func proxyHTTPWithFlushInterval(targetURL *url.URL, interval time.Duration) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.FlushInterval = interval
	return proxy
}

// AppendPathQueryProxy - Middleware that modifies the path and query parameters for the target URL
func AppendPathQueryProxy(targetURL *url.URL, removePrefix string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			req := ctx.Request()
			res := ctx.Response()

			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			proxy.Director = func(req *http.Request) {
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host

				// Remove the specified prefix from the request path
				req.URL.Path = singleJoiningSlash(targetURL.Path, strings.TrimPrefix(req.URL.Path, removePrefix))

				// Append the original request query parameters to the target URL query parameters
				if targetURL.RawQuery == "" || req.URL.RawQuery == "" {
					req.URL.RawQuery = targetURL.RawQuery + req.URL.RawQuery
				} else {
					req.URL.RawQuery = targetURL.RawQuery + "&" + req.URL.RawQuery
				}

				// Set the Host header to the target URL's host
				req.Host = targetURL.Host
				req.Header.Set("Host", targetURL.Host)

				// Set the host_header to us.i.posthog.com
				req.Header.Set("host_header", targetURL.Host)

				// Set X-Forwarded-For header
				if clientIP := ctx.RealIP(); clientIP != "" {
					req.Header.Set("X-Forwarded-For", clientIP)
				}

				// Set X-Forwarded-Proto header
				req.Header.Set("X-Forwarded-Proto", ctx.Scheme())
			}
			proxy.ServeHTTP(res, req)

			return nil
		}
	}
}

// singleJoiningSlash - Helper function to join two paths with a single slash
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
