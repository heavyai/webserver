// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"

	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

// ImmerseJWTClaims - ImmerseJWTClaims type
type ImmerseJWTClaims struct {
	*jwt.Claims
	SessionsInfoMap SessionsInfoMap `json:"sessionsInfoMap,omitempty"`
	Username        string          `json:"username,omitempty"`
	MaxExpiry       jwt.NumericDate
}

// AddSessionToToken - AddSessionToToken method
func (claims *ImmerseJWTClaims) AddSessionToToken(username string, sessionID heavy.TSessionId, DBName string, idleExpiry time.Time, maxExpiry time.Time) (token AuthToken, err error) {
	claims.SessionsInfoMap[DBName] = &DBSessionInfo{
		Username:    username,
		SessionID:   sessionID,
		IdleTimeout: idleExpiry,
		MaxTimeout:  maxExpiry,
		DBName:      DBName,
	}

	return NewAuthToken(AuthTokenConfig{
		SessionsInfoMap:  claims.SessionsInfoMap,
		Username:         claims.Username,
		MaxSessionExpiry: maxExpiry,
	})
}
