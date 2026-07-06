// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// FetchRoles - FetchRoles handler
func FetchRoles(ctx echo.Context) (err error) {
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo := *immerseCtx.Claims.SessionsInfoMap.GetSingleSessionInfo()

	encodedRole := ctx.Param(constants.RoleRouteParamName)
	decodedRole, err := url.QueryUnescape(encodedRole)
	if err != nil {
		config.Log.Error("Could not evaluate role '"+encodedRole+"'", err)
		ctx.String(http.StatusInternalServerError, "Could not evaluate role '"+encodedRole+"'")
		return err
	}

	thriftRoles, err := util.GetDBRoles(sessionInfo, decodedRole)

	if err != nil {
		userOrRole := sessionInfo.Username
		if len(decodedRole) > 0 {
			userOrRole = decodedRole
		}
		config.Log.Error("Could not fetch roles for user/role '"+userOrRole+"'", err)
		ctx.String(http.StatusNotFound, "Could not fetch roles for user/role '"+userOrRole+"'")
		return err
	}
	return ctx.JSON(http.StatusOK, thriftRoles)
}
