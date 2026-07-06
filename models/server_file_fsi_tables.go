// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

// ServerFileConfig - type
type ServerFileConfig struct {
	TableName     string                  `json:"tableName"`
	Path          string                  `json:"path"`
	RowDescriptor heavy.TRowDescriptor    `json:"rowDescriptor"`
	CopyParams    heavy.TCopyParams       `json:"copyParams"`
	RefreshInfo   heavy.TTableRefreshInfo `json:"refreshInfo"`
}

// CreateForeignTable - service util
func (fsiConfig ServerFileConfig) CreateForeignTable(
	sessionID heavy.TSessionId,
	serverName string,
) (err error) {
	var conf FSIConfig
	conf = &fsiConfig
	return createForeignTable(
		sessionID,
		conf,
		serverName,
	)
}

// CreateS3FSIServer - method
func (fsiConfig ServerFileConfig) CreateS3FSIServer(
	sessionID heavy.TSessionId,
	serverName string,
) (err error) {
	return createS3FSIServer(
		sessionID,
		serverName,
		fsiConfig.CopyParams.SourceType,
		"",
		"",
	)
}

// GetTableName - getter
func (fsiConfig ServerFileConfig) GetTableName() string {
	return fsiConfig.TableName
}

// GetCopyParams - getter
func (fsiConfig ServerFileConfig) GetCopyParams() heavy.TCopyParams {
	return fsiConfig.CopyParams
}

// GetRefreshInfo - getter
func (fsiConfig ServerFileConfig) GetRefreshInfo() heavy.TTableRefreshInfo {
	return fsiConfig.RefreshInfo
}

// GetRowDescriptor - getter
func (fsiConfig ServerFileConfig) GetRowDescriptor() heavy.TRowDescriptor {
	return fsiConfig.RowDescriptor
}

// BuildCreateForeignTableQuery - util function
func (fsiConfig ServerFileConfig) BuildCreateForeignTableQuery(serverName string) string {
	var conf FSIConfig
	conf = &fsiConfig
	return buildCreateForeignTableQuery(conf, serverName)
}

// BuildCreateTableWithClause - util function
func (fsiConfig ServerFileConfig) BuildCreateTableWithClause() string {
	var conf FSIConfig
	conf = &fsiConfig
	return buildWithClauseString(conf, fsiConfig.Path)
}

// NoOp methods to appease the FSIConfig interface

// CreateS3FSIUserMapping - noop
func (fsiConfig ServerFileConfig) CreateS3FSIUserMapping(sessionID heavy.TSessionId, serverName string) (err error) {
	return
}

// GetAccessType - noop
func (fsiConfig ServerFileConfig) GetAccessType() S3AccessType {
	return ""
}

// GetAccessID - noop
func (fsiConfig ServerFileConfig) GetAccessID() string {
	return ""
}

// GetSecretKey - noop
func (fsiConfig ServerFileConfig) GetSecretKey() string {
	return ""
}
