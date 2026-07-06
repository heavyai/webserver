// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"fmt"
	"strings"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/services"

	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/db_thrift_client/common"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

// FSIConfig - type
type FSIConfig interface {
	CreateS3FSIServer(heavy.TSessionId, string) error
	CreateForeignTable(heavy.TSessionId, string) error
	BuildCreateForeignTableQuery(string) string
	CreateS3FSIUserMapping(heavy.TSessionId, string) error
	BuildCreateTableWithClause() string
	GetTableName() string
	GetCopyParams() heavy.TCopyParams
	GetRowDescriptor() heavy.TRowDescriptor
	GetRefreshInfo() heavy.TTableRefreshInfo
	GetAccessType() S3AccessType
	GetAccessID() string
	GetSecretKey() string
}

var hackedEncodingTypeStringMap = map[common.TEncodingType]string{
	common.TEncodingType_GEOINT:       "COMPRESSED",
	common.TEncodingType_DATE_IN_DAYS: "DAYS",
}

// GetHackedEncodingTypeString - Util for overriding TEncodingTypes due to instructions in discussion referenced here: https://heavyai.atlassian.net/browse/FE-14771?focusedCommentId=71504
func GetHackedEncodingTypeString(encoding common.TEncodingType) string {
	var encodingTypeString = fmt.Sprintf(`%s`, encoding)
	if hackedEncodingTypeString, ok := hackedEncodingTypeStringMap[encoding]; ok == true {
		encodingTypeString = hackedEncodingTypeString
	}
	return encodingTypeString
}

// BuildEncodingString - util function
func BuildEncodingString(colType common.TTypeInfo) (encodingString string) {
	var builder strings.Builder
	fmt.Fprintf(&builder, ` ENCODING %s`, GetHackedEncodingTypeString(colType.Encoding))
	switch colType.Encoding {
	case common.TEncodingType_FIXED:
		fmt.Fprintf(&builder, `(%d)`, colType.Size)
		encodingString = builder.String()
	case common.TEncodingType_DICT:
		// Hack to deal w/ detect_column_types spitting out bad values for this field 🤷‍
		if colType.Size != 8 && colType.Size != 16 && colType.Size != 32 {
			fmt.Fprintf(&builder, `(%d)`, 32)
		} else {
			fmt.Fprintf(&builder, `(%d)`, colType.Size)
		}
		encodingString = builder.String()
	case common.TEncodingType_GEOINT:
		fmt.Fprintf(&builder, `(%d)`, colType.Size)
		encodingString = builder.String()
	case common.TEncodingType_DATE_IN_DAYS:
		fmt.Fprintf(&builder, `(%d)`, colType.Size)
		encodingString = builder.String()
	}
	return encodingString
}

var hackedDatumTypeStringMap = map[common.TDatumType]string{
	common.TDatumType_BOOL: "BOOLEAN",
	common.TDatumType_STR:  "TEXT",
}

// GetHackedDatumTypeString - Util for overriding TDatumTypes mismatches between heavyDB and thrift bindings
func GetHackedDatumTypeString(datumType common.TDatumType) string {
	var datumTypeString = fmt.Sprintf(`%s`, datumType)
	if hackedDatumTypeString, ok := hackedDatumTypeStringMap[datumType]; ok == true {
		datumTypeString = hackedDatumTypeString
	}
	return datumTypeString
}

// BuildColTypeString - util function
func BuildColTypeString(colType common.TTypeInfo) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`%s`, GetHackedDatumTypeString(colType.Type)))
	if colType.Type == common.TDatumType_DECIMAL {
		fmt.Fprintf(&builder, `(%d,%d)`, colType.Precision, colType.Scale)
	}
	if colType.IsArray == true {
		builder.WriteString("[]")
	}
	return builder.String()
}

var skipEncodingDatumTypes = []common.TDatumType{
	common.TDatumType_GEOMETRY,
	common.TDatumType_POINT,
	common.TDatumType_POLYGON,
	common.TDatumType_MULTIPOLYGON,
	common.TDatumType_LINESTRING,
}

func skipEncodingDatumTypesContains(datumType common.TDatumType, list []common.TDatumType) bool {
	for _, dType := range list {
		if dType == datumType {
			return true
		}
	}
	return false
}

