// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// ServersJSON - ServersJSON Handler
func ServersJSON(ctx echo.Context) error {
	jj, err := GetServersJSON(*ctx.Request())
	if err != nil {
		config.Log.Error("Error processing servers.json: ", err)

		return err
	}

	ctx.Response().Header().Del("Cache-Control")
	ctx.Response().Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")

	return ctx.JSONBlob(http.StatusOK, jj)
}

// GetServersJSON - GetServersJSON Helper
func GetServersJSON(req http.Request) ([]byte, error) {
	cache := getBaseServersJSON()
	if config.Enable.CustomIntegrationAuth {
		session, _ := config.Session.SessionStore.Get(&req, constants.CookieModifiedServersJSON)
		if session != nil || cache.Refreshed {
			cache.Refreshed = false
			return modifyServersJSON(&req, cache.Bytes)
		}
	}
	return cache.Bytes, nil
}

func modifyServersJSON(req *http.Request, orig []byte) ([]byte, error) {
	session, _ := config.Session.SessionStore.Get(req, constants.CookieModifiedServersJSON)
	j, err := gabs.ParseJSON(orig)
	if err != nil {
		return nil, err
	}

	jj := j.Children()
	if jj == nil {
		return nil, errors.New("Could not parse JSON children")
	}

	for _, key := range config.Other.ServersJSONParams {
		if session.Values[key] != nil {
			_, err = jj[0].Set(session.Values[key].(string), key)
			if err != nil {
				return nil, err
			}
		}
	}

	return j.BytesIndent("", "  "), nil
}

func getBaseServersJSON() (cache *models.FileCache) {
	_, validateJSONErr := gabs.ParseJSONFile(config.Paths.ServersJSON)
	if validateJSONErr == nil {
		cache, err := util.GetServersJSONFileCache()
		if err == nil && cache.Bytes != nil {
			return cache
		}
	}

	models.ServersJSONCache = models.FileCache{
		Bytes:   generateServersJSONBytes(),
		ModTime: time.Now(),
	}
	return &models.ServersJSONCache
}

func generateServersJSONBytes() (jsonBytes []byte) {
	s := models.ServerConfig{}
	s.Master = true
	s.Username = "admin"
	s.Password = "HyperInteractive"
	s.Database = "heavyai"

	serverArray := []models.ServerConfig{s}
	jsonBytes, _ = json.MarshalIndent(serverArray, "", "  ")
	return jsonBytes
}

// SetServersJSON - SetServersJSON Handler used only by Tutela
func SetServersJSON(ctx echo.Context) error {
	session, err := config.Session.SessionStore.Get(ctx.Request(), constants.CookieModifiedServersJSON)
	if err != nil {
		util.DeleteHTTPOnlyCookie(ctx, constants.CookieModifiedServersJSON)
		session, _ = config.Session.SessionStore.Get(ctx.Request(), constants.CookieModifiedServersJSON)
	}

	for _, key := range config.Other.ServersJSONParams {
		if len(ctx.Request().FormValue(key)) > 0 {
			session.Values[key] = ctx.Request().FormValue(key)
		}
	}

	session.Save(ctx.Request(), ctx.Response())

	return ctx.NoContent(http.StatusOK)
}

// ClearServersJSON - ClearServersJSON Handler used only by Tutela
func ClearServersJSON(ctx echo.Context) error {
	util.DeleteHTTPOnlyCookie(ctx, constants.CookieModifiedServersJSON)

	return ctx.NoContent(http.StatusOK)
}
