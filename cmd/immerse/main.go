// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/heavyai/webserver/handlers"
	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/constants"
	"github.com/heavyai/webserver/middlewares"
)

func main() {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = handlers.Immerse404

	// Middleware
	e.Use(
		middlewares.ExtendContext,
		middleware.LoggerWithConfig(middlewares.CustomLoggerConfig),
		middleware.Recover(),
		middlewares.DynamicCORSMiddleware,
		middleware.GzipWithConfig(middlewares.CustomCompressConfig),
		middlewares.StripRequestXHeaders(config.Other.StripXHeaders),
		middlewares.RedirectSAML,
		middlewares.SecureHeaders,
	)

	// Routes - Root
	e.POST(
		"/",
		handlers.Noop,
		middlewares.ThriftOnly,
		middlewares.ThriftExtend,
		middlewares.ThriftAuthEncryption,
		middlewares.ThriftAuthorization,
		middlewares.ThriftSessionIdleExpiryBump,
		middlewares.ACAOHeaderOverwrite,
		middlewares.ThriftProxy(),
	)
	if config.Enable.CustomIntegrationAuth {
		e.GET("/", handlers.AppIndex, middlewares.AppIndexSessionExchange, middlewares.QueryParamAuth)
	} else {
		e.GET("/", handlers.AppIndex, middlewares.AppIndexSessionExchange)
	}
	e.GET("/index.html", handlers.AppIndex)
	e.Static("/*", config.Paths.Frontend)
	e.Static("/docs", config.Paths.DocsDir)
	e.GET("/servers.json", handlers.ServersJSON)
	e.GET("/health", handlers.Health)
	e.GET("/version.txt", handlers.Version)

	// Routes - Session Management
	sessionGroup := e.Group("/session")
	sessionGroup.POST("/create", handlers.CreateSession)
	sessionGroup.POST("/exchange", handlers.ExchangeSession, middlewares.ProtectRoute)
	sessionGroup.DELETE("/destroy", handlers.DestroySession, middlewares.ProtectRoute)
	sessionGroup.DELETE("/destroy/all", handlers.DestroyAllSessions, middlewares.ProtectRoute)
	sessionGroup.GET("/validate", handlers.ValidateSession)
	sessionGroup.GET("/durations", handlers.SessionDurations, middlewares.ProtectRoute)
	sessionGroup.GET("/durations/current", handlers.SessionDurationsCurrent)
	sessionGroup.GET("/db-access-list", handlers.SessionGetDatabases, middlewares.ProtectRoute)
	sessionGroup.GET("/roles", handlers.FetchRoles, middlewares.ProtectRoute)
	sessionGroup.GET("/roles/:"+constants.RoleRouteParamName, handlers.FetchRoles, middlewares.ProtectRoute)
	sessionGroup.GET("/privileges", handlers.FetchSessionPrivileges, middlewares.ProtectRoute)
	sessionGroup.GET("/privileges/:"+constants.DBNameRouteParamName, handlers.FetchSessionPrivileges, middlewares.ProtectRoute)
	sessionGroup.GET("/privileges/table/:"+constants.TableNameRouteParamName, handlers.FetchTablePrivileges, middlewares.ProtectRoute)

	// Routes - Thrift Alternates
	e.GET("/thrift/interrupt", handlers.Interrupt, middlewares.ProtectRoute)

	// Routes - File Management
	e.POST("/upload", handlers.Upload, middlewares.ProtectRoute)
	e.POST("/export", handlers.Export, middlewares.ProtectRoute)

	// Dashboards - Dashboard Management
	e.POST("/dashboards/export", handlers.DashboardsExport, middlewares.ProtectRoute)

	// Routes - Configuration Management
	configRoutes := e.Group("/configuration")
	configRoutes.POST("/db", handlers.ConfigurationDBCreate, middlewares.ProtectRoute, middlewares.AuthorizeRoute(constants.DBConfigRole))
	configRoutes.GET("/db", handlers.ConfigurationDBRead, middlewares.ProtectRoute)
	configRoutes.POST("/instance", handlers.ConfigurationInstanceCreate, middlewares.ProtectRoute)
	configRoutes.GET("/instance", handlers.ConfigurationInstanceRead)
	configRoutes.GET("/supported-data-sources", handlers.ConfigurationSupportedDataSourcesHandler, middlewares.ProtectRoute)

	tables := e.Group("/tables")
	tables.Use(middlewares.ProtectRoute)
	tables.POST("/detect-column-types", handlers.DetectColumnTypesHandler)
	tables.POST("/create/connect/odbc", handlers.CreateOdbcFSITableHandler)
	tables.POST("/create/connect/s3", handlers.CreateS3FSITableHandler)
	tables.POST("/create/connect/server-file", handlers.CreateServerFileFSITableHandler)
	tables.PATCH(
		"/connected/:"+constants.TableNameRouteParamName+"/rename/:"+constants.NewTableNameRouteParamName,
		handlers.RenameForeignTableHandler,
	)
	tables.PATCH(
		"/connected/:"+constants.TableNameRouteParamName+"/rename-columns",
		handlers.RenameForeignTableColumnsHandler,
	)
	tables.PATCH(
		"/connected/:"+constants.TableNameRouteParamName+"/refresh-schedule",
		handlers.UpdateForeignTableRefreshScheduleHandler,
	)
	tables.GET(
		"/connected/:"+constants.TableNameRouteParamName+"/refresh",
		handlers.RefreshTableHandler,
	)
	tables.DELETE("/connected/:"+constants.TableNameRouteParamName, handlers.DropForeignTableHandler)
	tables.PATCH("/:"+constants.TableNameRouteParamName+"/table-comment", handlers.SetTableCommentHandler)
	tables.PATCH("/:"+constants.TableNameRouteParamName+"/column-comment", handlers.SetColumnCommentHandler)
	tables.DELETE("/:"+constants.TableNameRouteParamName+"/table-comment", handlers.DeleteTableCommentHandler)
	tables.DELETE("/:"+constants.TableNameRouteParamName+"/column-comment", handlers.DeleteColumnCommentHandler)

	// Routes - IQ Service
	if config.Enable.IQ {
		iqRoutes := e.Group("/iq")
		iqRoutes.POST("/query", handlers.IQQueryProxy, middlewares.ProtectRoute)
		iqRoutes.POST("/auto/query", handlers.IQAutoQueryProxy, middlewares.ProtectRoute)
		iqRoutes.POST("/question", handlers.IQQuestionProxy, middlewares.ProtectRoute)
		iqRoutes.POST("/answer", handlers.IQAnswerProxy, middlewares.ProtectRoute)
		iqRoutes.POST("/submit-feedback", handlers.IQFeedbackProxy, middlewares.ProtectRoute)

		snippets := e.Group("/guidance")
		snippets.POST("/get", handlers.SnippetGetProxy, middlewares.ProtectRoute)
		snippets.POST("/list", handlers.SnippetListProxy, middlewares.ProtectRoute)
		snippets.POST("/insert", handlers.SnippetInsertProxy, middlewares.ProtectRoute, middlewares.AuthorizeRoute(constants.GuidanceManageRole))
		snippets.POST("/update", handlers.SnippetUpdateProxy, middlewares.ProtectRoute, middlewares.AuthorizeRoute(constants.GuidanceManageRole))
		snippets.POST("/delete", handlers.SnippetDeleteProxy, middlewares.ProtectRoute, middlewares.AuthorizeRoute(constants.GuidanceManageRole))
	}

	if config.Enable.Transcription {
		e.POST("/transcribe", handlers.TranscribeProxyHandler, middlewares.ACAOHeaderOverwrite, middlewares.ProtectRoute)
	}

	e.GET("/query-history", handlers.GetQueryHistory, middlewares.ProtectRoute)
	e.POST("/query-history", handlers.AddQueryHistory, middlewares.ProtectRoute)

	// Routes - Logs
	if config.Enable.BrowserLogs {
		logs := e.Group("/logs")
		if !config.Enable.Development {
			logs.Use(middlewares.ProtectRoute)
		}
		logs.GET("/all", handlers.Logs(config.Paths.AllLogFile))
		logs.GET("/access", handlers.Logs(config.Paths.AccessLogFile))
		logs.GET("/error", handlers.Logs(config.Paths.ErrorDBLogFile))
		logs.GET("/info", handlers.Logs(config.Paths.InfoDBLogFile))
		logs.GET("/warning", handlers.Logs(config.Paths.WarningDBLogFile))

		if config.Enable.IQ {
			logs.GET("/iq/access", handlers.Logs(config.Paths.AccessIQLogFile))
			logs.GET("/iq/all", handlers.Logs(config.Paths.AllIQLogFile))
			logs.GET("/iq/build", handlers.Logs(config.Paths.BuildIQLogFile))
			logs.GET("/iq/console", handlers.Logs(config.Paths.ConsoleIQLogFile))
			logs.GET("/iq/guidance", handlers.Logs(config.Paths.GuidanceIQLogFile))
			logs.GET("/iq/chroma", handlers.Logs(config.Paths.ChromaIQLogFile))
		}
	}

	// Routes - Single Customer Custom Integration
	if config.Enable.CustomIntegrationAuth {
		e.Match(constants.AllowedHTTPMethods, "/_internal/clear-servers-json", handlers.ClearServersJSON)
		e.Match(constants.AllowedHTTPMethods, "/_internal/set-servers-json", handlers.SetServersJSON)
	}

	// Routes - Integrations
	if config.Enable.LocalDataCatalog {
		dataCatalog := e.Group("/data-catalog")
		dataCatalog.Use(middlewares.DataCatalogCache, middleware.Static(config.Paths.DataCatalogDir))

		e.POST("/data-catalog-upload", handlers.DataCatalog, middlewares.ProtectRoute)
	} else {
		e.GET("/data-catalog/*", handlers.DataCatalog404)
	}

	if config.Enable.SAML {
		e.POST("/saml-post", handlers.SamlPost)
	}

	if config.Enable.Jupyter {
		jupyterProxy := e.Group(config.Jupyter.JupyterPrefix)
		jupyterProxy.GET(
			"/hub/login",
			handlers.Noop,
			middlewares.ProtectRoute,
			middlewares.JupyterLogin,
			middlewares.ReverseProxy(config.Jupyter.JupyterURL),
		)
		jupyterProxy.Match(
			constants.AllowedHTTPMethods,
			"/user/:username/lab/api/workspaces/default",
			handlers.Noop,
			middlewares.JupyterWorkspaces,
			middlewares.ReverseProxy(config.Jupyter.JupyterURL),
		)
		jupyterProxy.Match(
			constants.AllowedHTTPMethods,
			"/user/:username/lab/api/workspaces/lab",
			handlers.Noop,
			middlewares.JupyterWorkspaces,
			middlewares.ReverseProxy(config.Jupyter.JupyterURL),
		)
		jupyterProxy.Match(
			constants.AllowedHTTPMethods,
			"/hub/*",
			handlers.Noop,
			middlewares.ReverseProxy(config.Jupyter.JupyterURL),
			middlewares.EventProxy(config.Jupyter.JupyterURL),
		)
		jupyterProxy.Match(
			constants.AllowedHTTPMethods,
			"/user/*",
			handlers.Noop,
			middlewares.ReverseProxy(config.Jupyter.JupyterURL),
		)
		jupyterProxy.Match(
			constants.AllowedHTTPMethods,
			"*",
			handlers.Noop,
			middlewares.ReverseProxy(config.Jupyter.JupyterURL),
		)
	}

	if config.Enable.JupyterDesktop {
		e.GET("/jupyter-desktop", handlers.JupyterDesktop)
	}

	if config.Enable.Geocoder {
		e.GET("/geocoder", handlers.Geocoder, middlewares.ProtectRoute)
	}

	// Routes - Reverse Proxies
	if config.Enable.ReverseProxies {
		for _, proxy := range config.Other.Proxies {
			e.Any(proxy.Path, handlers.ReverseProxy(proxy))
			e.Any(proxy.Path+"/*", handlers.ReverseProxy(proxy))
		}
	}

	// Server Config + Start
	if config.Enable.HTTPSRedirect {
		go configureRedirectListener()
	}

	go func() {
		if config.Enable.HTTPS {
			tlsServer, err := configureSecureServer(e)
			if err != nil {
				config.Log.Fatal("Failed to configure TLS server: ", err)
			}
			config.Log.Fatal(e.StartServer(tlsServer))
		} else {
			e.Server = configureInsecureServer(e)
			config.Log.Fatal(e.Start(config.HTTP.Addr))
		}
	}()

	configureGracefulShutdown(e, 5*time.Second)
}
