// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"fmt"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/services"
)

// AWSRegion - type
type AWSRegion string

// AWS Regions
const (
	UsEast1               AWSRegion = "us-east-1"
	UsEast2                         = "us-east-2"
	UsWest1                         = "us-west-1"
	UsWest2                         = "us-west-2"
	CanadaCentral1                  = "ca-central-1"
	AsiaPacificSouth1               = "ap-south-1"
	AsiaPacificNortheast1           = "ap-northeast-1"
	AsiaPacificNortheast2           = "ap-northeast-2"
	AsiaPacificNortheast3           = "ap-northeast-3"
	AsiaPacificSoutheast1           = "ap-southeast-1"
	AsiaPacificSoutheast2           = "ap-southeast-2"
	ChinaNorth1                     = "cn-north-1"
	ChinaNorthwest1                 = "cn-northwest-1"
	EuCentral1                      = "eu-central-1"
	EuWest1                         = "eu-west-1"
	EuWest2                         = "eu-west-2"
	EuWest3                         = "eu-west-3"
	SouthAmericaEast1               = "sa-east-1"
)

// S3AccessType - type
type S3AccessType string

// S3 Access types
const (
	S3AccessTypeDirect S3AccessType = "S3_DIRECT"
	S3AccessTypeSelect              = "S3_SELECT"
)

// S3FSIConfig - type
type S3FSIConfig struct {
	TableName     string                  `json:"tableName"`
	Bucket        string                  `json:"bucket"`
	BasePath      string                  `json:"basePath"`
	AwsRegion     AWSRegion               `json:"awsRegion"`
	AccessType    S3AccessType            `json:"accessType"`
	IsParquet     bool                    `json:"isParquet"`
	AccessID      string                  `json:"accessId"`
	SecretKey     string                  `json:"secretKey"`
	RowDescriptor heavy.TRowDescriptor    `json:"rowDescriptor"`
	CopyParams    heavy.TCopyParams       `json:"copyParams"`
	RefreshInfo   heavy.TTableRefreshInfo `json:"refreshInfo"`
}

// GetTableName - getter
func (fsiConfig S3FSIConfig) GetTableName() string {
	return fsiConfig.TableName
}

// GetCopyParams - getter
func (fsiConfig S3FSIConfig) GetCopyParams() heavy.TCopyParams {
	return fsiConfig.CopyParams
}

// GetAccessType - getter
func (fsiConfig S3FSIConfig) GetAccessType() S3AccessType {
	return fsiConfig.AccessType
}

// GetAccessID - getter
func (fsiConfig S3FSIConfig) GetAccessID() string {
	return fsiConfig.AccessID
}

// GetSecretKey - getter
func (fsiConfig S3FSIConfig) GetSecretKey() string {
	return fsiConfig.SecretKey
}

// GetRefreshInfo - getter
func (fsiConfig S3FSIConfig) GetRefreshInfo() heavy.TTableRefreshInfo {
	return fsiConfig.RefreshInfo
}

// GetIsParquet - getter
func (fsiConfig S3FSIConfig) GetIsParquet() bool {
	return fsiConfig.IsParquet
}

// GetRowDescriptor - getter
func (fsiConfig S3FSIConfig) GetRowDescriptor() heavy.TRowDescriptor {
	return fsiConfig.RowDescriptor
}

// BuildCreateTableWithClause - util function
func (fsiConfig S3FSIConfig) BuildCreateTableWithClause() string {
	var conf FSIConfig
	conf = &fsiConfig
	return buildWithClauseString(conf, fsiConfig.BasePath)
}

// BuildCreateForeignTableQuery - util function
func (fsiConfig S3FSIConfig) BuildCreateForeignTableQuery(serverName string) string {
	var conf FSIConfig
	conf = &fsiConfig
	return buildCreateForeignTableQuery(conf, serverName)
}

// CreateS3FSIServer - method
func (fsiConfig S3FSIConfig) CreateS3FSIServer(
	sessionID heavy.TSessionId,
	serverName string,
) (err error) {
	return createS3FSIServer(
		sessionID,
		serverName,
		fsiConfig.CopyParams.SourceType,
		fsiConfig.Bucket,
		fsiConfig.AwsRegion,
	)
}

// CreateS3FSIUserMapping - method
func (fsiConfig S3FSIConfig) CreateS3FSIUserMapping(
	sessionID heavy.TSessionId,
	serverName string,
) (err error) {
	createUserMappingQuery := fmt.Sprintf(
		`CREATE USER MAPPING FOR PUBLIC SERVER %s WITH (s3_access_key = '%s', s3_secret_key = '%s')`,
		serverName,
		fsiConfig.AccessID,
		fsiConfig.SecretKey,
	)
	if config.Enable.Verbose == true {
		config.Log.Info("CREATE USER MAPPING query: ", createUserMappingQuery)
	}
	_, err = services.SQLExecute(sessionID, createUserMappingQuery)
	return
}

// CreateForeignTable - service util
func (fsiConfig S3FSIConfig) CreateForeignTable(
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
