// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"

	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

// DBSessionInfo - DBSessionInfo struct
type DBSessionInfo struct {
	Username    string
	SessionID   heavy.TSessionId `json:"sessionId,omitempty"`
	IdleTimeout time.Time
	MaxTimeout  time.Time
	DBName      string
}

// IsExpired - IsExpired method
func (dbSessionInfo DBSessionInfo) IsExpired() bool {
	idleExpiry := *jwt.NewNumericDate(dbSessionInfo.IdleTimeout.UTC())
	maxSessionExpiry := *jwt.NewNumericDate(dbSessionInfo.MaxTimeout.UTC())
	now := *jwt.NewNumericDate(time.Now())
	return now >= idleExpiry || now >= maxSessionExpiry
}
