// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/models"
)

// GetServersJSONFileCache - GetServersJSONFileCache util
func GetServersJSONFileCache() (cache *models.FileCache, err error) {
	fileHeaders, err := os.Stat(config.Paths.ServersJSON)
	if err != nil {
		config.Log.Error("Servers JSON Path Error: ", err)
		return
	}

	cache = &models.ServersJSONCache
	modTime := fileHeaders.ModTime()

	// If the memory cache hasn't been set yet, or the file's ModTime has changed, then go ahead and read the file,
	// set the ModTime and set them on ServersJSONCache
	if len(models.ServersJSONCache.Bytes) == 0 || (models.ServersJSONCache.ModTime.Before(modTime)) {
		models.ServersJSONCache.Refreshed = true
		models.ServersJSONCache.ModTime = modTime
		bytes, err := os.ReadFile(config.Paths.ServersJSON)
		if err != nil {
			cache = &models.FileCache{}
			return cache, err
		}
		models.ServersJSONCache.Bytes = bytes
	}
	return cache, nil
}

// SaveFileToLocalDir - SaveFileToLocalDir util
func SaveFileToLocalDir(fileHeader *multipart.FileHeader, dir string) (err error) {
	infile, err := fileHeader.Open()
	if err != nil {
		return err
	}
	// 0750 file permissions applied here based on InfoSec requiremnt here: https://github.com/heavyai/webserver/pull/28#discussion_r321890368
	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}
	outFile, err := os.Create(filepath.Join(dir, fileHeader.Filename))
	if err != nil {
		return err
	}
	_, err = io.Copy(outFile, infile)
	if err != nil {
		return err
	}
	return
}

// GetDBConfigJSON - GetDBConfigJSON util
func GetDBConfigJSON() (dbcJSON []byte, err error) {
	parseData := make([]map[string]interface{}, 0, 0)
	for k, v := range config.Other.DBConfig {
		var c = make(map[string]interface{})
		var m echo.Map
		err = json.Unmarshal(v, &m)
		if err != nil {
			config.Log.Error("Failed to create DB Config JSON", err)
			return nil, err
		}
		c[k] = m
		parseData = append(parseData, c)
	}

	return json.MarshalIndent(parseData, "", "  ")
}

// UpdateDBConfigJSON - UpdateDBConfigJSON util
func UpdateDBConfigJSON() (err error) {
	dbcJSON, err := GetDBConfigJSON()
	if err != nil {
		config.Log.Error("Failed to create DB Config JSON", err)
		return err
	}

	err = os.WriteFile(filepath.Join(config.Paths.DataDir, constants.DBConfigJSONFileName), dbcJSON, 0666)
	if err != nil {
		config.Log.Error("Failed to update DB Config JSON file", err)
		return err
	}

	return
}

// IsSupportedFileUploadExtension - IsSupportedFileUploadExtension util
func IsSupportedFileUploadExtension(extension string) bool {
	trimmedExtension := strings.TrimLeft(extension, ".")
	isSupported := false
	supportedExtensions := append(
		config.Other.AdditionalFileUploadExtensions,
		constants.SupportedUploadExtensions...,
	)

	for _, supportedExtension := range supportedExtensions {
		if trimmedExtension == supportedExtension {
			isSupported = true
			break
		}
	}
	return isSupported
}

// WriteJSONToFile - WriteJSONToFile util
func WriteJSONToFile(jsonParsed *gabs.Container, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(jsonParsed.StringIndent("", "  "))
	if err != nil {
		return err
	}

	return nil
}
