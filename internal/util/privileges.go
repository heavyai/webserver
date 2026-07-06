// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/models"
	"github.com/heavyai/webserver/services"
)

// Privileges - Privileges struct
type Privileges struct {
	Dashboard     []*heavy.TDBObject `json:"dashboard"`
	ViewSQLEditor bool               `json:"viewSQLEditor"`
	DataSource    []*heavy.TDBObject `json:"dataSource"`
}

// GetDBPrivileges - GetDBPrivileges util method
func GetDBPrivileges(sessionInfo models.DBSessionInfo) (privileges Privileges, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	privileges.Dashboard, err = thriftClient.Client.GetDbObjectPrivs(thriftClient.Ctx, sessionInfo.SessionID, "", heavy.TDBObjectType_DashboardDBObjectType)
	if err != nil {
		config.Log.Error("Could not fetch dashboard privileges: ", err)
		return
	}
	privileges.ViewSQLEditor, err = thriftClient.Client.HasObjectPrivilege(
		thriftClient.Ctx,
		sessionInfo.SessionID,
		sessionInfo.Username,
		sessionInfo.DBName,
		heavy.TDBObjectType_DatabaseDBObjectType,
		&heavy.TDBObjectPermissions{
			DatabasePermissions_: &heavy.TDatabasePermissions{
				ViewSqlEditor_: true,
			},
		},
	)
	if err != nil {
		config.Log.Error("Could not fetch SQL Editor privileges: ", err)
		return
	}
	privileges.DataSource, err = thriftClient.Client.GetDbObjectPrivs(
		thriftClient.Ctx,
		sessionInfo.SessionID,
		"",
		heavy.TDBObjectType_TableDBObjectType,
	)
	if err != nil {
		config.Log.Error("Could not fetch table privileges: ", err)
	}

	return
}

// GetTablePrivileges - GetTablePrivileges util
func GetTablePrivileges(sessionID heavy.TSessionId, tableName string) (privs []*heavy.TDBObject, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	return thriftClient.Client.GetDbObjectPrivs(
		thriftClient.Ctx,
		sessionID,
		tableName,
		heavy.TDBObjectType_TableDBObjectType,
	)
}
