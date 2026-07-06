// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/json"
)

// QueryHistory - AddQueryHistory payload
type QueryHistory struct {
	IsVega    bool            `json:"isVega"`
	Query     string          `json:"query"`
	Error     *string         `json:"error"`
	Results   json.RawMessage `json:"results"`
	Timestamp int64           `json:"timestamp"`
}
