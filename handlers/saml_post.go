// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

type samlRequestBody struct {
	SAMLResponse string
	RelayState   string
}

type processedSAMLBody struct {
	sessionID    heavy.TSessionId
	targetPath   string
	samlUsername string
}

var httpClient = &http.Client{
	Timeout: config.HTTP.ConnTimeout,
}

func processSAMLResponseBody(samlReqBody samlRequestBody) (processed processedSAMLBody, err error) {
	// Some IdPs add new lines to the base64 SAMLResponse so we must replace those.
	replacer := strings.NewReplacer("\n", "", "\r", "")
	b64ResponseXML := samlReqBody.SAMLResponse
	b64CleanedXML := replacer.Replace(b64ResponseXML)
	parsedSAMLResponse, err := util.ParseEncodedResponse(b64CleanedXML)
	if err != nil {
		return
	}
	SAMLUsername := parsedSAMLResponse.Assertion.Subject.NameID.Value

	sessionID, err := util.CreateDBSession(models.AuthCredentials{
		Username: "",
		Password: b64CleanedXML,
		DBName:   "",
	})
	if err != nil {
		return
	}

	return processedSAMLBody{
		sessionID:    sessionID,
		samlUsername: SAMLUsername,
	}, err
}

// SamlPost - SAML Post Handler
// samlPostHandler receives a XML(base64-encoded) SAML payload from an identity provider (e.g. Okta) and
// then makes a connect call to HeavyDB with the payload in the JSON boody. If the call succeeds
// we then create a new AuthToken w/ the returned SessionID and set that in the auth cookie, redirecting to
// the Immerse application.
func SamlPost(ctx echo.Context) (err error) {
	redirectPath := "/"

	// Read body without using echo; binding with echo v4.9.0 returns empty body
	body, err := io.ReadAll(ctx.Request().Body)

	if err != nil {
		return doErrRedirect(ctx, err)
	}

	bdy := string(body)
	values, err := url.ParseQuery(bdy)

	if err != nil {
		return doErrRedirect(ctx, err)
	}

	relayState, ok := values["RelayState"]

	if !ok || len(relayState) == 0 {
		return doErrRedirect(ctx, errors.New("Could not get RelayState from SAML payload"))
	}

	samlResponse, ok := values["SAMLResponse"]

	if !ok || len(samlResponse) == 0 {
		return doErrRedirect(ctx, errors.New("Could not get SAMLResponse from SAML payload"))
	}

	samlReqBody := samlRequestBody{RelayState: relayState[0], SAMLResponse: samlResponse[0]}

	// Replace the request body
	ctx.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	processed, err := processSAMLResponseBody(samlReqBody)
	if err != nil {
		return doErrRedirect(ctx, err)
	}

	if samlReqBody.RelayState != "" {
		samlReqBody.RelayState = path.Clean(samlReqBody.RelayState)
	}

	sessionID := processed.sessionID
	redirectPath = samlReqBody.RelayState
	samlUsername := processed.samlUsername

	thriftSessionInfo, err := util.GetThriftSessionInfo(sessionID)
	if err != nil {
		doErrRedirect(ctx, errors.New("Could not get Thrift session info for sessionID: "+err.Error()))
	}

	dbName := thriftSessionInfo.Database
	relayDBContext := util.URLPathExtractDBContext(samlReqBody.RelayState)

	// If we're given a dbName via the SAML relay state, then first check to see if it's different from the Database
	//  that the initial connect call sessionID is associated w/.  If so, verify that the user has access to the relayed
	//  DB, and if so, exchange the sessionID to the relayed DB.
	if len(relayDBContext) > 0 {
		if relayDBContext != thriftSessionInfo.Database {
			if util.InaccessibleDB(&models.DBSessionInfo{
				SessionID: sessionID,
			}, dbName) {
				return doErrRedirect(ctx, errors.New("Inaccessible or nonexistent database: "+dbName))
			}
			sessionID, thriftSessionInfo, err = util.ExchangeSessionID(sessionID, relayDBContext)
			if err != nil {
				return doErrRedirect(ctx, errors.New("Could not exchange sessionID to database '"+dbName+"': "+err.Error()))
			}
			dbName = relayDBContext
		}
	}

	// TODO: Either fetch session timeout info from thrift to retrieve timeout values, or else mitigate out of sync
	//   timeouts by taking into account session creation latency to calculate session timeouts here.
	//    https://heavyai.atlassian.net/browse/FE-10542
	idleSessionExpiry := time.Now().Add(config.Session.IdleTimeout)
	maxSessionExpiry := time.Now().Add(config.Session.MaxTimeout)
	immerseCtx := ctx.(*models.ImmerseReqContext)
	immerseCtx.Claims.Username = samlUsername
	immerseCtx.DBName = dbName

	token, err := immerseCtx.Claims.AddSessionToToken(samlUsername, sessionID, dbName, idleSessionExpiry, maxSessionExpiry)
	if err != nil {
		return doErrRedirect(ctx, err)
	}

	util.SetAuthCookie(
		ctx,
		token,
		time.Now().Add(config.Session.MaxTimeout),
	)
	util.DeleteDocumentCookie(ctx, constants.CookieLogout)

	// If no redirectPath is set on the SamlRequestBody XML, then we should redirect w/ the DB context that we stored on
	//  on the JWT.
	if redirectPath == "/" {
		redirectPath = dbName + "/dashboards"
	}

	return ctx.Redirect(http.StatusMovedPermanently, redirectPath)
}

func doErrRedirect(ctx echo.Context, err error) error {
	config.Log.Error("Error authenticating user via SAML: ", err.Error())
	return ctx.Redirect(http.StatusSeeOther, constants.SamlErrorPage)
}
