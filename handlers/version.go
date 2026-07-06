// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
)

// Version - Version Handler
// Generates a version.txt file
func Version(ctx echo.Context) error {
	dbVersion, err := util.GetHeavyDBVersion(ctx)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Could not get HeavyDB version: "+err.Error(),
		)
	}
	outVers := "HeavyDB:\n" + dbVersion

	versTxt := config.Paths.Frontend + "/version.txt"
	feVers, err := os.ReadFile(versTxt)
	if err == nil {
		outVers += "\n\n"
		outVers += "Immerse:\n"
		outVers += string(feVers)
	}

	outVers += "\n\n"
	outVers += "Web Server:\n"
	outVers += string(config.Other.Commit)

	if config.Enable.IQ {
		iqVers, err := getIQVersion()
		if err != nil {
			config.Log.Error("Could not get HeavyIQ version")
		} else {
			outVers += ("\n\n" + "HeavyIQ:\n" + iqVers)
		}
	}

	return ctx.String(http.StatusOK, outVers)
}

func getIQVersion() (version string, err error) {
	resp, err := http.Get(config.Paths.IQServiceURL.String() + "/version.txt")
	if err != nil {
		config.Log.Error("Could not GET HeavyIQ version")
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		config.Log.Error("Could not access HeavyIQ version")
		return "", err
	}

	return string(body), nil
}
