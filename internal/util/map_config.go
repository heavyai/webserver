// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"

	"github.com/heavyai/webserver/internal/config"
)

type mapConfigVars struct {
	MapboxToken  string `json:"mapboxToken"`
	GoogleAPIKey string `json:"googleApiKey"`
}

// GetMapConfigJSON - GetMapConfigJSON util
func GetMapConfigJSON() (mcvJSON []byte, err error) {
	mcv := &mapConfigVars{
		MapboxToken:  config.Other.MapboxToken,
		GoogleAPIKey: config.Other.GoogleAPIKey,
	}

	return json.MarshalIndent(mcv, "", "  ")
}
