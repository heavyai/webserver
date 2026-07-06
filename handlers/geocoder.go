// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
)

// Geocoder handles the search request.
func Geocoder(ctx echo.Context) error {
	query := ctx.QueryParam("location")
	if query == "" {
		return ctx.String(http.StatusBadRequest, "Missing 'location' query parameter")
	}

	for _, location := range config.Other.Locations {
		if strings.Contains(strings.ToLower(location.Location), strings.ToLower(query)) {
			return ctx.JSON(http.StatusOK, location)
		}
	}

	return ctx.String(http.StatusNotFound, "No matching location found")
}
