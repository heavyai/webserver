// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"errors"
)

// SessionsInfoMap - SessionsInfoMap type
type SessionsInfoMap map[string]*DBSessionInfo

// GetSingleSessionInfo - Just "do our best to handle a legacy urls w/o DB contexts 🤞
func (sessionsInfoMap SessionsInfoMap) GetSingleSessionInfo() (sessionInfoRef *DBSessionInfo) {
	if len(sessionsInfoMap) == 0 {
		return
	}
	// Only way I could find to get an arbitrary key from a map in golang
	for dbName := range sessionsInfoMap {
		sessionInfoRef = sessionsInfoMap[dbName]
		// return now since we just want the first one, whatever that ends up being
		return
	}
	return
}

// GetDBSessionInfo - GetDBSessionInfo method
func (sessionsInfoMap SessionsInfoMap) GetDBSessionInfo(DBName string) (sessionInfoRef *DBSessionInfo, err error) {
	if len(DBName) > 0 {
		if sessionsInfoMap[DBName] == nil {
			return nil, errors.New("Could not find SessionInfo: No session for given DB context '" + DBName + "'")
		}
		return sessionsInfoMap[DBName], nil
	}
	sessionInfoRef = sessionsInfoMap.GetSingleSessionInfo()
	if sessionInfoRef == nil {
		err = errors.New("Could not retrieve implicit SessionInfo")
	}
	return
}

// DeleteSessionInfo - DeleteSessionInfo method
func (sessionsInfoMap SessionsInfoMap) DeleteSessionInfo(DBName string) {
	delete(sessionsInfoMap, DBName)
}

// MergeMap - MergeMap method
func (sessionsInfoMap SessionsInfoMap) MergeMap(mergeMap SessionsInfoMap) {
	for k := range mergeMap {
		sessionsInfoMap[k] = mergeMap[k]
	}
}
