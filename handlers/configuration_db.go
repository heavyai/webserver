// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// ConfigurationDBCreate - POST
func ConfigurationDBCreate(ctx echo.Context) (err error) {
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		config.Log.Error(err)
		ctx.NoContent(http.StatusInternalServerError)
		return
	}

	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := util.GetDBSessionInfo(immerseCtx.Claims, immerseCtx.DBName)
	if err != nil {
		config.Log.Error(err)
		ctx.NoContent(http.StatusInternalServerError)
		return
	}

	config.Other.DBConfig[sessionInfo.Database] = body
	err = util.UpdateDBConfigJSON()
	if err != nil {
		config.Log.Error("Unable to persist DB Configuration update", err)
		ctx.NoContent(http.StatusUnprocessableEntity)
		return
	}
	ctx.NoContent(http.StatusCreated)
	return
}

// ConfigurationDBRead - GET
func ConfigurationDBRead(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := util.GetDBSessionInfo(immerseCtx.Claims, immerseCtx.DBName)
	if err != nil {
		config.Log.Error(err)
		ctx.JSONBlob(http.StatusInternalServerError, []byte("{}"))
		return
	}

	if val, ok := config.Other.DBConfig[sessionInfo.Database]; ok {
		return ctx.JSONBlob(http.StatusOK, val)
	}

	return ctx.JSONBlob(http.StatusBadRequest, []byte("{}"))
}
