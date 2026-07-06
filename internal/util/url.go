// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"

	"github.com/heavyai/webserver/internal/config"
)

func isNotKnownRootSegment(segment string) bool {
	immerseRootURLSegments := []string{
		"login", "dashboards", "dashboard", "sql-editor", "data-manager", "tables",
		"importer",
		"control-panel",
		"logged-out",
		strings.Trim(config.Jupyter.JupyterPrefix, "/"),
	}
	for i := range immerseRootURLSegments {
		if segment == immerseRootURLSegments[i] && !(segment == "saml-post") {
			return false
		}
	}
	return true
}

// URLPathExtractDBContext - URLPathExtractDBContext method
func URLPathExtractDBContext(path string) (dbName string) {
	if len(path) == 0 || path == "/" {
		return
	}
	pathSegments := strings.Split(path, "/")
	if len(pathSegments) > 1 && isNotKnownRootSegment(pathSegments[1]) {
		dbName = pathSegments[1]
	}
	return
}
