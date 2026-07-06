// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/heavyai/webserver/models"
	"github.com/heavyai/webserver/services"
)

// GetDBRoles - GetDBRoles util method
func GetDBRoles(sessionInfo models.DBSessionInfo, optionalRole string) (roles []string, err error) {
	thriftClient, err := services.CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()
	userOrRole := sessionInfo.Username
	if len(optionalRole) > 0 {
		userOrRole = optionalRole
	}
	return thriftClient.Client.GetAllEffectiveRolesForUser(thriftClient.Ctx, sessionInfo.SessionID, userOrRole)
}

// HasRole - HasRole util method
func HasRole(sessionInfo models.DBSessionInfo, role string) (hasRole bool, err error) {
	roles, err := GetDBRoles(sessionInfo, "")

	if err != nil {
		return false, err
	}

	for _, v := range roles {
		if v == role {
			return true, nil
		}
	}

	return false, nil
}
