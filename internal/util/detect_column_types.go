// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/heavyai/webserver/internal/db_thrift_client/common"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/models"
	"github.com/heavyai/webserver/services"
)

// ModifiedTCopyParams - type
type ModifiedTCopyParams struct {
	Delimiter                string `thrift:"delimiter,1" db:"delimiter" json:"delimiter"`
	NullStr                  string `thrift:"null_str,2" db:"null_str" json:"null_str"`
	HasHeader                int64  `thrift:"has_header,3" db:"has_header" json:"has_header"`
	Quoted                   bool   `thrift:"quoted,4" db:"quoted" json:"quoted"`
	Quote                    string `thrift:"quote,5" db:"quote" json:"quote"`
	Escape                   string `thrift:"escape,6" db:"escape" json:"escape"`
	LineDelim                string `thrift:"line_delim,7" db:"line_delim" json:"line_delim"`
	ArrayDelim               string `thrift:"array_delim,8" db:"array_delim" json:"array_delim"`
	ArrayBegin               string `thrift:"array_begin,9" db:"array_begin" json:"array_begin"`
	ArrayEnd                 string `thrift:"array_end,10" db:"array_end" json:"array_end"`
	Threads                  int32  `thrift:"threads,11" db:"threads" json:"threads"`
	SourceType               int64  `thrift:"source_type,12" db:"source_type" json:"source_type"`
	S3AccessKey              string `thrift:"s3_access_key,13" db:"s3_access_key" json:"s3_access_key"`
	S3SecretKey              string `thrift:"s3_secret_key,14" db:"s3_secret_key" json:"s3_secret_key"`
	S3Region                 string `thrift:"s3_region,15" db:"s3_region" json:"s3_region"`
	GeoCoordsEncoding        int64  `thrift:"geo_coords_encoding,16" db:"geo_coords_encoding" json:"geo_coords_encoding"`
	GeoCoordsCompParam       int32  `thrift:"geo_coords_comp_param,17" db:"geo_coords_comp_param" json:"geo_coords_comp_param"`
	GeoCoordsType            int64  `thrift:"geo_coords_type,18" db:"geo_coords_type" json:"geo_coords_type"`
	GeoCoordsSrid            int32  `thrift:"geo_coords_srid,19" db:"geo_coords_srid" json:"geo_coords_srid"`
	SanitizeColumnNames      bool   `thrift:"sanitize_column_names,20" db:"sanitize_column_names" json:"sanitize_column_names"`
	GeoLayerName             string `thrift:"geo_layer_name,21" db:"geo_layer_name" json:"geo_layer_name"`
	S3Endpoint               string `thrift:"s3_endpoint,22" db:"s3_endpoint" json:"s3_endpoint"`
	GeoAssignRenderGroups    bool   `thrift:"geo_assign_render_groups,23" db:"geo_assign_render_groups" json:"geo_assign_render_groups"`
	GeoExplodeCollections    bool   `thrift:"geo_explode_collections,24" db:"geo_explode_collections" json:"geo_explode_collections"`
	SourceSrid               int32  `thrift:"source_srid,25" db:"source_srid" json:"source_srid"`
	S3SessionToken           string `thrift:"s3_session_token,26" db:"s3_session_token" json:"s3_session_token"`
	RasterPointType          int64  `thrift:"raster_point_type,27" db:"raster_point_type" json:"raster_point_type"`
	RasterImportBands        string `thrift:"raster_import_bands,28" db:"raster_import_bands" json:"raster_import_bands"`
	RasterScanlinesPerThread int32  `thrift:"raster_scanlines_per_thread,29" db:"raster_scanlines_per_thread" json:"raster_scanlines_per_thread"`
	RasterPointTransform     int64  `thrift:"raster_point_transform,30" db:"raster_point_transform" json:"raster_point_transform"`
	RasterPointComputeAngle  bool   `thrift:"raster_point_compute_angle,31" db:"raster_point_compute_angle" json:"raster_point_compute_angle"`
	RasterImportDimensions   string `thrift:"raster_import_dimensions,32" db:"raster_import_dimensions" json:"raster_import_dimensions"`
	OdbcDsn                  string `thrift:"odbc_dsn,33" db:"odbc_dsn" json:"odbc_dsn"`
	OdbcConnectionString     string `thrift:"odbc_connection_string,34" db:"odbc_connection_string" json:"odbc_connection_string"`
	OdbcSQLSelect            string `thrift:"odbc_sql_select,35" db:"odbc_sql_select" json:"odbc_sql_select"`
	OdbcSQLOrderBy           string `thrift:"odbc_sql_order_by,36" db:"odbc_sql_order_by" json:"odbc_sql_order_by"`
	OdbcUsername             string `thrift:"odbc_username,37" db:"odbc_username" json:"odbc_username"`
	OdbcPassword             string `thrift:"odbc_password,38" db:"odbc_password" json:"odbc_password"`
	OdbcCredentialString     string `thrift:"odbc_credential_string,39" db:"odbc_credential_string" json:"odbc_credential_string"`
	AddMetadataColumns       string `thrift:"add_metadata_columns,40" db:"add_metadata_columns" json:"add_metadata_columns"`
}

