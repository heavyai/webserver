// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/internal/util"
)

type templateVars struct {
	AppConfigDeclaration     template.JS
	DBConfigDeclaration      template.JS
	MetricsConfigDeclaration template.JS
}

const scriptTemplateStr = `
	<script>
		{{.AppConfigDeclaration}}
		{{.DBConfigDeclaration}}
		{{.MetricsConfigDeclaration}}
	</script>
`

func getTemplateVars(serverConfig []byte) templateVars {
	dbc, err := util.GetDBConfigJSON()
	if err != nil {
		dbc = []byte("{}")
	}

	mc, err := util.GetMetricsConfigJSON()
	if err != nil {
		mc = []byte("{}")
	}

	return templateVars{
		AppConfigDeclaration:     template.JS("window." + constants.AppConfigJSConstName + " = " + string(serverConfig)),
		DBConfigDeclaration:      template.JS("window." + constants.DBConfigJSConstName + " = " + string(dbc)),
		MetricsConfigDeclaration: template.JS("window." + constants.MetricsConfigJSConstName + " = " + string(mc)),
	}
}

func renderRespBody(request http.Request) (respBody string, err error) {
	var respBodyBuffer bytes.Buffer

	if config.Other.IndexHTMLbytes == nil {
		err = errors.New("Could not read index.html for rendering")
		return
	}

	indexHTMLStr := constants.BodyTagRegex.ReplaceAllString(
		string(config.Other.IndexHTMLbytes),
		"$0"+scriptTemplateStr,
	)

	templatePointer := template.New("index.html")
	templatePointer, err = templatePointer.Parse(indexHTMLStr)
	if err != nil {
		config.Log.Error("Could not parse index.html template: ", err)
		return
	}

	serverConfig, err := GetServersJSON(request)
	if err != nil {
		config.Log.Error("Could not get server config: ", err)
		return
	}

	templateVars := getTemplateVars(serverConfig)

	err = templatePointer.ExecuteTemplate(&respBodyBuffer, "index.html", templateVars)
	if err != nil {
		config.Log.Error("Could not render index template: ", err)
		return
	}
	return respBodyBuffer.String(), nil
}

// AppIndex - AppIndex Handler
func AppIndex(ctx echo.Context) (err error) {
	respBody, err := renderRespBody(*ctx.Request())
	if err != nil {
		config.Log.Error("Could not render index request body: ", err)
		return
	}
	return ctx.HTML(
		http.StatusOK,
		respBody,
	)
}
