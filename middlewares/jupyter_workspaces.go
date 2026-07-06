// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
	"github.com/heavyai/webserver/internal/util"
)

type jupyterNotebookParams struct {
	NotebookName string
	SQL          string
}

type jupyterSessionFileParams struct {
	SessionID heavy.TSessionId
}

// Time format, not a hardcoded string - See https://golang.org/src/time/format.go
var jupyterNotebookNameFormat = "HeavyDB_2006-01-02_15:04:05.ipynb"
var jupyterNotebookTemplateText = `{
	"name": "{{.NotebookName}}",
	"path": "{{.NotebookName}}",
	"type": "notebook",
	"format": "json",
	"content": {
		"cells": [
			{
				"cell_type": "code",
				"execution_count": null,
				"metadata": {
					"trusted": true
				},
				"outputs": [],
				"source": [
					"# An Ibis connection object (con) is created on notebook startup, which\n",
					"# includes a pymapd.connection object as a property (con.con).\n",
					"# If you receive a session invalid or object not found error using it,\n",
					"# please close the Jupyter Lab browser tab, relaunch from Immerse,\n",
					"# and run this cell to recreate your con object using the\n",
					"# heavydb function.\n",
					"con = omnisci_connect()\n",
					{{ if .SQL -}}
						"o = con.sql(\"\"\"{{ .EscapedSQL }}\"\"\")\n",
						"o.execute()"
					{{- else -}}
						"con.list_tables()"
					{{- end }}
				]
			}
		],
		"metadata": {
			"kernelspec": {
				"display_name": "Python 3",
				"language": "python",
				"name": "python3"
			},
			"language_info": {
				"codemirror_mode": {
					"name": "ipython",
					"version": 3
				},
				"file_extension": ".py",
				"mimetype": "text/x-python",
				"name": "python",
				"nbconvert_exporter": "python",
				"pygments_lexer": "ipython3",
				"version": "3.7.3"
			}
		},
		"nbformat": 4,
		"nbformat_minor": 4
	}
}`

var jupyterSessionFileTemplateText = `{
	"name": ".omniscisession",
	"path": ".omniscisession",
	"type": "file",
	"format": "text",
	"content": "{\"session\": \"{{.SessionID}}\"}"
}`

// Function copied from https://golang.org/src/net/http/httputil/reverseproxy.go source
func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// Function partially from https://golang.org/src/net/http/httputil/reverseproxy.go source
func cloneRequest(r *http.Request) *http.Request {
	outreq := r.WithContext(r.Context())
	if r.ContentLength == 0 {
		outreq.Body = nil
	}

	outreq.Header = cloneHeader(r.Header)

	return outreq
}

func cloneJupyterRequest(r *http.Request) *http.Request {
	// Create a copy of the incoming request and point it towards Jupyter
	newReq := cloneRequest(r)
	newReq.URL.Scheme = config.Jupyter.JupyterURL.Scheme
	newReq.URL.Host = config.Jupyter.JupyterURL.Host

	return newReq
}

// Double-escape any double quotes in SQL, since it will be placed
// inside both the Ibis string value in the notebook as well as
// the JSON string value of the notebook JSON document
func (p jupyterNotebookParams) EscapedSQL() string {
	json, err := json.Marshal(p.SQL)

	if err != nil {
		config.Log.Fatalln("Errors parsing SQL input to Jupyter notebook:", err)
	}

	jsonString := string(json)
	innerJSONString := jsonString[1 : len(jsonString)-1]
	escapedInnerJSONString := strings.Replace(innerJSONString, `\"`, `\\\"`, -1)

	return escapedInnerJSONString
}

func checkForJupyterError(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.New(r.Status)
	}

	json, err := gabs.ParseJSON(body)
	if err != nil {
		return errors.New(r.Status)
	}

	msg, ok := json.Path("message").Data().(string)
	if ok {
		return errors.New(msg)
	}
	return errors.New(r.Status)
}

func createNewNotebook(r *http.Request, jupyterNotebookTemplate *template.Template, sessionID heavy.TSessionId, username string, sql string) error {
	notebookName := time.Now().Format(jupyterNotebookNameFormat)

	// Create new notebook
	notebookParams := jupyterNotebookParams{
		NotebookName: notebookName,
		SQL:          sql,
	}

	var notebookBuffer bytes.Buffer
	jupyterNotebookTemplate.Execute(&notebookBuffer, notebookParams)

	// For this and following requests, use the incoming request to clone from and reuse auth headers
	notebookCreateReq := cloneJupyterRequest(r)
	notebookCreateReq.Method = http.MethodPut
	notebookCreateReq.URL.Path = config.Jupyter.JupyterPrefix + "/user/" + username + "/api/contents/" + notebookName
	notebookCreateReq.Body = io.NopCloser(&notebookBuffer)

	notebookCreateResp, err := http.DefaultTransport.RoundTrip(notebookCreateReq)
	if err != nil {
		return err
	}
	defer notebookCreateResp.Body.Close()

	err = checkForJupyterError(notebookCreateResp)
	if err != nil {
		return err
	}

	// Get the Jupyter workspace definition

	// The original request is already to what we want - just clone and send it
	// Later on, the original will go through the main reverse proxy and get the updated one we put
	getWorkspaceRequest := cloneJupyterRequest(r)

	workspaceResp, err := http.DefaultTransport.RoundTrip(getWorkspaceRequest)
	if err != nil {
		return err
	}
	defer workspaceResp.Body.Close()

	if workspaceResp.StatusCode > 299 {
		return errors.New("Getting workspace definition failed with status " + workspaceResp.Status)
	}

	workspaceBodyBytes, err := io.ReadAll(workspaceResp.Body)
	if err != nil {
		return err
	}

	workspace, err := gabs.ParseJSON(workspaceBodyBytes)
	if err != nil {
		return err
	}

	// Modify workspace to have our new notebook open in a tab and focused
	notebookRef := "notebook:" + notebookName
	workspace.ArrayAppend(notebookRef, "data", "layout-restorer:data", "main", "dock", "widgets")
	workspace.Set(notebookRef, "data", "layout-restorer:data", "main", "current")
	workspace.Set(notebookName, "data", notebookRef, "data", "path")
	workspace.Set("Notebook", "data", notebookRef, "data", "factory")

	workspaceString := workspace.String()

	// Now put our modified workspace back again
	putWorkspaceRequest := cloneJupyterRequest(r)
	putWorkspaceRequest.Method = http.MethodPut
	putWorkspaceRequest.Header["Content-Type"] = []string{"application/json;charset=utf-8"}
	putWorkspaceRequest.Header["Content-Length"] = []string{strconv.Itoa(len(workspaceString))}
	putWorkspaceRequest.ContentLength = int64(len(workspaceString))
	putWorkspaceRequest.Body = io.NopCloser(strings.NewReader(workspaceString))

	putWorkspaceResp, err := http.DefaultTransport.RoundTrip(putWorkspaceRequest)
	if err != nil {
		return err
	}
	defer putWorkspaceResp.Body.Close()

	err = checkForJupyterError(putWorkspaceResp)
	if err != nil {
		return err
	}

	return nil
}

