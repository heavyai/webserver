// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
)

// JupyterDesktop - Jupyter Desktop Handler
func JupyterDesktop(ctx echo.Context) error {
	resp, err := http.Get(config.Jupyter.DesktopURL.String() + "?token=" + constants.JupyterDesktopToken)
	if err != nil {
		return ctx.NoContent(http.StatusServiceUnavailable)
	}

	return ctx.NoContent(resp.StatusCode)
}
