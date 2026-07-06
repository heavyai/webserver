// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

// DBAccessList - DBAccessList model
type DBAccessList []*DBInfo

// CreateDBAccessList - CreateDBAccessList factory
func CreateDBAccessList(thriftDBAccessList []*heavy.TDBInfo) (list DBAccessList) {
	for i := range thriftDBAccessList {
		list = append(list, &DBInfo{DBName: thriftDBAccessList[i].DbName})
	}
	return
}