// ModifiedTTypeInfo - type
type ModifiedTTypeInfo struct {
	Type      int64 `thrift:"type,1" db:"type" json:"type"`
	Nullable  bool  `thrift:"nullable,2" db:"nullable" json:"nullable"`
	IsArray   bool  `thrift:"is_array,3" db:"is_array" json:"is_array"`
	Encoding  int64 `thrift:"encoding,4" db:"encoding" json:"encoding"`
	Precision int32 `thrift:"precision,5" db:"precision" json:"precision"`
	Scale     int32 `thrift:"scale,6" db:"scale" json:"scale"`
	CompParam int32 `thrift:"comp_param,7" db:"comp_param" json:"comp_param"`
	Size      int32 `thrift:"size,8" db:"size" json:"size"`
}

// ModifiedTColumnType - type
type ModifiedTColumnType struct {
	ColName           string             `thrift:"col_name,1" db:"col_name" json:"col_name"`
	ColType           *ModifiedTTypeInfo `thrift:"col_type,2" db:"col_type" json:"col_type"`
	IsReservedKeyword bool               `thrift:"is_reserved_keyword,3" db:"is_reserved_keyword" json:"is_reserved_keyword"`
	SrcName           string             `thrift:"src_name,4" db:"src_name" json:"src_name"`
	IsSystem          bool               `thrift:"is_system,5" db:"is_system" json:"is_system"`
	IsPhysical        bool               `thrift:"is_physical,6" db:"is_physical" json:"is_physical"`
	ColID             int64              `thrift:"col_id,7" db:"col_id" json:"col_id"`
	DefaultValue      *string            `thrift:"default_value,8" db:"default_value" json:"default_value,omitempty"`
}

// ModifiedTRowDescriptor - type
type ModifiedTRowDescriptor []*ModifiedTColumnType

// ModifiedTRowSet - type
type ModifiedTRowSet struct {
	RowDesc    ModifiedTRowDescriptor `thrift:"row_desc,1" db:"row_desc" json:"row_desc"`
	Rows       []*heavy.TRow          `thrift:"rows,2" db:"rows" json:"rows"`
	Columns    []*heavy.TColumn       `thrift:"columns,3" db:"columns" json:"columns"`
	IsColumnar bool                   `thrift:"is_columnar,4" db:"is_columnar" json:"is_columnar"`
}

// ModifiedTDetectResult - type
type ModifiedTDetectResult struct {
	RowSet     *ModifiedTRowSet     `thrift:"row_set,1" db:"row_set" json:"row_set"`
	CopyParams *ModifiedTCopyParams `thrift:"copy_params,2" db:"copy_params" json:"copy_params"`
}

