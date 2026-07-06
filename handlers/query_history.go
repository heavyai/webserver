// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// GetQueryHistory - Gets history from Immerse user metadata for the current database
func GetQueryHistory(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)

	if err != nil {
		config.Log.Error("Session does not exist for given DB context: '"+immerseCtx.DBName+"'", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	userInfo, err := util.GetUserInfo(sessionInfo.SessionID)

	if err != nil {
		config.Log.Error("Could not get user info: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	metadataJSON, err := models.GetMetadataForUser(ctx, userInfo)

	if err != nil {
		return ctx.NoContent(http.StatusInternalServerError)
	}

	parsedMetadata, err := gabs.ParseJSON([]byte(metadataJSON))

	if err != nil {
		config.Log.Error("Could not parse existing user metadata; invalid JSON: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	history := make([]interface{}, 0)

	if historyForDb := parsedMetadata.Path("queryHistory." + sessionInfo.DBName).Data(); historyForDb != nil {
		if arr, ok := historyForDb.([]interface{}); ok {
			history = arr
		} else {
			config.Log.Error("Could not fetch query history; query history is not an array")
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	return ctx.JSON(http.StatusOK, history)
}

// AddQueryHistory - Appends history to Immerse user metadata for the current database
func AddQueryHistory(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)

	if err != nil {
		config.Log.Error("Session does not exist for given DB context: '"+immerseCtx.DBName+"'", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	var queryHistory models.QueryHistory
	err = ctx.Bind(&queryHistory)

	if err != nil {
		return ctx.String(http.StatusBadRequest, "Bad request")
	}

	userInfo, err := util.GetUserInfo(sessionInfo.SessionID)

	if err != nil {
		config.Log.Error("Could not get user info: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	metadataJSON, err := models.GetMetadataForUser(ctx, userInfo)

	if err != nil {
		return ctx.NoContent(http.StatusInternalServerError)
	}

	parsedMetadata, err := gabs.ParseJSON([]byte(metadataJSON))

	if err != nil {
		config.Log.Error("Could not parse existing user metadata; invalid JSON: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	queryHistoryDbPath := "queryHistory." + sessionInfo.DBName
	existingQueryHistoryContainer := parsedMetadata.Path(queryHistoryDbPath)

	// Init query history if it does not exist
	if existingQueryHistoryContainer == nil {
		parsedMetadata.ArrayP(queryHistoryDbPath)
	}

	queryHistoryLength, err := parsedMetadata.ArrayCountP(queryHistoryDbPath)

	if err != nil {
		config.Log.Error("Could not get query history length: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// Remove oldest query if max query history is stored
	if queryHistoryLength >= 10 {
		parsedMetadata.ArrayRemoveP(0, queryHistoryDbPath)
	}

	parsedMetadata.ArrayAppendP(queryHistory, queryHistoryDbPath)

	// Persist updated metadata
	userMetadata := []*heavy.TImmerseUserMetadata{
		{Username: immerseCtx.Claims.Username, ImmerseMetadataJSON: parsedMetadata.String()},
	}

	err = util.PutImmerseUsersMetadata(sessionInfo.SessionID, userMetadata)

	if err != nil {
		config.Log.Error("Could not set user info: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