// BuildColumnDefinitionString - util function
func BuildColumnDefinitionString(colName string, colType common.TTypeInfo) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`%s %s`, colName, BuildColTypeString(colType)))

	if !skipEncodingDatumTypesContains(colType.Type, skipEncodingDatumTypes) &&
		colType.Encoding != common.TEncodingType_NONE {
		fmt.Fprintf(&builder, `%s`, BuildEncodingString(colType))
	}
	return builder.String()
}

// BuildColumnsDefinitionsString - util function
func BuildColumnsDefinitionsString(rowDescriptor heavy.TRowDescriptor) string {
	var colDefstrings []string
	for _, colType := range rowDescriptor {
		colDefstrings = append(colDefstrings, BuildColumnDefinitionString(colType.ColName, *colType.ColType))
	}
	return strings.Join(colDefstrings, ",")
}

// BuildForeignTableRefreshScheduleQueryArguments - util
func BuildForeignTableRefreshScheduleQueryArguments(refreshInfo heavy.TTableRefreshInfo) string {
	refreshArgumentStrings := []string{
		fmt.Sprintf(`REFRESH_TIMING_TYPE='%s'`, refreshInfo.TimingType),
		fmt.Sprintf(`REFRESH_UPDATE_TYPE='%s'`, refreshInfo.UpdateType),
	}
	if refreshInfo.StartDateTime != "" {
		refreshArgumentStrings = append(
			refreshArgumentStrings,
			fmt.Sprintf(`REFRESH_START_DATE_TIME='%s'`, refreshInfo.StartDateTime),
		)
	}
	if refreshInfo.IntervalType != heavy.TTableRefreshIntervalType_NONE {
		intervalTypeString := fmt.Sprintf(`%s`, refreshInfo.IntervalType)
		refreshArgumentStrings = append(
			refreshArgumentStrings,
			fmt.Sprintf(
				`REFRESH_INTERVAL='%s'`,
				// Grabbing slice of first character of refreshInfo.IntervalType here because the query doesn't accept
				//  the full value of refreshInfo.IntervalType, but just the first character of the IntervalType.
				//  See: https://heavyai.gitbook.io/omnilink/-MQTgUutGCpgl8jKW_g2/reference#example-change-the-foreign-table-update-schedule
				fmt.Sprintf(`%d%s`, refreshInfo.IntervalCount, intervalTypeString[0:1]),
			),
		)
	}
	return strings.Join(refreshArgumentStrings, ",")
}

// GetFsiServerName - util function
func GetFsiServerName(tableName string) string {
	return fmt.Sprintf(`%s%s`, tableName, constants.FsiServerNameSuffix)
}

// ColumnRenameMap - util
type ColumnRenameMap struct {
	ColumnName    string `json:"columnName"`
	NewColumnName string `json:"newColumnName"`
}

// BuildRenameServerQuery - util
func BuildRenameServerQuery(serverName string, newServerName string) string {
	return fmt.Sprintf(
		`ALTER SERVER %s RENAME TO %s`,
		serverName,
		newServerName,
	)
}

// BuildRenameForeignTableQuery - util
func BuildRenameForeignTableQuery(tableName string, newTableName string) string {
	return fmt.Sprintf(
		`ALTER FOREIGN TABLE %s RENAME TO %s`,
		tableName,
		newTableName,
	)
}

// QueryError - type
type QueryError struct {
	Error error
	Query string
}

// BuildUpdateForeignTableRefreshScheduleQuery - util
func BuildUpdateForeignTableRefreshScheduleQuery(
	tableName string,
	refreshInfo heavy.TTableRefreshInfo,
) string {
	return fmt.Sprintf(`ALTER FOREIGN TABLE %s SET(%s)`,
		tableName,
		BuildForeignTableRefreshScheduleQueryArguments(refreshInfo),
	)
}

// DropServer - service util
func DropServer(sessionID heavy.TSessionId, serverName string) error {
	query := fmt.Sprintf(`DROP SERVER IF EXISTS %s`, serverName)
	if config.Enable.Verbose == true {
		config.Log.Info("DROP SERVER query: ", query)
	}
	_, err := services.SQLExecute(sessionID, query)
	return err
}

// DropForeignTable - service util
func DropForeignTable(sessionID heavy.TSessionId, tableName string) error {
	query := fmt.Sprintf(`DROP FOREIGN TABLE IF EXISTS %s`, tableName)
	if config.Enable.Verbose == true {
		config.Log.Info("DROP FOREIGN TABLE query: ", query)
	}
	_, err := services.SQLExecute(sessionID, query)
	return err
}

