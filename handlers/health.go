// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
)

type serverStatus struct {
	Commit              string `json:"Commit"`
	AllowAnyOrigin      string `json:"AllowAnyOrigin"`
	Compress            string `json:"Compression"`
	LocalDataCatalog    string `json:"LocalDataCatalog"`
	Jupyter             string `json:"Jupyter"`
	HTTPS               string `json:"HTTPS"`
	HTTPSAuth           string `json:"HTTPSAuth"`
	HTTPSRedirect       string `json:"HTTPSRedirect"`
	LegacyAuth          string `json:"LegacyAuth"`
	ReverseProxies      string `json:"ReverseProxies"`
	ReadOnly            string `json:"ReadOnly"`
	SAML                string `json:"SAML"`
	Verbose             string `json:"Verbose"`
	IdleSessionDuration string `json:"IdleSessionDuration"`
	MaxSessionDuration  string `json:"MaxSessionDuration"`
}

// Health - Health Check Handler
func Health(ctx echo.Context) error {
	if !config.Enable.Development {
		return ctx.NoContent(http.StatusOK)
	}

	s := &serverStatus{
		Commit:              config.Other.Commit,
		AllowAnyOrigin:      strconv.FormatBool(config.Enable.AllowAnyOrigin),
		Compress:            strconv.FormatBool(config.Enable.Compress),
		LocalDataCatalog:    strconv.FormatBool(config.Enable.LocalDataCatalog),
		Jupyter:             strconv.FormatBool(config.Enable.Jupyter),
		HTTPS:               strconv.FormatBool(config.Enable.HTTPS),
		HTTPSAuth:           strconv.FormatBool(config.Enable.HTTPSAuth),
		HTTPSRedirect:       strconv.FormatBool(config.Enable.HTTPSRedirect),
		LegacyAuth:          strconv.FormatBool(config.Enable.LegacyAuth),
		ReverseProxies:      strconv.FormatBool(config.Enable.ReverseProxies),
		ReadOnly:            strconv.FormatBool(config.Enable.ReadOnly),
		SAML:                strconv.FormatBool(config.Enable.SAML),
		Verbose:             strconv.FormatBool(config.Enable.Verbose),
		IdleSessionDuration: config.Session.IdleTimeout.String(),
		MaxSessionDuration:  config.Session.MaxTimeout.String(),
	}

	return ctx.JSONPretty(http.StatusOK, s, "  ")
}
