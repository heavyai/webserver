// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Noop - No Op Handler
func Noop(ctx echo.Context) (err error) {
	ctx.String(
		http.StatusNotImplemented,
		"No op handler should never be reached!",
	)

	return err
}
