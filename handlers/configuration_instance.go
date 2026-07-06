// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"io"
	"net/http"
	"path/filepath"

	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// ConfigurationInstanceCreate - POST
// Sets user-configurable settings that are persisted at instance level
func ConfigurationInstanceCreate(ctx echo.Context) (err error) {
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		config.Log.Error(err)
		ctx.NoContent(http.StatusInternalServerError)
		return
	}

	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := util.GetDBSessionInfo(immerseCtx.Claims, immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return
	}

	if !sessionInfo.IsSuper {
		singleSessionInfo := *immerseCtx.Claims.SessionsInfoMap.GetSingleSessionInfo()

		hasControlPanelRole, err := util.HasRole(singleSessionInfo, constants.ControlPanelAdminRole)
		if err != nil {
			config.Log.Error(err)
			return ctx.NoContent(http.StatusInternalServerError)
		}

		if !hasControlPanelRole {
			return ctx.NoContent(http.StatusUnauthorized)
		}
	}

	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		return err
	}

	err = util.WriteJSONToFile(jsonParsed, filepath.Join(config.Paths.DataDir, constants.InstanceConfigJSONFileName))

	if err != nil {
		config.Log.Error("Unable to persist instance configuration update", err)
		ctx.NoContent(http.StatusUnprocessableEntity)
		return
	}
	config.Other.InstanceConfig = jsonParsed

	ctx.NoContent(http.StatusCreated)
	return
}

// ConfigurationInstanceRead - GET
// Retrieves all user-configurable settings that are persisted at instance level
func ConfigurationInstanceRead(ctx echo.Context) (err error) {
	jsonBytes := config.Other.InstanceConfig.S().Bytes()

	if config.Other.InstanceConfig.ExistsP(constants.InstanceConfigLoadError) && len(config.Other.InstanceConfig.ChildrenMap()) == 1 {
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSONBlob(http.StatusOK, jsonBytes)
}