// DetectColumnTypes - util
func DetectColumnTypes(
	sessionID heavy.TSessionId,
	dctConfig models.DetectColumnTypesConfig,
) (result *ModifiedTDetectResult, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	var copyParams heavy.TCopyParams
	if dctConfig.SourceType == heavy.TSourceType_ODBC.String() {
		sourceType, err := heavy.TSourceTypeFromString(dctConfig.SourceType)
		if err != nil {
			return result, err
		}
		copyParams.SourceType = sourceType
		copyParams.OdbcCredentialString = models.CreateOdbcCredentialString(
			dctConfig.DBUserName,
			dctConfig.DBPassword,
		)
		copyParams.OdbcSqlSelect = dctConfig.SQLSelect
		copyParams.OdbcSqlOrderBy = dctConfig.SQLOrderBy
		copyParams.OdbcConnectionString = models.GetOdbcConnectionString(
			dctConfig.Driver,
			dctConfig.DBName,
			dctConfig.DBHost,
			dctConfig.DBPort,
			dctConfig.DataWarehouse,
			dctConfig.Role,
		)
	} else {
		// TODO: implement non-ODBC DCT requests
	}

	if !(copyParams.GeoCoordsType == common.TDatumType_GEOMETRY ||
		copyParams.GeoCoordsType == common.TDatumType_GEOGRAPHY) {
		copyParams.GeoCoordsType = common.TDatumType_GEOMETRY
	}

	// Based on WITH options given here: https://docs.heavy.ai/loading-and-exporting-data/command-line/load-data#geo-import
	if copyParams.GeoCoordsSrid != 4326 {
		copyParams.GeoCoordsSrid = 4326
	}

	brokenThriftResult, err := thriftClient.Client.DetectColumnTypes(
		thriftClient.Ctx,
		sessionID,
		dctConfig.FileName,
		&copyParams,
	)
	if err != nil {
		return
	}

	modifiedRowDescriptor := ModifiedTRowDescriptor{}
	for _, brokenRow := range brokenThriftResult.RowSet.RowDesc {
		modifiedRowDescriptor = append(
			modifiedRowDescriptor,
			&ModifiedTColumnType{
				ColName: brokenRow.ColName,
				ColType: &ModifiedTTypeInfo{
					Type:      int64(brokenRow.ColType.Type),
					Nullable:  brokenRow.ColType.Nullable,
					IsArray:   brokenRow.ColType.IsArray,
					Encoding:  int64(brokenRow.ColType.Encoding),
					Precision: brokenRow.ColType.Precision,
					Scale:     brokenRow.ColType.Scale,
					CompParam: brokenRow.ColType.CompParam,
					Size:      brokenRow.ColType.Size,
				},
				IsReservedKeyword: brokenRow.IsReservedKeyword,
				SrcName:           brokenRow.SrcName,
				IsSystem:          brokenRow.IsSystem,
				IsPhysical:        brokenRow.IsPhysical,
				ColID:             brokenRow.ColID,
				DefaultValue:      brokenRow.DefaultValue,
			},
		)
	}

	modifiedThriftResult := ModifiedTDetectResult{
		RowSet: &ModifiedTRowSet{
			RowDesc:    modifiedRowDescriptor,
			Rows:       brokenThriftResult.RowSet.Rows,
			Columns:    brokenThriftResult.RowSet.Columns,
			IsColumnar: brokenThriftResult.RowSet.IsColumnar,
		},
		CopyParams: &ModifiedTCopyParams{
			Delimiter:                brokenThriftResult.CopyParams.Delimiter,
			NullStr:                  brokenThriftResult.CopyParams.NullStr,
			HasHeader:                int64(brokenThriftResult.CopyParams.HasHeader),
			Quoted:                   brokenThriftResult.CopyParams.Quoted,
			Quote:                    brokenThriftResult.CopyParams.Quote,
			Escape:                   brokenThriftResult.CopyParams.Escape,
			LineDelim:                brokenThriftResult.CopyParams.LineDelim,
			ArrayDelim:               brokenThriftResult.CopyParams.ArrayDelim,
			ArrayBegin:               brokenThriftResult.CopyParams.ArrayBegin,
			ArrayEnd:                 brokenThriftResult.CopyParams.ArrayEnd,
			Threads:                  brokenThriftResult.CopyParams.Threads,
			SourceType:               int64(brokenThriftResult.CopyParams.SourceType),
			S3AccessKey:              brokenThriftResult.CopyParams.S3AccessKey,
			S3SecretKey:              brokenThriftResult.CopyParams.S3SecretKey,
			S3Region:                 brokenThriftResult.CopyParams.S3Region,
			GeoCoordsEncoding:        int64(brokenThriftResult.CopyParams.GeoCoordsEncoding),
			GeoCoordsCompParam:       brokenThriftResult.CopyParams.GeoCoordsCompParam,
			GeoCoordsType:            int64(brokenThriftResult.CopyParams.GeoCoordsType),
			GeoCoordsSrid:            brokenThriftResult.CopyParams.GeoCoordsSrid,
			SanitizeColumnNames:      brokenThriftResult.CopyParams.SanitizeColumnNames,
			GeoLayerName:             brokenThriftResult.CopyParams.GeoLayerName,
			S3Endpoint:               brokenThriftResult.CopyParams.S3Endpoint,
			GeoAssignRenderGroups:    brokenThriftResult.CopyParams.GeoAssignRenderGroups,
			GeoExplodeCollections:    brokenThriftResult.CopyParams.GeoExplodeCollections,
			SourceSrid:               brokenThriftResult.CopyParams.SourceSrid,
			S3SessionToken:           brokenThriftResult.CopyParams.S3SessionToken,
			RasterPointType:          int64(brokenThriftResult.CopyParams.RasterPointType),
			RasterImportBands:        brokenThriftResult.CopyParams.RasterImportBands,
			RasterScanlinesPerThread: brokenThriftResult.CopyParams.RasterScanlinesPerThread,
			RasterPointTransform:     int64(brokenThriftResult.CopyParams.RasterPointTransform),
			RasterPointComputeAngle:  brokenThriftResult.CopyParams.RasterPointComputeAngle,
			RasterImportDimensions:   brokenThriftResult.CopyParams.RasterImportDimensions,
			OdbcDsn:                  brokenThriftResult.CopyParams.OdbcDsn,
			OdbcConnectionString:     brokenThriftResult.CopyParams.OdbcConnectionString,
			OdbcSQLSelect:            brokenThriftResult.CopyParams.OdbcSqlSelect,
			OdbcSQLOrderBy:           brokenThriftResult.CopyParams.OdbcSqlOrderBy,
			OdbcUsername:             brokenThriftResult.CopyParams.OdbcUsername,
			OdbcPassword:             brokenThriftResult.CopyParams.OdbcPassword,
			OdbcCredentialString:     brokenThriftResult.CopyParams.OdbcCredentialString,
			AddMetadataColumns:       brokenThriftResult.CopyParams.AddMetadataColumns,
		},
	}

	return &modifiedThriftResult, err
}
