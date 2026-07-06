// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
	"github.com/heavyai/webserver/services"
)

// SetTableCommentHandler - SetTableComment Handler
func SetTableCommentHandler(ctx echo.Context) (err error) {
	var tableComment models.TableComment
	err = ctx.Bind(&tableComment)

	if err != nil {
		ctx.String(http.StatusBadRequest, "Bad request")
		return err
	}

	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	tableName := ctx.Param(constants.TableNameRouteParamName)

	query := models.BuildSetTableCommentQuery(tableName, tableComment.Comment)
	config.Log.Info("Table comment query: ", query)

	_, err = services.SQLExecute(sessionID, query)

	if err != nil {
		config.Log.Error("Failed to set table comment", err)
		ctx.String(http.StatusBadGateway, util.ErrorMsg(err))
	}

	return ctx.JSON(http.StatusOK, tableComment)
}

// SetColumnCommentHandler - SetColumnComment Handler
func SetColumnCommentHandler(ctx echo.Context) (err error) {
	var columnComment models.ColumnComment

	err = ctx.Bind(&columnComment)

	if err != nil {
		ctx.String(http.StatusBadRequest, "Bad request")
		return err
	}

	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	tableName := ctx.Param(constants.TableNameRouteParamName)

	query := models.BuildSetColumnCommentQuery(tableName, columnComment.ColumnName, columnComment.Comment)
	config.Log.Info("Column comment query: ", query)

	_, err = services.SQLExecute(sessionID, query)

	if err != nil {
		config.Log.Error("Failed to set column comment", err)
		ctx.String(http.StatusBadGateway, util.ErrorMsg(err))

	}

	return ctx.JSON(http.StatusOK, columnComment)
}

// DeleteTableCommentHandler - DeleteTableComment Handler
func DeleteTableCommentHandler(ctx echo.Context) (err error) {
	tableName := ctx.Param(constants.TableNameRouteParamName)
	query := models.BuildDeleteTableCommentQuery(tableName)
	config.Log.Info("Delete table comment query: ", query)

	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	_, err = services.SQLExecute(sessionID, query)

	if err != nil {
		config.Log.Error("Failed to delete table comment", err)
		ctx.String(http.StatusBadGateway, util.ErrorMsg(err))
	}

	return ctx.JSON(http.StatusOK, tableName)
}

// DeleteColumnCommentHandler - DeleteColumnComment Handler
func DeleteColumnCommentHandler(ctx echo.Context) (err error) {
	var column models.Column
	err = ctx.Bind(&column)

	if err != nil {
		ctx.String(http.StatusBadRequest, "Bad request")
		return err
	}

	tableName := ctx.Param(constants.TableNameRouteParamName)

	query := models.BuildDeleteColumnCommentQuery(tableName, column.ColumnName)
	config.Log.Info("Delete column comment query: ", query)

	sessionID, err := util.GetSessionIDFromCtx(ctx)
	if err != nil {
		return
	}

	_, err = services.SQLExecute(sessionID, query)

	if err != nil {
		config.Log.Error("Failed to delete column comment", err)
		ctx.String(http.StatusBadGateway, util.ErrorMsg(err))
	}

	return ctx.JSON(http.StatusOK, column)
}
