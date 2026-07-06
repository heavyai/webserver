// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

import (
	"net/http"
	"regexp"
	"time"
)

const (
	// CookieModifiedServersJSON - The name of the cookie that holds the modified client-specific servers.json params
	CookieModifiedServersJSON = "servers-json"
	// CookieLogout - The name of the cookie set on logout
	CookieLogout = "owl"
	// SamlErrorPage - The page to redirect the user to when there are errors with SAML auth
	SamlErrorPage = "/saml-error.html"
	// JupyterDesktopToken - Jupyter Desktop Token
	JupyterDesktopToken = "ea436a31623ee0c6ef25f8f340fcf3bdd17f914260dc3e17620ae04f93dfbb3c"
	// LoggerMessageFormat - Logger string template following Common Log Format
	LoggerMessageFormat = "${remote_ip} - - ${time_custom} \"${method} ${uri} ${protocol}\" ${status} ${bytes_out}\n"
	// LoggerTimeFormat - Custom time format following Common Log Format
	LoggerTimeFormat = "[02/Jan/2006:15:04:05 -0700]"
	// AppConfigJSConstName - Name of variable that will push servers.json into index.html for JS in Immerse app
	AppConfigJSConstName = "APP_CONFIG"
	// DBConfigJSConstName - Name of variable that will push db configuration into index.html for JS in Immerse app
	DBConfigJSConstName = "DB_CONFIG"
	// DBConfigJSONFileName - constant
	DBConfigJSONFileName = "immerse_db_config.json"
	// DBConfigRole - constant
	DBConfigRole = "immerse_db_config"
	// GuidanceManageRole - constant
	GuidanceManageRole = "immerse_guidance_manage"
	// InstanceConfigJSONFileName - constant
	InstanceConfigJSONFileName = "immerse_instance_config.json"
	// MetricsConfigJSConstName - Name of variable that will push metrics configuration into index.html for JS in Immerse app
	MetricsConfigJSConstName = "METRICS_CONFIG"
	// RelativeImportPath - Relative Import path
	RelativeImportPath = "import"
	// RelativeExportPath - Relative Export path
	RelativeExportPath = "export"
	// RelativeLogPath - Relative Log path
	RelativeLogPath = "log"
	// HeavyDBUsernameXHeaderName -  HeavyDB username X header name
	HeavyDBUsernameXHeaderName = "X-HeavyDB-Username"
	// ControlPanelAdminRole - Role that allows access to control panel regardless of superuser status
	ControlPanelAdminRole = "immerse_control_panel"

	/**
	Config Defaults
	*/

	// DefaultAuth - The name of the cookie used for auth
	DefaultAuth = "oat"
	// DefaultJupyterPrefix - constant
	DefaultJupyterPrefix = "/jupyter"
	// DefaultCertFilePath - constant
	DefaultCertFilePath = "cert.pem"
	// DefaultDataPath - constant
	DefaultDataPath = "data"
	// DefaultDocsPath - constant
	DefaultDocsPath = "docs"
	// DefaultFrontendPath - constant
	DefaultFrontendPath = "frontend"
	// DefaultHTTPSKeyFilePath - constant
	DefaultHTTPSKeyFilePath = "key.pem"
	// DefaultPeerCertPath - constant
	DefaultPeerCertPath = "peercert.pem"
	// DefaultTimeoutDuration - constant
	DefaultTimeoutDuration = 60 * time.Minute
	// DefaultSessionIDHeaderName - constant
	DefaultSessionIDHeaderName = "immersesid"
	// DefaultSubstituteSessionID - The default substitution string for the client Thrift session ID
	DefaultSubstituteSessionID = "IMMERSE_FAKE_SESSION_ID"
	// DefaultIdleSessionDurationMins - constant
	DefaultIdleSessionDurationMins = 60
	// ThirtyDaysInMins - constant
	ThirtyDaysInMins = 60 * 24 * 30
	// DefaultMaxSessionDurationMins - constant
	DefaultMaxSessionDurationMins = ThirtyDaysInMins
	// DefaultSSLCertFilePath - constant
	DefaultSSLCertFilePath = "sslcert.pem"
	// DefaultSSLPrivateKeyFile - constant
	DefaultSSLPrivateKeyFile = "sslprivate.key"
	// DefaultRedirectPort - constant
	DefaultRedirectPort = 6280
	// DefaultPort - constant
	DefaultPort = 6273
	// DefaultBinaryThriftPort - constant
	DefaultBinaryThriftPort = 6276
	// DefaultThriftPort - constant
	DefaultThriftPort = 6278
	// DefaultEnvPrefix - constant
	DefaultEnvPrefix = "HEAVY"
	// DefaultConfigType - constant
	DefaultConfigType = "toml"
	// DefaultIQURL - constant
	DefaultIQURL = "http://localhost:6275"
	// DBNameRouteParamName - constant
	DBNameRouteParamName = "dbName"
	// RoleRouteParamName - constant
	RoleRouteParamName = "role"
	// TableNameRouteParamName - constant
	TableNameRouteParamName = "tableName"
	// NewTableNameRouteParamName - constant
	NewTableNameRouteParamName = "newTableName"
	// DefaultMinEncryptionKeyLength - constant
	DefaultMinEncryptionKeyLength = 32
	// FsiServerNameSuffix - constant
	FsiServerNameSuffix = "_server"
	// InstanceConfigLoadError - property that is present on instance config when error occurs reading instance config
	InstanceConfigLoadError = "__load_error__"
)

var (
	// AllowedHTTPMethods - Allowed HTTP Methods constant
	AllowedHTTPMethods = []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions, http.MethodPatch}
	// BodyTagRegex - Body tag regex constant
	BodyTagRegex = regexp.MustCompile("<body [^>]*>")
	// SupportedUploadExtensions - Supported upload file extensions
	SupportedUploadExtensions = []string{
		"aih",
		"ain",
		"atx",
		"bzip2",
		"cpg",
		"csv",
		"dbf",
		"fbn",
		"fbx",
		"geojson",
		"gzip",
		"ixs",
		"jp2",
		"json",
		"kml",
		"kmz",
		"mxs",
		"parquet",
		"prj",
		"qix",
		"rar",
		"sbn",
		"sbx",
		"shp",
		"shp.xml",
		"shx",
		"tar",
		"tgz",
		"tif",
		"tiff",
		"tsv",
		"txt",
		"xml",
		"zip",
		"zip-7",
	}
)
