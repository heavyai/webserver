// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestConfigureTrustedProxies(t *testing.T) {
	tests := []struct {
		name           string
		cidrs          []string
		remoteAddr     string
		forwardedFor   string
		forwardedProto string
		wantIP         string
		wantScheme     string
	}{
		{
			name:           "forwarding headers ignored by default",
			remoteAddr:     "192.0.2.10:443",
			forwardedFor:   "198.51.100.20",
			forwardedProto: "https",
			wantIP:         "192.0.2.10",
			wantScheme:     "http",
		},
		{
			name:           "headers accepted from configured proxy",
			cidrs:          []string{"192.0.2.0/24"},
			remoteAddr:     "192.0.2.10:443",
			forwardedFor:   "198.51.100.20, 192.0.2.20",
			forwardedProto: "https",
			wantIP:         "198.51.100.20",
			wantScheme:     "https",
		},
		{
			name:           "headers ignored from untrusted public peer",
			cidrs:          []string{"192.0.2.0/24"},
			remoteAddr:     "203.0.113.10:443",
			forwardedFor:   "198.51.100.20",
			forwardedProto: "https",
			wantIP:         "203.0.113.10",
			wantScheme:     "http",
		},
		{
			name:           "private peers are not trusted implicitly",
			cidrs:          []string{"192.0.2.0/24"},
			remoteAddr:     "10.0.0.10:443",
			forwardedFor:   "198.51.100.20",
			forwardedProto: "https",
			wantIP:         "10.0.0.10",
			wantScheme:     "http",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			if err := configureTrustedProxies(e, test.cidrs); err != nil {
				t.Fatalf("configureTrustedProxies() error = %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = test.remoteAddr
			req.Header.Set(echo.HeaderXForwardedFor, test.forwardedFor)
			req.Header.Set(echo.HeaderXForwardedProto, test.forwardedProto)
			ctx := e.NewContext(req, httptest.NewRecorder())

			if got := ctx.RealIP(); got != test.wantIP {
				t.Errorf("RealIP() = %q, want %q", got, test.wantIP)
			}
			if got := ctx.Scheme(); got != test.wantScheme {
				t.Errorf("Scheme() = %q, want %q", got, test.wantScheme)
			}
		})
	}
}

func TestConfigureTrustedProxiesRejectsInvalidCIDR(t *testing.T) {
	e := echo.New()

	err := configureTrustedProxies(e, []string{"not-a-cidr"})

	if err == nil {
		t.Fatal("configureTrustedProxies() error = nil, want invalid CIDR error")
	}
	if e.IPExtractor != nil || e.SchemeExtractor != nil {
		t.Fatal("configureTrustedProxies() changed extractors after invalid configuration")
	}
}
