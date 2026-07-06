// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

// AuthCredentials - AuthCredentials model
type AuthCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
}
