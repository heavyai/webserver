// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/constants"
)

// ImmerseReqContext - ImmerseReqContext type
type ImmerseReqContext struct {
	echo.Context
	Claims *ImmerseJWTClaims
	// Fingerprint is a UUID for each allowing for better debugging and logging for concurrent requests
	Fingerprint   uuid.UUID
	UsesTokenAuth bool
	RequestTime   time.Time
	DBName        string
}

// GetDBNameFromParams - GetDBNameFromParams util
func (immerseCtx ImmerseReqContext) GetDBNameFromParams() (dbName string) {
	dbName = immerseCtx.Context.Param(constants.DBNameRouteParamName)
	if len(dbName) == 0 {
		dbName = immerseCtx.DBName
	}
	return
}
