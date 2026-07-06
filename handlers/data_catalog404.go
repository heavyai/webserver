// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// DataCatalog404 - Data Catalog 404 Handler
// Returns a 404 for every request to Data Catalog when it is not configured
func DataCatalog404(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotFound)
}
