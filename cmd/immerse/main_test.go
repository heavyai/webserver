// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestStaticRoutesServeFrontendAndDocs(t *testing.T) {
	root := t.TempDir()
	frontend := filepath.Join(root, "frontend")
	docs := filepath.Join(root, "docs")

	writeTestFile(t, filepath.Join(frontend, "assets", "app.js"), "frontend asset")
	writeTestFile(t, filepath.Join(frontend, "docs", "guide.html"), "wrong docs root")
	writeTestFile(t, filepath.Join(docs, "guide.html"), "documentation")

	e := echo.New()
	configureStaticRoutes(e, frontend, docs)

	for _, testCase := range []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "nested frontend asset",
			path:     "/assets/app.js",
			expected: "frontend asset",
		},
		{
			name:     "docs route takes precedence over frontend wildcard",
			path:     "/docs/guide.html",
			expected: "documentation",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
			}
			if response.Body.String() != testCase.expected {
				t.Fatalf("expected body %q, got %q", testCase.expected, response.Body.String())
			}
		})
	}
}

func TestStaticRoutesRejectEncodedPathTraversal(t *testing.T) {
	root := t.TempDir()
	frontend := filepath.Join(root, "frontend")
	docs := filepath.Join(root, "docs")

	writeTestFile(t, filepath.Join(frontend, "index.html"), "frontend")
	writeTestFile(t, filepath.Join(root, "secret.txt"), "must not be served")

	e := echo.New()
	configureStaticRoutes(e, frontend, docs)

	request := httptest.NewRequest(http.MethodGet, "/%2e%2e/secret.txt", nil)
	response := httptest.NewRecorder()

	e.ServeHTTP(response, request)

	if response.Code == http.StatusOK {
		t.Fatalf("expected traversal request to be rejected, got body %q", response.Body.String())
	}
	if strings.Contains(response.Body.String(), "must not be served") {
		t.Fatal("static route served a file outside the configured frontend root")
	}
}

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
}
