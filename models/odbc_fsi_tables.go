// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/services"
)

// OdbcFSIConfig - type
type OdbcFSIConfig struct {
	TableName     string                  `json:"tableName"`
	Driver        string                  `json:"driver"`
	DBName        string                  `json:"dbName"`
	DBHost        string                  `json:"dbHost"`
	Port          int                     `json:"port"`
	SQLSelect     string                  `json:"sqlSelect"`
	SQLOrderBy    string                  `json:"sqlOrderBy"`
	Username      string                  `json:"username"`
	Password      string                  `json:"password"`
	RowDescriptor heavy.TRowDescriptor    `json:"rowDescriptor"`
	RefreshInfo   heavy.TTableRefreshInfo `json:"refreshInfo"`
	DataWarehouse string                  `json:"dataWarehouse"`
	Role          string                  `json:"role"`
}

// DetectColumnTypesConfig - struct
type DetectColumnTypesConfig struct {
	FileName      string `json:"fileName"`
	SourceType    string `json:"sourceType"`
	DBUserName    string `json:"dbUserName"`
	DBPassword    string `json:"dbPassword"`
	Driver        string `json:"driver"`
	DBName        string `json:"dbName"`
	DBHost        string `json:"dbHost"`
	DBPort        int    `json:"dbPort"`
	SQLSelect     string `json:"sqlSelect"`
	SQLOrderBy    string `json:"sqlOrderBy"`
	DataWarehouse string `json:"dataWarehouse"`
	Role          string `json:"role"`
}

// GetTableName - getter
func (fsiConfig OdbcFSIConfig) GetTableName() string {
	return fsiConfig.TableName
}

// GetDriver - getter
func (fsiConfig OdbcFSIConfig) GetDriver() string {
	return fsiConfig.Driver
}

// GetDBName - getter
func (fsiConfig OdbcFSIConfig) GetDBName() string {
	return fsiConfig.DBName
}

// GetDBHost - getter
func (fsiConfig OdbcFSIConfig) GetDBHost() string {
	return fsiConfig.DBHost
}

// GetPort - getter
func (fsiConfig OdbcFSIConfig) GetPort() int {
	return fsiConfig.Port
}

// GetSQLSelect - getter
func (fsiConfig OdbcFSIConfig) GetSQLSelect() string {
	return fsiConfig.SQLSelect
}

// GetRefreshInfo - getter
func (fsiConfig OdbcFSIConfig) GetRefreshInfo() heavy.TTableRefreshInfo {
	return fsiConfig.RefreshInfo
}

// GetRowDescriptor - getter
func (fsiConfig OdbcFSIConfig) GetRowDescriptor() heavy.TRowDescriptor {
	return fsiConfig.RowDescriptor
}

// GetOdbcConnectionString - util
func GetOdbcConnectionString(
	driver string,
	DBName string,
	DBHost string,
	port int,
	dataWarehouse string,
	role string,
) string {
	var connectionStringBits []string
	if driver != "" {
		connectionStringBits = append(
			connectionStringBits,
			fmt.Sprintf(`Driver=%s`, driver),
		)
	}
	if DBName != "" {
		connectionStringBits = append(
			connectionStringBits,
			fmt.Sprintf(`Database=%s`, DBName),
		)
	}
	if DBHost != "" {
		connectionStringBits = append(
			connectionStringBits,
			fmt.Sprintf(`Server=%s`, DBHost),
		)
	}
	if port != 0 {
		connectionStringBits = append(
			connectionStringBits,
			fmt.Sprintf(
				`Port=%s`,
				strconv.Itoa(port),
			),
		)
	}
	if dataWarehouse != "" {
		connectionStringBits = append(
			connectionStringBits,
			fmt.Sprintf(`Warehouse=%s`, dataWarehouse),
		)
	}

	if role != "" {
		connectionStringBits = append(
			connectionStringBits,
			fmt.Sprintf(`Role=%s`, role),
		)
	}

	return strings.Join(connectionStringBits, ";")
}

