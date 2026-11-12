// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestReverseProxyAppendsDirectPeerAndSanitizesForwardedHeader(t *testing.T) {
	var forwardedHeader string
	var forwardedFor string
	var forwardedProto string
	var realIP string

	backend := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		forwardedHeader = request.Header.Get("Forwarded")
		forwardedFor = request.Header.Get(echo.HeaderXForwardedFor)
		forwardedProto = request.Header.Get(echo.HeaderXForwardedProto)
		realIP = request.Header.Get(echo.HeaderXRealIP)
		response.WriteHeader(http.StatusNoContent)
	}))
	defer backend.Close()

	targetURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	e.Use(ReverseProxy(targetURL))
	e.Any("/*", func(ctx echo.Context) error {
		return ctx.NoContent(http.StatusTeapot)
	})

	// Echo intentionally ignores client-supplied forwarding headers unless a trusted IP extractor is configured.
	request := httptest.NewRequest(http.MethodGet, "/proxy", nil)
	request.RemoteAddr = "192.0.2.10:4321"
	request.Header.Set("Forwarded", "for=198.51.100.23;proto=https")
	request.Header.Set(echo.HeaderXForwardedFor, "198.51.100.23")
	request.Header.Set(echo.HeaderXForwardedProto, "https")
	request.Header.Set(echo.HeaderXRealIP, "198.51.100.23")
	response := httptest.NewRecorder()

	e.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected backend status %d, got %d", http.StatusNoContent, response.Code)
	}
	if forwardedHeader != "" {
		t.Fatalf("expected Forwarded header to be stripped, got %q", forwardedHeader)
	}
	if forwardedFor != "198.51.100.23, 192.0.2.10" {
		t.Errorf("expected direct peer appended to X-Forwarded-For, got %q", forwardedFor)
	}
	if forwardedProto != "http" {
		t.Errorf("expected request scheme in X-Forwarded-Proto, got %q", forwardedProto)
	}
	if realIP != "192.0.2.10" {
		t.Errorf("expected direct peer in X-Real-IP, got %q", realIP)
	}
}

func TestReverseProxySkipsEventStreams(t *testing.T) {
	backendCalled := false
	backend := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		backendCalled = true
		response.WriteHeader(http.StatusNoContent)
	}))
	defer backend.Close()

	targetURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	e.Use(ReverseProxy(targetURL))
	e.GET("/*", func(ctx echo.Context) error {
		return ctx.NoContent(http.StatusAccepted)
	})

	request := httptest.NewRequest(http.MethodGet, "/events", nil)
	request.Header.Set(echo.HeaderAccept, "text/event-stream")
	response := httptest.NewRecorder()

	e.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected skipped request status %d, got %d", http.StatusAccepted, response.Code)
	}
	if backendCalled {
		t.Fatal("expected event stream request to bypass the Echo reverse proxy")
	}
}

func TestReverseProxySupportsWebSocketUpgrade(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if !strings.EqualFold(request.Header.Get(echo.HeaderConnection), "upgrade") {
			http.Error(response, "missing connection upgrade", http.StatusBadRequest)
			return
		}
		if !strings.EqualFold(request.Header.Get(echo.HeaderUpgrade), "websocket") {
			http.Error(response, "missing websocket upgrade", http.StatusBadRequest)
			return
		}

		response.Header().Set(echo.HeaderConnection, "Upgrade")
		response.Header().Set(echo.HeaderUpgrade, "websocket")
		response.WriteHeader(http.StatusSwitchingProtocols)
	}))
	defer backend.Close()

	targetURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	e.Use(ReverseProxy(targetURL))
	e.Any("/*", func(ctx echo.Context) error {
		return ctx.NoContent(http.StatusTeapot)
	})

	frontend := httptest.NewServer(e)
	defer frontend.Close()

	request, err := http.NewRequest(http.MethodGet, frontend.URL+"/socket", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderConnection, "Upgrade")
	request.Header.Set(echo.HeaderUpgrade, "websocket")
	request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	request.Header.Set("Sec-WebSocket-Version", "13")

	response, err := frontend.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSwitchingProtocols {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected websocket upgrade status %d, got %d: %s", http.StatusSwitchingProtocols, response.StatusCode, body)
	}
}