func createJupyterSessionFile(r *http.Request, jupyterSessionFileTemplate *template.Template, sessionID heavy.TSessionId, username string) error {
	sessionParams := jupyterSessionFileParams{
		SessionID: sessionID,
	}

	var sessionBuffer bytes.Buffer
	jupyterSessionFileTemplate.Execute(&sessionBuffer, sessionParams)

	// clone incoming request to reuse auth headers
	sessionCreateReq := cloneJupyterRequest(r)
	sessionCreateReq.Method = http.MethodPut
	sessionCreateReq.URL.Path = config.Jupyter.JupyterPrefix + "/user/" + username + "/api/contents/.jupyterscratch/.omniscisession"
	sessionCreateReq.Header["Content-Length"] = []string{strconv.Itoa(sessionBuffer.Len())}
	sessionCreateReq.ContentLength = int64(sessionBuffer.Len())
	sessionCreateReq.Body = io.NopCloser(&sessionBuffer)

	sessionCreateResp, err := http.DefaultTransport.RoundTrip(sessionCreateReq)
	if err != nil {
		return err
	}
	defer sessionCreateResp.Body.Close()

	err = checkForJupyterError(sessionCreateResp)
	if err != nil {
		return err
	}

	return nil
}

// JupyterWorkspaces - Jupyter Workspaces Middleware
// This is for the request Hub sends to get the user's visible workspace (open notebooks)
func JupyterWorkspaces(next echo.HandlerFunc) echo.HandlerFunc {
	jupyterNotebookTemplate, err := template.New("jupyter-notebook").Parse(jupyterNotebookTemplateText)
	if err != nil {
		config.Log.Fatalln("Error parsing Jupyter notebook template: ", err)
	}

	jupyterSessionFileTemplate, err := template.New("jupyter-session").Parse(jupyterSessionFileTemplateText)
	if err != nil {
		config.Log.Fatalln("Error parsing Jupyter session file template: ", err)
	}

	return func(ctx echo.Context) error {
		if ctx.Request().Method != "GET" {
			return next(ctx)
		}

		// Retrieve the cookies we set earlier
		newNotebookCookie, err := ctx.Cookie("newnotebook")
		if err == nil && newNotebookCookie.Value == "true" {
			// We need to "protect" this bit of code, but we don't want to protect
			// all requests to this endpoint. Therefore, instead of adding the
			// ProtectRoute middleware in the route chain, we explicitly call it
			// here. Little hacky, but it works.
			protectedRoute := ProtectRoute(func(ctx echo.Context) error {
				newNotebookClearCookie := &http.Cookie{
					Name:     "newnotebook",
					Value:    "",
					Path:     "/",
					Expires:  time.Unix(0, 0),
					MaxAge:   -1,
					HttpOnly: true,
				}
				ctx.SetCookie(newNotebookClearCookie)

				sql := ""
				newNotebookSQLCookie, err := ctx.Cookie("newnotebooksql")
				if err == nil {
					sql, err = url.QueryUnescape(newNotebookSQLCookie.Value)

					if err != nil {
						return echo.NewHTTPError(http.StatusInternalServerError, "Failed to escape SQL")
					}

					newNotebookSQLClearCookie := &http.Cookie{
						Name:     "newnotebooksql",
						Value:    "",
						Path:     "/",
						Expires:  time.Unix(0, 0),
						MaxAge:   -1,
						HttpOnly: true,
					}
					ctx.SetCookie(newNotebookSQLClearCookie)
				}

				// Create new notebook
				sessionID, err := util.GetSessionID(ctx)
				if err != nil || sessionID == "" {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get session ID from request", err)
				}
				username := ctx.Param("username")

				// create session file
				createErr := createJupyterSessionFile(ctx.Request(), jupyterSessionFileTemplate, sessionID, username)
				if createErr != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error creating session file: %s", createErr.Error()))
				}

				// create new notebook
				createErr = createNewNotebook(ctx.Request(), jupyterNotebookTemplate, sessionID, username, sql)
				if createErr != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error creating notebook: %s", createErr.Error()))
				}

				return nil
			})

			if err := protectedRoute(ctx); err != nil {
				return err
			}
		}

		return next(ctx)
	}
}
