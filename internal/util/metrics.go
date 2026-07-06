// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
)

type metricsConfigVars struct {
	GoogleTagID string `json:"googleTagID"`
}

// DeleteCookie - Delete Cookie utility function
func DeleteCookie(ctx echo.Context, name string) {
	ctx.SetCookie(&http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: true,
	})
}

// GetMetricsConfigJSON - GetMetricsConfigJSON util
func GetMetricsConfigJSON() (mcvJSON []byte, err error) {
	var googleTagID = ""

	if config.Enable.GoogleMetrics {
		googleTagID = config.Other.GoogleTagID
	}

	mcv := &metricsConfigVars{
		googleTagID,
	}

	return json.MarshalIndent(mcv, "", "  ")
}
