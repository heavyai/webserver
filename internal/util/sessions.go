// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/services"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/models"
)

// DestroyDBSession - DestroyDBSession util
func DestroyDBSession(sessionID heavy.TSessionId) (err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.Disconnect(
		thriftClient.Ctx,
		sessionID,
	)
}

// GetDBSessionInfo - GetDBSessionInfo util method
func GetDBSessionInfo(claims *models.ImmerseJWTClaims, DBName string) (sessionInfo *heavy.TSessionInfo, err error) {
	dbSessionInfo, err := claims.SessionsInfoMap.GetDBSessionInfo(DBName)
	if err != nil {
		return
	}
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.GetSessionInfo(
		thriftClient.Ctx,
		dbSessionInfo.SessionID,
	)
}

// CloneDBSession - CloneDBSession util method
func CloneDBSession(sessionID heavy.TSessionId) (newSessionID heavy.TSessionId, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.CloneSession(thriftClient.Ctx, sessionID)
}

// GetDBAccessList - GetDBAccessList util method
func GetDBAccessList(sessionID heavy.TSessionId) (dbAccessList []*heavy.TDBInfo, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.GetDatabases(thriftClient.Ctx, sessionID)
}

// SwitchDBSession - SwitchDBSession util method
func SwitchDBSession(sessionID heavy.TSessionId, DBName string) (err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.SwitchDatabase(thriftClient.Ctx, sessionID, DBName)
}

// GetThriftSessionInfo - GetThriftSessionInfo util method
func GetThriftSessionInfo(sessionID heavy.TSessionId) (thriftSessionInfo *heavy.TSessionInfo, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.GetSessionInfo(thriftClient.Ctx, sessionID)
}

// CreateDBSession - convenience method help w/ DRY when connecting to HeavyDB
func CreateDBSession(creds models.AuthCredentials) (sessionID heavy.TSessionId, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	sessionID, err = thriftClient.Client.Connect(
		thriftClient.Ctx,
		creds.Username,
		creds.Password,
		creds.DBName,
	)
	return
}

// ExchangeSessionID - ExchangeSessionID util method
func ExchangeSessionID(sessionID heavy.TSessionId, dbName string) (newSessionID heavy.TSessionId, thriftSessionInfo *heavy.TSessionInfo, err error) {
	newSessionID, err = CloneDBSession(sessionID)
	if err != nil {
		config.Log.Error("Failed to clone HeavyDB session: ", err)
		return
	}
	err = SwitchDBSession(newSessionID, dbName)
	if err != nil {
		config.Log.Error("Failed to switch HeavyDB session: ", err)
		err = DestroyDBSession(newSessionID)
		if err != nil {
			config.Log.Error("Could not destroy cloned HeavyDB session ID: ", newSessionID, err)
		}
		newSessionID = ""
	} else {
		thriftSessionInfo, err = GetThriftSessionInfo(newSessionID)
		if err != nil {
			config.Log.Error("Could not retrieve thriftSessionInfo for new sessionId: ", err)
			return
		}
	}
	return
}

// InaccessibleDB - util method
func InaccessibleDB(sessionInfo *models.DBSessionInfo, name string) bool {
	thriftDBAccessList, err := GetDBAccessList(sessionInfo.SessionID)
	if err == nil {
		for _, db := range thriftDBAccessList {
			if db.DbName == name {
				return false
			}
		}
	}

	return true
}

// GetSessionID - util method
func GetSessionID(ctx echo.Context) (sessionID heavy.TSessionId, err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	dbSession, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		return
	}

	return dbSession.SessionID, nil
}
