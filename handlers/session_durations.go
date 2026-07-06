// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/models"
)

// SessionDurations - session durations handler method
func SessionDurations(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, config.Session.Durations)
}

type sessionDurationsCurrentResponse struct {
	DBName       string `json:"dbName"`
	IdleDuration int64  `json:"idleDuration"`
	MaxDuration  int64  `json:"maxDuration"`
}

// SessionDurationsCurrent - SessionDurationsCurrent handler
func SessionDurationsCurrent(ctx echo.Context) error {
	authCookie, err := ctx.Cookie(config.Other.CookieHeavyDBAuth)
	if err != nil || authCookie.Value == "" {
		ctx.String(http.StatusUnauthorized, "Not Authenticated: no auth token cookie read")
		return err
	}

	token := models.AuthToken(authCookie.Value)
	_, sessionInfoRef, err := token.Validate(ctx)
	if err != nil {
		ctx.String(http.StatusUnauthorized, err.Error())
		return err
	}

	idleDuration := (sessionInfoRef.IdleTimeout.Unix() - time.Now().Unix()) * 1000
	maxDuration := (sessionInfoRef.MaxTimeout.Unix() - time.Now().Unix()) * 1000

	return ctx.JSON(http.StatusOK, sessionDurationsCurrentResponse{
		DBName:       sessionInfoRef.DBName,
		IdleDuration: idleDuration,
		MaxDuration:  maxDuration,
	})
}