// GetConnectionString - method
func (fsiConfig OdbcFSIConfig) GetConnectionString() string {
	return GetOdbcConnectionString(
		fsiConfig.Driver,
		fsiConfig.DBName,
		fsiConfig.DBHost,
		fsiConfig.Port,
		fsiConfig.DataWarehouse,
		fsiConfig.Role,
	)
}

// CreateOdbcFSIServer - method
func (fsiConfig OdbcFSIConfig) CreateOdbcFSIServer(
	sessionID heavy.TSessionId,
	serverName string,
) (err error) {
	return createOdbcFSIServer(
		sessionID,
		serverName,
		fsiConfig,
	)
}

// CreateOdbcCredentialString - util
func CreateOdbcCredentialString(username string, password string) string {
	return fmt.Sprintf(
		`UID=%s;PWD=%s`,
		username,
		password,
	)
}

// CreateOdbcFSIUserMapping - method
func (fsiConfig OdbcFSIConfig) CreateOdbcFSIUserMapping(
	sessionID heavy.TSessionId,
	serverName string,
) (err error) {
	createUserMappingQuery := fmt.Sprintf(
		`CREATE USER MAPPING FOR PUBLIC SERVER %s WITH (credential_string = '%s')`,
		serverName,
		CreateOdbcCredentialString(
			fsiConfig.Username,
			fsiConfig.Password,
		),
	)
	if config.Enable.Verbose == true {
		config.Log.Info("CREATE USER MAPPING query: ", createUserMappingQuery)
	}
	_, err = services.SQLExecute(sessionID, createUserMappingQuery)
	return
}

func buildOdbcWithClauseString(
	fsiConfig OdbcFSIConfig,
) string {
	withParamStrings := []string{
		fmt.Sprintf(
			`sql_select = '%s',sql_order_by = '%s'`,
			fsiConfig.SQLSelect,
			fsiConfig.SQLOrderBy,
		),
	}

	withParamStrings = append(
		withParamStrings,
		BuildForeignTableRefreshScheduleQueryArguments(fsiConfig.GetRefreshInfo()),
	)

	return strings.Join(withParamStrings, ",")
}

func buildOdbcCreateForeignTableQuery(
	fsiConfig OdbcFSIConfig,
	serverName string,
) string {
	var builder strings.Builder
	builder.WriteString(
		fmt.Sprintf(
			`CREATE FOREIGN TABLE %s (%s)`,
			fsiConfig.GetTableName(),
			BuildColumnsDefinitionsString(fsiConfig.GetRowDescriptor()),
		),
	)

	if serverName != "" {
		builder.WriteString(
			fmt.Sprintf(
				` SERVER %s`,
				serverName,
			),
		)
	}

	builder.WriteString(
		fmt.Sprintf(
			` WITH(%s)`,
			buildOdbcWithClauseString(fsiConfig),
		),
	)
	return builder.String()
}

func createOdbcForeignTable(
	sessionID heavy.TSessionId,
	fsiConfig OdbcFSIConfig,
	serverName string,
) (err error) {
	createForeignTableQuery := buildOdbcCreateForeignTableQuery(
		fsiConfig,
		serverName,
	)
	if config.Enable.Verbose == true {
		config.Log.Info("CREATE FOREIGN TABLE query: ", createForeignTableQuery)
	}
	_, err = services.SQLExecute(sessionID, createForeignTableQuery)
	if err != nil {
		if serverName != "" {
			_ = DropServer(sessionID, serverName)
		}
		_ = DropForeignTable(sessionID, fsiConfig.GetTableName())
	}
	return
}

// CreateForeignTable - service util
func (fsiConfig OdbcFSIConfig) CreateForeignTable(
	sessionID heavy.TSessionId,
	serverName string,
) (err error) {
	return createOdbcForeignTable(
		sessionID,
		fsiConfig,
		serverName,
	)
}
