// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/services"

	"github.com/heavyai/webserver/internal/util"
)

// SupportedDataSourcesRow struct
type SupportedDataSourcesRow struct {
	SourceCategory string `json:"sourceCategory,omitempty"`
	SourceType     string `json:"sourceType,omitempty"`
	SourceSubType  string `json:"sourceSubType,omitempty"`
}

// SupportedDataSourcesResponseBody type
type SupportedDataSourcesResponseBody []SupportedDataSourcesRow

// ConfigurationSupportedDataSourcesHandler - GET
func ConfigurationSupportedDataSourcesHandler(ctx echo.Context) (err error) {
	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	supportedDataSources, err := services.SQLExecute(
		sessionID,
		"SHOW SUPPORTED DATA SOURCES",
	)
	if err != nil {
		config.Log.Error("failed to retrieve supported data sources", err)
		return ctx.NoContent(http.StatusBadGateway)
	}

	var sourceRows SupportedDataSourcesResponseBody
	var sourceCategories []string
	var sourceTypes []string
	var sourceSubTypes []string

	for i, column := range supportedDataSources.RowSet.Columns {
		if i == 0 {
			sourceCategories = column.Data.StrCol
		} else if i == 1 {
			sourceTypes = column.Data.StrCol
		} else if i == 2 {
			sourceSubTypes = column.Data.StrCol
		}
	}

	for n, categroy := range sourceCategories {
		row := SupportedDataSourcesRow{
			SourceCategory: categroy,
			SourceType:     sourceTypes[n],
			SourceSubType:  sourceSubTypes[n],
		}
		sourceRows = append(sourceRows, row)
	}

	return ctx.JSON(http.StatusOK, sourceRows)
}
