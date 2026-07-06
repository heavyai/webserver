// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/labstack/echo/v4/middleware"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
)

var (
	// CustomLoggerConfig - Custom logger configuration
	CustomLoggerConfig = middleware.LoggerConfig{
		Format:           constants.LoggerMessageFormat,
		CustomTimeFormat: constants.LoggerTimeFormat,
		Output:           config.Other.AccessLog,
	}
	// CustomCompressConfig - Custom compression configuration
	CustomCompressConfig = middleware.GzipConfig{
		Skipper: SkipCompress,
		// Compression level defaults to Gzip DefaultCompression = -1
	}
)