func createForeignTable(
	sessionID heavy.TSessionId,
	fsiConfig FSIConfig,
	serverName string,
) (err error) {
	createForeignTableQuery := fsiConfig.BuildCreateForeignTableQuery(serverName)
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

func buildCreateForeignTableQuery(
	fsiConfig FSIConfig,
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
			fsiConfig.BuildCreateTableWithClause(),
		),
	)
	return builder.String()
}

func parseDelimiter(
	delimiter string,
) (parsed string) {
	// Because apparently this is the only way to get the literal string "\t" in Go 🤦‍
	if delimiter == "\t" {
		parsed = `\t`
	} else if delimiter == "\n" {
		parsed = `\n`
	} else {
		parsed = delimiter
	}
	return parsed
}

func buildWithClauseString(
	fsiConfig FSIConfig,
	filePath string,
) string {
	delimiter := parseDelimiter(fsiConfig.GetCopyParams().Delimiter)

	withParamStrings := []string{
		fmt.Sprintf(`file_path='%s'`, filePath),
	}
	if fsiConfig.GetCopyParams().SourceType != heavy.TSourceType_PARQUET_FILE {
		withParamStrings = append(
			withParamStrings,
			fmt.Sprintf(`delimiter='%s'`, delimiter),
		)
	}
	withParamStrings = append(
		withParamStrings,
		BuildForeignTableRefreshScheduleQueryArguments(fsiConfig.GetRefreshInfo()),
	)
	if fsiConfig.GetAccessType() != "" {
		withParamStrings = append(
			withParamStrings,
			fmt.Sprintf(`S3_ACCESS_TYPE='%s'`, fsiConfig.GetAccessType()),
		)
	}

	return strings.Join(withParamStrings, ",")
}

func buildS3CreateServerWithClause(
	bucket string,
	region AWSRegion,
) (withClause string) {
	withClauseBits := []string{}
	if bucket != "" || region != "" {
		withClauseBits = append(
			withClauseBits,
			"storage_type='AWS_S3'",
		)
		if bucket != "" {
			withClauseBits = append(
				withClauseBits,
				fmt.Sprintf(
					`s3_bucket='%s'`,
					bucket,
				),
			)
		}
		if region != "" {
			withClauseBits = append(
				withClauseBits,
				fmt.Sprintf(
					`aws_region='%s'`,
					region,
				),
			)
		}
	} else {
		withClauseBits = append(
			withClauseBits,
			"storage_type='LOCAL_FILE'",
		)
	}

	return fmt.Sprintf(
		` WITH (%s)`,
		strings.Join(withClauseBits, ","),
	)
}

func createOdbcFSIServer(
	sessionID heavy.TSessionId,
	serverName string,
	odbcFSIConfig OdbcFSIConfig,
) (err error) {
	builder := strings.Builder{}
	builder.WriteString(
		fmt.Sprintf(
			`CREATE SERVER %s FOREIGN DATA WRAPPER %s`,
			serverName,
			heavy.TSourceType_ODBC,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			` WITH(connection_string = '%s')`,
			odbcFSIConfig.GetConnectionString(),
		),
	)

	createServerQuery := builder.String()

	if config.Enable.Verbose == true {
		config.Log.Info("CREATE SERVER query: ", createServerQuery)
	}
	_, err = services.SQLExecute(sessionID, createServerQuery)
	return
}

func createS3FSIServer(
	sessionID heavy.TSessionId,
	serverName string,
	sourceType heavy.TSourceType,
	bucket string,
	region AWSRegion,
) (err error) {
	dataWrapperType := heavy.TSourceType_DELIMITED_FILE.String()
	if sourceType == heavy.TSourceType_PARQUET_FILE {
		dataWrapperType = heavy.TSourceType_PARQUET_FILE.String()
	}
	builder := strings.Builder{}
	builder.WriteString(
		fmt.Sprintf(
			`CREATE SERVER %s FOREIGN DATA WRAPPER %s`,
			serverName,
			dataWrapperType,
		),
	)

	builder.WriteString(buildS3CreateServerWithClause(
		bucket,
		region,
	))

	createServerQuery := builder.String()

	if config.Enable.Verbose == true {
		config.Log.Info("CREATE SERVER query: ", createServerQuery)
	}
	_, err = services.SQLExecute(sessionID, createServerQuery)
	return
}
