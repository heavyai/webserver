// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/middlewares"
)

// Immerse404 - Immerse 404 Handler
func Immerse404(err error, ctx echo.Context) {
	code := http.StatusInternalServerError
	if httpErr, ok := err.(*echo.HTTPError); ok {
		code = httpErr.Code
	}

	if code == http.StatusNotFound {
		// If we get a request for a route that doesn't exist in the API, if could be a frontend app route.
		// So go ahead and serve up the HTML for the frontend app.
		// TODO: Figure out how to properly handle frontend app 404's
		if err := middlewares.ExtendContext(middlewares.AppIndexSessionExchange(AppIndex))(ctx); err != nil {
			config.Log.Error(err)
			ctx.Echo().DefaultHTTPErrorHandler(err, ctx)
		}
	} else {
		ctx.Echo().DefaultHTTPErrorHandler(err, ctx)
	}
}
