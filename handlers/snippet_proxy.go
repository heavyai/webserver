// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/models"
)

// iqProxyRequest handles the common logic for sending a request to the IQ service and processing the response
func iqProxyRequest(ctx echo.Context, url string, reqBody interface{}, respBody interface{}) error {
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		config.Log.Error("Could not marshal request body", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		config.Log.Error("Could not make POST request to IQ Service", err)
		return ctx.NoContent(http.StatusBadGateway)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		config.Log.Error("Could not read IQ Service response", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if resp.StatusCode == 200 {
		err = json.Unmarshal(body, respBody)
		if err != nil {
			config.Log.Error("Could not unmarshal response body", err)
			ctx.String(http.StatusInternalServerError, "Server Error")
			return err
		}
		return ctx.JSON(http.StatusOK, respBody)
	} else if resp.StatusCode == 503 {
		config.Log.Error("IQ Service Unavailable")
		ctx.String(http.StatusInternalServerError, "IQ Service Unavailable")
		return err
	}
	// Unmarshal response body into IQErrorResponseDTO
	var iqError models.IQErrorResponseDTO
	err = json.Unmarshal(body, &iqError)
	if err != nil {
		config.Log.Error("Could not unmarshal IQErrorResponseDTO", err)
		ctx.String(http.StatusInternalServerError, "IQ Service Error")
		return err
	}
	// Modify error if contains failed SQL
	failedSQL, errorFailedSQL := findFailedSQL(iqError.Error)
	if failedSQL != "" {
		iqError.SQL = failedSQL
		iqError.Error = errorFailedSQL
	}
	// Return IQErrorResponseDTO
	config.Log.Error("IQ Service error response: ", iqError.Error)
	ctx.JSON(http.StatusInternalServerError, iqError)
	return err
}

// SnippetInsertProxy - Proxies requests to the IQ Snippet service for inserting snippets
func SnippetInsertProxy(ctx echo.Context) error {
	// Transfer parameters
	var req models.SnippetInsertRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.String(http.StatusBadRequest, "Bad request")
	}
	// Set session id on Request
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}
	req.SessionID = string(sessionInfo.SessionID)
	// Additional arguments
	var resp models.SnippetInsertResponse
	url := config.Paths.IQServiceURL.String() + "/rag/snippets/bulk-insert"

	return iqProxyRequest(ctx, url, req, &resp)
}

// SnippetUpdateProxy - Proxies requests to the IQ Snippet service for updating snippets
func SnippetUpdateProxy(ctx echo.Context) error {
	var req models.SnippetUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.String(http.StatusBadRequest, "Bad request")
	}
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}
	req.SessionID = string(sessionInfo.SessionID)
	var resp models.SnippetUpdateResponse
	url := config.Paths.IQServiceURL.String() + "/rag/snippets/update"
	return iqProxyRequest(ctx, url, req, &resp)
}

// SnippetGetProxy - Proxies requests to the IQ Snippet service for getting a snippet
func SnippetGetProxy(ctx echo.Context) error {
	var req models.SnippetGetRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.String(http.StatusBadRequest, "Bad request")
	}
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}
	req.SessionID = string(sessionInfo.SessionID)
	var resp models.SnippetGetResponse
	url := config.Paths.IQServiceURL.String() + "/rag/snippets/get"
	return iqProxyRequest(ctx, url, req, &resp)
}

// SnippetListProxy - Proxies requests to the IQ Snippet service for listing snippets
func SnippetListProxy(ctx echo.Context) error {
	var req models.SnippetListRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.String(http.StatusBadRequest, "Bad request")
	}
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}
	req.SessionID = string(sessionInfo.SessionID)
	var resp models.SnippetListResponse
	url := config.Paths.IQServiceURL.String() + "/rag/snippets/list"
	return iqProxyRequest(ctx, url, req, &resp)
}

// SnippetDeleteProxy - Proxies requests to the IQ Snippet service for deleting snippets
func SnippetDeleteProxy(ctx echo.Context) error {
	var req models.SnippetDeleteRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.String(http.StatusBadRequest, "Bad request")
	}
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}
	req.SessionID = string(sessionInfo.SessionID)
	var resp models.SnippetDeleteResponse
	url := config.Paths.IQServiceURL.String() + "/rag/snippets/delete"
	return iqProxyRequest(ctx, url, req, &resp)
}
