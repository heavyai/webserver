// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/util"
)

// JupyterLogin - Jupyter Login Middleware
func JupyterLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		// Set session ID header for the Jupyter Hub HeavyDB authenticator
		sessionID, err := util.GetSessionID(ctx)
		if err != nil || sessionID == "" {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get session ID from request", err)
		}

		// Use previous branded session ID until Jupyter integration app is updated
		ctx.Request().Header["X-OmniSci-SessionID"] = []string{string(sessionID)}

		newnotebook := ctx.QueryParam("newnotebook")
		if len(newnotebook) > 0 {
			// Set cookies on the client indicating it wants to make a new HeavyDB notebook on login
			newNotebookCookie := &http.Cookie{
				Name:     "newnotebook",
				Value:    "true",
				Path:     "/",
				Expires:  time.Now().Add(5 * time.Minute),
				HttpOnly: true,
			}
			ctx.SetCookie(newNotebookCookie)

			sqlValue := ctx.QueryParam("sql")
			if len(sqlValue) > 0 {
				escapedSQL := url.QueryEscape(sqlValue)

				newNotebookSQLCookie := &http.Cookie{
					Name:     "newnotebooksql",
					Value:    escapedSQL,
					Path:     "/",
					Expires:  time.Now().Add(5 * time.Minute),
					HttpOnly: true,
				}
				ctx.SetCookie(newNotebookSQLCookie)
			}
		}
		return next(ctx)
	}
}
