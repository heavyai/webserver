// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/models"
)

// findFailedSQL extracts and returns the SQL query following "Failed SQL: " and the error message
// preceding "Failed SQL: "
// If "Failed SQL: " is not found, it returns an empty strings and the original error message
func findFailedSQL(input string) (string, string) {
	index := strings.Index(input, "Failed SQL: ")
	if index == -1 {
		return "", input
	}

	start := index + len("Failed SQL: ")

	return input[start:], input[:index]
}

// IQQueryProxy - IQQueryProxy Handler
// Proxies query requests to the IQ service
func IQQueryProxy(ctx echo.Context) error {
	// Get session id
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}
	userSessionID := sessionInfo.SessionID

	// Get request params
	var iqQuery models.IQQuery
	err = ctx.Bind(&iqQuery)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Bad request")
		return err
	}

	// Compose IQQueryDTO
	iqqDTO := models.IQQueryDTO{
		Tables:    iqQuery.Tables,
		Question:  iqQuery.Question,
		SessionID: string(userSessionID),
	}

	// Marshal into JSON and send POST request to IQ service
	iqqJSON, err := json.Marshal(iqqDTO)
	if err != nil {
		config.Log.Error("Could not marshal IQQueryDTO", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	resp, err := http.Post(config.Paths.IQServiceURL.String()+"/api/v1/lcel/query", "application/json",
		bytes.NewBuffer(iqqJSON))
	if err != nil {
		config.Log.Error("Could not make POST request to IQ Service", err)
		ctx.NoContent(http.StatusBadGateway)
		return err
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		config.Log.Error("Could not read IQ Service response", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	// Successful response case
	if resp.StatusCode == 200 {
		// Unmarshal response body into IQQueryResponseDTO
		var iqqrDTO models.IQQueryResponseDTO
		err = json.Unmarshal(body, &iqqrDTO)
		if err != nil {
			config.Log.Error("Could not unmarshal IQQueryResponseDTO", err)
			ctx.String(http.StatusInternalServerError, "IQ Server Error")
			return err
		}

		// Return IQQueryResponseDTO
		return ctx.JSON(http.StatusOK, iqqrDTO)

		// Error response - Service Unavailable
	} else if resp.StatusCode == 503 {
		config.Log.Error("IQ Service Unavailable", err)
		ctx.String(http.StatusInternalServerError, "IQ Service Unavailable")
		return err

		// Error response - generic errors
	} else {
		// Unmarshal response body into IQErrorResponseDTO
		var iqError models.IQErrorResponseDTO
		err = json.Unmarshal(body, &iqError)
		if err != nil {
			config.Log.Error("Could not unmarshal IQErrorResponseDTO", err)
			ctx.String(http.StatusInternalServerError, "IQ Server Error")
			return err
		}

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
}

// IQAutoQueryProxy - IQAutoQueryProxy Handler
// Proxies auto query requests to the IQ service
func IQAutoQueryProxy(ctx echo.Context) error {
	// Get session id
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}
	userSessionID := sessionInfo.SessionID

	// Get request params
	var iqAutoQuery models.IQAutoQueryDTO
	err = ctx.Bind(&iqAutoQuery)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Bad request")
		return err
	}

	// Compose IQQueryDTO
	iqaqDTO := models.IQAutoQueryDTO{
		AllowedTables: iqAutoQuery.AllowedTables,
		Question:      iqAutoQuery.Question,
		SessionID:     string(userSessionID),
	}

	// Marshal into JSON and send POST request to IQ service
	iqaqJSON, err := json.Marshal(iqaqDTO)
	if err != nil {
		config.Log.Error("Could not marshal IQAutoQueryDTO", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	resp, err := http.Post(config.Paths.IQServiceURL.String()+"/api/v1/lcel/auto/query", "application/json",
		bytes.NewBuffer(iqaqJSON))
	if err != nil {
		config.Log.Error("Could not make POST request to IQ Service", err)
		ctx.NoContent(http.StatusBadGateway)
		return err
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		config.Log.Error("Could not read IQ Service response", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	// Successful response case
	if resp.StatusCode == 200 {
		// Unmarshal response body into IQQueryResponseDTO
		var iqaqrDTO models.IQAutoQueryResponseDTO
		err = json.Unmarshal(body, &iqaqrDTO)
		if err != nil {
			config.Log.Error("Could not unmarshal IQAutoQueryResponseDTO", err)
			ctx.String(http.StatusInternalServerError, "IQ Server Error")
			return err
		}

		// Return IQQueryResponseDTO
		return ctx.JSON(http.StatusOK, iqaqrDTO)

		// Error response - Service Unavailable
	} else if resp.StatusCode == 503 {
		config.Log.Error("IQ Service Unavailable", err)
		ctx.String(http.StatusInternalServerError, "IQ Service Unavailable")
		return err

		// Error response - generic errors
	} else {
		// Unmarshal response body into IQErrorResponseDTO
		var iqError models.IQErrorResponseDTO
		err = json.Unmarshal(body, &iqError)
		if err != nil {
			config.Log.Error("Could not unmarshal IQErrorResponseDTO", err)
			ctx.String(http.StatusInternalServerError, "IQ Server Error")
			return err
		}

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
}

// IQQuestionProxy - IQQuestionProxy Handler
// Proxies question requests to the IQ service
func IQQuestionProxy(ctx echo.Context) error {
	// Get session id
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}
	userSessionID := sessionInfo.SessionID

	// Get request params
	var iqQuestion models.IQQuestion
	err = ctx.Bind(&iqQuestion)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Bad request")
		return err
	}

	// Compose IQQuestionDTO
	iqqDTO := models.IQQuestionDTO{
		Tables:    iqQuestion.Tables,
		Question:  iqQuestion.Question,
		SessionID: string(userSessionID),
	}

	// Marshal into JSON and send POST request to IQ service
	iqqJSON, err := json.Marshal(iqqDTO)
	if err != nil {
		config.Log.Error("Could not marshal IQQuestionDTO", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	resp, err := http.Post(config.Paths.IQServiceURL.String()+"/api/v1/lcel/question", "application/json",
		bytes.NewBuffer(iqqJSON))
	if err != nil {
		config.Log.Error("Could not make POST request to IQ Service", err)
		ctx.NoContent(http.StatusBadGateway)
		return err
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		config.Log.Error("Could not read IQ Service response", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	// Successful response
	if resp.StatusCode == 200 {
		// Unmarshal response body into IQQuestionResponseDTO
		var iqqrDTO models.IQQuestionResponseDTO
		err = json.Unmarshal(body, &iqqrDTO)
		if err != nil {
			config.Log.Error("Could not unmarshal IQQuestionResponseDTO", err)
			ctx.String(http.StatusInternalServerError, "IQ Service Error")
			return err
		}

		// Return IQQuestionResponseDTO
		return ctx.JSON(http.StatusOK, iqqrDTO)

		// Error response - Service Unavailable
	} else if resp.StatusCode == 503 {
		config.Log.Error("IQ Service Unavailable")
		ctx.String(http.StatusInternalServerError, "IQ Service Unavailable")
		return err

		// Error response - generic errors
	} else {
		// Unmarshal response body into IQErrorResponseDTO
		var iqError models.IQErrorResponseDTO
		err = json.Unmarshal(body, &iqError)
		if err != nil {
			config.Log.Error("Could not unmarshal IQErrorResponseDTO", err)
			ctx.String(http.StatusInternalServerError, "IQ Service Error")
			return err
		}

		// Return IQErrorResponseDTO
		config.Log.Error("IQ Service error response: ", iqError.Error)
		ctx.JSON(http.StatusInternalServerError, iqError)
		return err
	}
}

// IQAnswerProxy - IQAnswerProxy Handler
// Proxies answer requests to the IQ service
func IQAnswerProxy(ctx echo.Context) error {
	// Get session id
	immerseCtx := ctx.(*models.ImmerseReqContext)
	sessionInfo, err := immerseCtx.Claims.SessionsInfoMap.GetDBSessionInfo(immerseCtx.DBName)
	if err != nil {
		config.Log.Error("Session does not exist for given DB context '", immerseCtx.DBName, "': ", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}
	userSessionID := sessionInfo.SessionID

	// Get request params
	var iqAnswer models.IQAnswer
	err = ctx.Bind(&iqAnswer)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Bad request")
		return err
	}

	// Compose IQQuestionDTO
	iqaDTO := models.IQAnswerDTO{
		Tables:    iqAnswer.Tables,
		Query:     iqAnswer.Query,
		Question:  iqAnswer.Question,
		SessionID: string(userSessionID),
	}

	// Marshal into JSON and send POST request to IQ service
	iqaJSON, err := json.Marshal(iqaDTO)
	if err != nil {
		config.Log.Error("Could not marshal IQAnswerDTO", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	resp, err := http.Post(config.Paths.IQServiceURL.String()+"/api/v1/lcel/answer", "application/json",
		bytes.NewBuffer(iqaJSON))
	if err != nil {
		config.Log.Error("Could not make POST request to IQ Service", err)
		ctx.NoContent(http.StatusBadGateway)
		return err
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		config.Log.Error("Could not read IQ Service response", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	// Successful response
	if resp.StatusCode == 200 {
		// Unmarshal response body into IQAnswerResponseDTO
		var iqarDTO models.IQAnswerResponseDTO
		err = json.Unmarshal(body, &iqarDTO)
		if err != nil {
			config.Log.Error("Could not unmarshal IQQuestionResponseDTO", err)
			ctx.String(http.StatusInternalServerError, "IQ Service Error")
			return err
		}

		// Return IQAnswerResponseDTO
		return ctx.JSON(http.StatusOK, iqarDTO)

		// Error response - Service Unavailable
	} else if resp.StatusCode == 503 {
		config.Log.Error("IQ Service Unavailable")
		ctx.String(http.StatusInternalServerError, "IQ Service Unavailable")
		return err

		// Error response - generic errors
	} else {
		// Unmarshal response body into IQErrorResponseDTO
		var iqError models.IQErrorResponseDTO
		err = json.Unmarshal(body, &iqError)
		if err != nil {
			config.Log.Error("Could not unmarshal IQErrorResponseDTO", err)
			ctx.String(http.StatusInternalServerError, "IQ Service Error")
			return err
		}

		// Return IQErrorResponseDTO
		config.Log.Error("IQ Service error response: ", iqError.Error)
		ctx.JSON(http.StatusInternalServerError, iqError)
		return err
	}
}

// IQFeedbackProxy - IQFeedbackProxy Handler
// Proxies feedback for question and query requests to the IQ service
func IQFeedbackProxy(ctx echo.Context) error {
	// Get request params
	var iqFeedbackDTO models.IQFeedbackDTO
	err := ctx.Bind(&iqFeedbackDTO)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Bad request")
		return err
	}

	// Marshal into JSON and send POST request to IQ service
	iqfbJSON, err := json.Marshal(iqFeedbackDTO)
	if err != nil {
		config.Log.Error("Could not marshal IQFeedbackDTO", err)
		ctx.NoContent(http.StatusInternalServerError)
		return err
	}

	resp, err := http.Post(config.Paths.IQServiceURL.String()+"/api/v1/submit-feedback", "application/json",
		bytes.NewBuffer(iqfbJSON))
	if err != nil {
		config.Log.Error("Could not make POST request to IQ Service", err)
		ctx.NoContent(http.StatusBadGateway)
		return err
	}

	// Successful response - Status OK
	if resp.StatusCode == 200 {
		return ctx.NoContent(http.StatusOK)

		// Error response - Service Unavailable
	} else if resp.StatusCode == 503 {
		config.Log.Error("IQ Service Unavailable")
		ctx.String(http.StatusInternalServerError, "IQ Service Unavailable")
		return err

		// Error response - generic errors
	} else {
		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			config.Log.Error("Could not read IQ Service response", err)
			ctx.NoContent(http.StatusInternalServerError)
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

		// Return IQErrorResponseDTO
		config.Log.Error("IQ Service error response: ", iqError.Error)
		ctx.JSON(http.StatusInternalServerError, iqError)
		return err
	}
}
