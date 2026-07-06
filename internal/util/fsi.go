// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"

	"github.com/heavyai/webserver/services"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/models"
)

// RenameForeignTable - util
func RenameForeignTable(sessionID heavy.TSessionId, tableName string, newTableName string) (err error) {
	serverName := models.GetFsiServerName(tableName)
	newServerName := models.GetFsiServerName(newTableName)
	renameServerQuery := models.BuildRenameServerQuery(
		serverName,
		newServerName,
	)
	renameTableQuery := models.BuildRenameForeignTableQuery(
		tableName,
		newTableName,
	)
	if config.Enable.Verbose == true {
		config.Log.Info("Rename SERVER rename query: ", renameServerQuery)
		config.Log.Info("Rename FOREIGN TABLE rename query: ", renameTableQuery)
	}
	_, err = services.SQLExecute(sessionID, renameServerQuery)
	if err != nil {
		config.Log.Error("Could not rename server: ", err)
		return
	}
	_, err = services.SQLExecute(sessionID, renameTableQuery)
	if err != nil {
		config.Log.Error("Could not rename table: ", err)
		_, err = services.SQLExecute(sessionID, models.BuildRenameServerQuery(newServerName, serverName))
		if err != nil {
			config.Log.Error("Could not revert server rename: ", err)
		}
	}
	return
}

// RenameForeignTableColumns - util
func RenameForeignTableColumns(
	sessionID heavy.TSessionId,
	tableName string,
	renameMaps *[]models.ColumnRenameMap) (errors []models.QueryError) {
	for _, columnRenameMap := range *renameMaps {
		query := fmt.Sprintf(
			`ALTER FOREIGN TABLE %s RENAME COLUMN %s TO %s`,
			tableName,
			columnRenameMap.ColumnName,
			columnRenameMap.NewColumnName,
		)
		if config.Enable.Verbose == true {
			config.Log.Info("Rename FOREIGN TABLE query: ", query)
		}
		_, err := services.SQLExecute(sessionID, query)
		if err != nil {
			errors = append(errors, models.QueryError{
				Error: err,
				Query: query,
			})
		}
	}

	return
}

// UpdateForeignTableRefreshSchedule - util
func UpdateForeignTableRefreshSchedule(
	sessionID heavy.TSessionId,
	tableName string,
	refreshConfig heavy.TTableRefreshInfo,
) error {
	query := models.BuildUpdateForeignTableRefreshScheduleQuery(tableName, refreshConfig)
	if config.Enable.Verbose == true {
		config.Log.Info("Update refresh schedule query: ", query)
	}
	_, err := services.SQLExecute(sessionID, query)
	return err
}

// RefreshForeignTable - util
func RefreshForeignTable(
	sessionID heavy.TSessionId,
	tableName string,
) (tableDetails *heavy.TTableDetails, err error) {
	query := fmt.Sprintf(`REFRESH FOREIGN TABLES %s`, tableName)
	if config.Enable.Verbose == true {
		config.Log.Info("Refresh table query: ", query)
	}
	_, err = services.SQLExecute(sessionID, query)
	if err != nil {
		return
	}
	return GetTableDetails(sessionID, tableName)
}
