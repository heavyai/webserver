// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/andrewseidl/viper"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/heavyai/webserver/internal/constants"
)

var (
	Enable struct {
		AllowAnyOrigin          bool
		IQ                      bool
		BrowserLogs             bool
		Compress                bool
		CrossDomainAuth         bool
		Development             bool
		EncryptedCredentials    bool
		LocalDataCatalog        bool
		Jupyter                 bool
		JupyterDesktop          bool
		HTTPS                   bool
		HTTPSAuth               bool
		HTTPSRedirect           bool
		LegacyAuth              bool
		Geocoder                bool
		GoogleMetrics           bool
		CustomIntegrationAuth   bool
		QueryInterrupt          bool
		ReverseProxies          bool
		ReadOnly                bool
		SAML                    bool
		Transcription           bool
		Verbose                 bool
		StripXHeadersMiddleware bool
		UltraSecureMode         bool
	}
	HTTP struct {
		Addr              string
		HTTPSRedirectPort int
		Port              int
		ConnTimeout       time.Duration
		TLSConfig         tlsConfig
		SecureACAOUri     string
	}
	Jupyter struct {
		JupyterPrefix string
		JupyterURL    *url.URL
		DesktopURL    *url.URL
	}
	Paths struct {
		IQServiceURL      *url.URL
		BackendURL        *url.URL
		BinaryBackendURL  *url.URL
		CertFile          string
		DataCatalogDir    string
		DataDir           string
		DocsDir           string
		Frontend          string
		KeyFile           string
		JWTKeyFile        string
		PeerCertFile      string
		ServersJSON       string
		AccessLogFile     string
		AllLogFile        string
		ErrorDBLogFile    string
		InfoDBLogFile     string
		WarningDBLogFile  string
		AccessIQLogFile   string
		AllIQLogFile      string
		BuildIQLogFile    string
		ChromaIQLogFile   string
		ConsoleIQLogFile  string
		GuidanceIQLogFile string
		TranscriptionURL  *url.URL
	}
	Session struct {
		SessionStore    *sessions.CookieStore
		SessionIDHeader string
		PrivateKey      *rsa.PrivateKey
		IdleTimeout     time.Duration
		MaxTimeout      time.Duration
		Durations       sessionTimeoutDurations
	}
	Other struct {
		AccessLog                      io.Writer
		EncryptedCredentialsKey        string
		DBConfig                       map[string][]byte
		Commit                         string
		Locations                      []Location
		Proxies                        []ReverseProxy
		ServersJSONParams              []string
		SAMLurl                        *url.URL
		ReturnURLFlag                  bool
		StripXHeaders                  []string
		CookieHeavyDBAuth              string
		SubstituteSessionID            string
		IndexHTMLbytes                 []byte
		GoogleTagID                    string
		EnableBinaryThrift             bool
		EnableUploadExtensionCheck     bool
		AdditionalFileUploadExtensions []string
		InstanceConfig                 *gabs.Container
	}
	Log *logrus.Logger
)

var (
	config        string
	sslCert       string
	sslPrivateKey string
	tmpDir        string
	commit        string
	serversJSON   *gabs.Container
)

var defaultStripXHeaders = []string{
	constants.HeavyDBUsernameXHeaderName,
}

// Location - Location model
type Location struct {
	Location string  `json:"location"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

// ReverseProxy - Reverse proxy path and target
type ReverseProxy struct {
	Path   string
	Target *url.URL
}

type tlsConfig struct {
	CiperSuites []uint16
	MinVersion  uint16
	MaxVersion  uint16
	Curves      []tls.CurveID
}

type sessionTimeoutDurations struct {
	IdleSessionDuration int64 `json:"idleSessionDuration" xml:"idleSessionDuration"`
	MaxSessionDuration  int64 `json:"maxSessionDuration" xml:"maxSessionDuration"`
}

func getLogPath(ext string) string {
	return filepath.Join(Paths.DataDir, constants.RelativeLogPath, ext)
}

func getLogName(lvl string) string {
	n := filepath.Base(os.Args[0])
	h, _ := os.Hostname()
	us, _ := user.Current()
	u := us.Username
	t := time.Now().Format("20060102-150405")
	p := strconv.Itoa(os.Getpid())

	return n + "." + h + "." + u + ".log." + lvl + "." + t + "." + p
}

func createLogSymLink(lvl string) {
	if _, err := os.Lstat(getLogPath("heavy_web_server." + lvl)); err == nil {
		os.Remove(getLogPath("heavy_web_server." + lvl))
	}
	err := os.Symlink(getLogName(lvl), getLogPath("heavy_web_server."+lvl))
	if err != nil {
		Log.Error("Error creating "+lvl+" log file SymLink: ", err)
	}
}

func getKeyFromFile(path string) (key *rsa.PrivateKey, err error) {
	pemstr, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode([]byte(pemstr))

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes); err != nil {
			return nil, err
		}
	}

	var ok bool
	key, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("Unable to cast key to RSA key")
	}

	return key, err
}

func getEncryptedCredentialsKey(encryptionKeyFilePath string) (key string, err error) {
	if len(encryptionKeyFilePath) > 0 {
		fileContent, err := os.ReadFile(encryptionKeyFilePath)
		if err != nil {
			return "", err
		}
		keyString := strings.TrimSpace(string(fileContent))
		bits := len(keyString) * 8
		if bits != 256 {
			err = fmt.Errorf(
				"encryption key must be 256 bits in length. Retrieved %d bits from file",
				bits,
			)
			return "", err
		}
		return keyString, nil
	}
	return
}

func parseAdditionalFileUploadExtensions(optionValue string) []string {
	bits := strings.Split(
		viper.GetString("web.additional-file-upload-extensions"),
		",",
	)
	for i := range bits {
		bits[i] = strings.Trim(bits[i], ". ")
	}
	return bits
}

// LoadCSVData loads the CSV data from a file into memory.
func LoadCSVData(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip the header row
	if _, err := reader.Read(); err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		lat, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return err
		}

		lng, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return err
		}

		location := Location{
			Location: record[0],
			Lat:      lat,
			Lng:      lng,
		}
		Other.Locations = append(Other.Locations, location)
	}

	return nil
}

func init() {
	var err error
	Log = logrus.New()
	Log.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	// Flags - HTTP
	pflag.Bool("enable-https", false, "Enable HTTPS support")
	pflag.Bool("enable-https-authentication", false, "Enable PKI authentication")
	pflag.Bool("enable-https-redirect", false, "Enable HTTP to HTTPS redirect")
	pflag.Bool("enable-cross-domain", false, "Enable frontend cross-domain authentication")
	pflag.Bool("enable-cert-verification", true, "Enable TLS cert verification")
	pflag.Int("http-to-https-redirect-port", constants.DefaultRedirectPort, "Frontend server port for HTTP redirect, when HTTPS enabled")
	pflag.IntP("port", "p", constants.DefaultPort, "Frontend server port")
	pflag.String("secure-acao-uri", "", "URI to send as Access-Control-Allow-Origin header value [Note: Will default to '*' if --development enabled]")

	// Flags - Internal
	pflag.BoolP("quiet", "q", true, "Suppress non-error messages")
	pflag.BoolP("verbose", "v", false, "Print all log messages to stdout")

	// Flags - Jupyter
	pflag.String("jupyter-url", "", "Url for Jupyter integration")
	pflag.String("jupyter-prefix", constants.DefaultJupyterPrefix, "Jupyter Hub base_url for Jupyter integration")
	pflag.String("jupyter-desktop-url", "", "URL for Jupyter Desktop API")

	// Flags - Paths
	pflag.String("iq-url", constants.DefaultIQURL, "Url to HeavyIQ service")
	pflag.StringP("backend-url", "b", "", "Url to http-port on heavydb [http://localhost:6278]")
	pflag.StringP("binary-backend-url", "B", "", "Url to http-binary-port on heavydb [http://localhost:6276]")
	pflag.String("cert", constants.DefaultCertFilePath, "Certificate file for HTTPS")
	pflag.StringP("data", "d", constants.DefaultDataPath, "Path to HeavyDB data directory")
	pflag.String("docs", constants.DefaultDocsPath, "Path to documentation directory")
	pflag.StringP("frontend", "f", constants.DefaultFrontendPath, "Path to frontend directory")
	pflag.String("geocoder-csv", "", "Path to Geocoder CSV")
	pflag.String("data-catalog", "", "Path to data catalog directory")
	pflag.String("key", constants.DefaultHTTPSKeyFilePath, "Key file for HTTPS")
	pflag.String("peer-cert", constants.DefaultPeerCertPath, "Peer CA certificate PKI authentication")
	pflag.String("servers-json", "", "Path to servers.json")
	pflag.StringP("jwt-key-file", "j", "", "Path to .pem file containing RSA private key for JWT encryption")
	pflag.String("transcription-url", "", "Url to transcription service")

	// Flags - Sessions
	pflag.Duration("timeout", constants.DefaultTimeoutDuration, "Maximum request duration")
	pflag.String("session-id-header", constants.DefaultSessionIDHeaderName, "Session ID header")
	pflag.Int("idle-session-duration", constants.DefaultIdleSessionDurationMins, "Idle session timeout in minutes")
	pflag.Int("max-session-duration", constants.DefaultMaxSessionDurationMins, "Maximum session timeout in minutes")

	// Flags - Other
	pflag.Bool("ultra-secure-mode", false, "Enable Ultra Secure Mode mode")
	pflag.Bool("allow-any-origin", false, "Allow any origin in ACAO HTTP header")
	pflag.Bool("custom-integration-auth", false, "Allow alternative authentication modes for a custom integration")
	pflag.Bool("development", false, "Enable development mode")
	pflag.StringSlice("reverse-proxy", nil, "Additional endpoints to act as reverse proxies, format '/endpoint/:http://target.example.com'")
	pflag.Bool("enable-legacy-auth", true, "Enable legacy client authentication")
	pflag.Bool("enable-browser-logs", true, "Enable browser logs")
	pflag.Bool("compress", true, "Enable gzip compression")
	pflag.BoolP("read-only", "r", false, "Enable read-only mode")
	pflag.StringSlice("strip-x-headers", defaultStripXHeaders, "List of custom 'X' http request headers to be removed from incoming requests. Use --strip-x-headers=\"\" to allow all X headers through.")
	pflag.String("auth-cookie-name", constants.DefaultAuth, "HeavyDB auth cookie name (HeavyDB Auth Token)")
	pflag.String("encryption-key-file-path", "", "Path to file containing credential payload cipher key. Key must be 256bits in length.")
	pflag.String("substitute-session-id", constants.DefaultSubstituteSessionID, "Subsitute Session ID")
	pflag.Bool("enable-runtime-query-interrupt", false, "Enable runtime query interrupt")
	pflag.Bool("enable-non-kernel-time-query-interrupt", true, "Enable non-kernel-time query interrupt")
	pflag.String("google-tag-id", "", "Google Analytics Tag ID")
	pflag.Bool("enable-binary-thrift", true, "Use the binary thrift protocol")
	pflag.Bool("enable-upload-extension-check", false, "Disable file extension check for file uploads")
	pflag.String("additional-file-upload-extensions", "", "Denote additional file extensions allowable for uploads")

	// Flags - init
	pflag.String("tmpdir", "", "Path for temporary file storage [/tmp]")
	pflag.StringP("config", "c", "", "Path to HeavyDB configuration file")
	pflag.String("ssl-cert", constants.DefaultSSLCertFilePath, "SSL validated public certificate")
	pflag.String("ssl-private-key", constants.DefaultSSLPrivateKeyFile, "SSL private key file")
	pflag.Bool("help", false, "Prints to standard output the default values of all defined command-line flags")
	pflag.Bool("version", false, "Print the current daemon version")

	// Flags - Cyphers
	pflag.String("min-tls-version", defaultMinTLSVersion, "Minimum supported TLS version supported by cipher_suites lib: https://golang.org/src/crypto/tls/cipher_suites.go")
	pflag.String("max-tls-version", "", "Maximum supported TLS version supported by cipher_suites lib: https://golang.org/src/crypto/tls/cipher_suites.go")
	pflag.StringSlice("tls-cipher-suites", defaultCipherKeys, "List of supported by cipher_suites go lib: https://golang.org/src/crypto/tls/cipher_suites.go")
	pflag.StringSlice("tls-curves", defaultTLSCurveKeys, "List of supported by cipher_suites go lib: https://golang.org/src/crypto/tls/cipher_suites.go")

	// Flags - Hidden from --help
	pflag.CommandLine.MarkHidden("help")
	pflag.CommandLine.MarkHidden("compress")
	pflag.CommandLine.MarkHidden("geocoder-csv")
	pflag.CommandLine.MarkHidden("quiet")
	pflag.CommandLine.MarkHidden("reverse-proxy")
	pflag.CommandLine.MarkHidden("enable-legacy-auth")
	pflag.CommandLine.MarkHidden("jupyter-desktop-url")
	pflag.CommandLine.MarkHidden("google-tag-id")
	pflag.CommandLine.MarkHidden("custom-integration-auth")
	pflag.CommandLine.MarkHidden("development")
	pflag.CommandLine.MarkHidden("auth-cookie-name")
	pflag.CommandLine.MarkHidden("encrypted-credentials-key")
	pflag.CommandLine.MarkHidden("substitute-session-id")

	pflag.CommandLine.SetOutput(os.Stdout)
	pflag.ErrHelp = nil

	pflag.Parse()

	if h, _ := pflag.CommandLine.GetBool("help"); h {
		fmt.Fprintf(os.Stdout, "Usage of %s:\n", os.Args[0])
		pflag.CommandLine.PrintDefaults()
		os.Exit(0)
	}

	if h, _ := pflag.CommandLine.GetBool("version"); h {
		fmt.Fprintf(os.Stdout, "%s\n", commit)
		os.Exit(0)
	}

	// Viper Bind - HTTP
	viper.BindPFlag("web.enable-https", pflag.CommandLine.Lookup("enable-https"))
	viper.BindPFlag("web.enable-https-authentication", pflag.CommandLine.Lookup("enable-https-authentication"))
	viper.BindPFlag("web.enable-cross-domain", pflag.CommandLine.Lookup("enable-cross-domain"))
	viper.BindPFlag("web.enable-https-redirect", pflag.CommandLine.Lookup("enable-https-redirect"))
	viper.BindPFlag("web.enable-cert-verification", pflag.CommandLine.Lookup("enable-cert-verification"))
	viper.BindPFlag("web.http-to-https-redirect-port", pflag.CommandLine.Lookup("http-to-https-redirect-port"))
	viper.BindPFlag("web.port", pflag.CommandLine.Lookup("port"))
	viper.BindPFlag("web.min-tls-version", pflag.CommandLine.Lookup("min-tls-version"))
	viper.BindPFlag("web.max-tls-version", pflag.CommandLine.Lookup("max-tls-version"))
	viper.BindPFlag("web.tls-cipher-suites", pflag.CommandLine.Lookup("tls-cipher-suites"))
	viper.BindPFlag("web.tls-curves", pflag.CommandLine.Lookup("tls-curves"))

	// Viber Bind - Internal
	viper.BindPFlag("quiet", pflag.CommandLine.Lookup("quiet"))
	viper.BindPFlag("verbose", pflag.CommandLine.Lookup("verbose"))

	// Viper Bind - Jupyter
	viper.BindPFlag("web.jupyter-url", pflag.CommandLine.Lookup("jupyter-url"))
	viper.BindPFlag("web.jupyter-prefix", pflag.CommandLine.Lookup("jupyter-prefix"))
	viper.BindPFlag("web.jupyter-desktop-url", pflag.CommandLine.Lookup("jupyter-desktop-url"))

	// Viper Bind - Paths
	viper.BindPFlag("web.iq-url", pflag.CommandLine.Lookup("iq-url"))
	viper.BindPFlag("web.backend-url", pflag.CommandLine.Lookup("backend-url"))
	viper.BindPFlag("web.binary-backend-url", pflag.CommandLine.Lookup("binary-backend-url"))
	viper.BindPFlag("web.cert", pflag.CommandLine.Lookup("cert"))
	viper.BindPFlag("data", pflag.CommandLine.Lookup("data"))
	viper.BindPFlag("web.docs", pflag.CommandLine.Lookup("docs"))
	viper.BindPFlag("web.data-catalog", pflag.CommandLine.Lookup("data-catalog"))
	viper.BindPFlag("web.frontend", pflag.CommandLine.Lookup("frontend"))
	viper.BindPFlag("web.geocoder-csv", pflag.CommandLine.Lookup("geocoder-csv"))
	viper.BindPFlag("web.key", pflag.CommandLine.Lookup("key"))
	viper.BindPFlag("web.peer-cert", pflag.CommandLine.Lookup("peer-cert"))
	viper.BindPFlag("web.servers-json", pflag.CommandLine.Lookup("servers-json"))
	viper.BindPFlag("web.jwt-key-file", pflag.CommandLine.Lookup("jwt-key-file"))
	viper.BindPFlag("web.transcription-url", pflag.CommandLine.Lookup("transcription-url"))

	// Viper Bind - Sessions
	viper.BindPFlag("web.timeout", pflag.CommandLine.Lookup("timeout"))
	viper.BindPFlag("web.session-id-header", pflag.CommandLine.Lookup("session-id-header"))
	viper.BindPFlag("idle-session-duration", pflag.CommandLine.Lookup("idle-session-duration"))
	viper.BindPFlag("max-session-duration", pflag.CommandLine.Lookup("max-session-duration"))

	// Viper Bind - Other
	viper.BindPFlag("read-only", pflag.CommandLine.Lookup("read-only"))
	viper.BindPFlag("web.allow-any-origin", pflag.CommandLine.Lookup("allow-any-origin"))
	viper.BindPFlag("web.development", pflag.CommandLine.Lookup("development"))
	viper.BindPFlag("web.reverse-proxy", pflag.CommandLine.Lookup("reverse-proxy"))
	viper.BindPFlag("web.compress", pflag.CommandLine.Lookup("compress"))
	viper.BindPFlag("web.enable-browser-logs", pflag.CommandLine.Lookup("enable-browser-logs"))
	viper.BindPFlag("web.enable-legacy-auth", pflag.CommandLine.Lookup("enable-legacy-auth"))
	viper.BindPFlag("web.custom-integration-auth", pflag.CommandLine.Lookup("custom-integration-auth"))
	viper.BindPFlag("web.strip-x-headers", pflag.CommandLine.Lookup("strip-x-headers"))
	viper.BindPFlag("web.auth-cookie-name", pflag.CommandLine.Lookup("auth-cookie-name"))
	viper.BindPFlag("web.encryption-key-file-path", pflag.CommandLine.Lookup("encryption-key-file-path"))
	viper.BindPFlag("web.substitute-session-id", pflag.CommandLine.Lookup("substitute-session-id"))
	viper.BindPFlag("web.ultra-secure-mode", pflag.CommandLine.Lookup("ultra-secure-mode"))
	viper.BindPFlag("web.secure-acao-uri", pflag.CommandLine.Lookup("secure-acao-uri"))
	viper.BindPFlag("enable-runtime-query-interrupt", pflag.CommandLine.Lookup("enable-runtime-query-interrupt"))
	viper.BindPFlag("enable-non-kernel-time-query-interrupt", pflag.CommandLine.Lookup("enable-non-kernel-time-query-interrupt"))
	viper.BindPFlag("web.google-tag-id", pflag.CommandLine.Lookup("google-tag-id"))
	viper.BindPFlag("web.enable-binary-thrift", pflag.CommandLine.Lookup("enable-binary-thrift"))
	viper.BindPFlag("web.enable-upload-extension-check", pflag.CommandLine.Lookup("enable-upload-extension-check"))
	viper.BindPFlag("web.additional-file-upload-extensions", pflag.CommandLine.Lookup("additional-file-upload-extensions"))

	// Viper Bind - init
	viper.BindPFlag("web.tmpdir", pflag.CommandLine.Lookup("tmpdir"))
	viper.BindPFlag("config", pflag.CommandLine.Lookup("config"))
	viper.BindPFlag("ssl-cert", pflag.CommandLine.Lookup("ssl-cert"))
	viper.BindPFlag("ssl-private-key", pflag.CommandLine.Lookup("ssl-private-key"))

	// Viper Settings
	r := strings.NewReplacer(".", "_")

	viper.SetDefault("http-binary-port", constants.DefaultBinaryThriftPort)
	viper.SetDefault("http-port", constants.DefaultThriftPort)
	viper.SetEnvPrefix(constants.DefaultEnvPrefix)
	viper.SetEnvKeyReplacer(r)
	viper.AutomaticEnv()

	if viper.IsSet("config") {
		viper.SetConfigType(constants.DefaultConfigType)
		viper.SetConfigFile(viper.GetString("config"))

		if err := viper.ReadInConfig(); err != nil {
			Log.Fatal("Error reading config file: " + err.Error())
		}
	}

	// Viper Get - Config
	Enable.AllowAnyOrigin = viper.GetBool("web.allow-any-origin") || viper.GetBool("web.development")
	Enable.BrowserLogs = viper.GetBool("web.enable-browser-logs") || viper.GetBool("web.development")
	Enable.HTTPS = viper.GetBool("web.enable-https")
	Enable.HTTPSAuth = viper.GetBool("web.enable-https-authentication")
	Enable.HTTPSRedirect = viper.GetBool("web.enable-https-redirect")
	Enable.CrossDomainAuth = viper.GetBool("web.enable-cross-domain") && viper.GetBool("web.enable-https")
	Enable.LegacyAuth = viper.GetBool("web.enable-legacy-auth")
	Enable.CustomIntegrationAuth = viper.GetBool("web.custom-integration-auth")
	Enable.QueryInterrupt = viper.GetBool("enable-runtime-query-interrupt") || viper.GetBool("enable-non-kernel-time-query-interrupt")
	Other.EnableBinaryThrift = viper.GetBool("web.enable-binary-thrift")
	Other.EnableUploadExtensionCheck = viper.GetBool("web.enable-upload-extension-check")
	Other.AdditionalFileUploadExtensions = parseAdditionalFileUploadExtensions(
		viper.GetString("web.additional-file-upload-extensions"),
	)
	Other.GoogleTagID = viper.GetString("web.google-tag-id")
	Enable.GoogleMetrics = len(Other.GoogleTagID) != 0

	iqURLStr := viper.GetString("web.iq-url")
	Paths.IQServiceURL, err = url.Parse(iqURLStr)
	if err != nil {
		Log.Fatal("Could not parse IQ URL: ", err)
	}
	Enable.IQ = len(iqURLStr) != 0

	transcriptionURLStr := viper.GetString("web.transcription-url")
	Paths.TranscriptionURL, err = url.Parse(transcriptionURLStr)
	if err != nil {
		Log.Fatal("Could not parse transcription URL: ", err)
	}

	Enable.Transcription = len(transcriptionURLStr) != 0

	HTTP.HTTPSRedirectPort = viper.GetInt("web.http-to-https-redirect-port")
	HTTP.Port = viper.GetInt("web.port")
	HTTP.Addr = ":" + strconv.Itoa(HTTP.Port)
	HTTP.ConnTimeout = viper.GetDuration("web.timeout")

	tlsCipherSuites := getTLSCipherSuites()
	tlsCurves := getTLSCurves()
	minTLSVer := tlsVersions[viper.GetString("web.min-tls-version")]
	maxTLSVer := tlsVersions[viper.GetString("web.max-tls-version")]

	HTTP.TLSConfig = tlsConfig{
		CiperSuites: tlsCipherSuites,
		MinVersion:  minTLSVer,
		MaxVersion:  maxTLSVer,
		Curves:      tlsCurves,
	}

	Other.StripXHeaders = viper.GetStringSlice("web.strip-x-headers")
	Other.CookieHeavyDBAuth = viper.GetString("web.auth-cookie-name")
	Other.SubstituteSessionID = viper.GetString("web.substitute-session-id")
	Other.EncryptedCredentialsKey, err = getEncryptedCredentialsKey(
		viper.GetString("web.encryption-key-file-path"),
	)
	if err != nil {
		Log.Fatal("Error setting encrypted credentials key: ", err)
	}
	Enable.EncryptedCredentials = len(Other.EncryptedCredentialsKey) > 0

	Jupyter.JupyterPrefix = viper.GetString("web.jupyter-prefix")

	Paths.CertFile = viper.GetString("web.cert")
	Paths.DataDir = viper.GetString("data")
	Paths.DocsDir = viper.GetString("web.docs")
	Paths.DataCatalogDir = viper.GetString("web.data-catalog")
	Paths.Frontend = viper.GetString("web.frontend")
	Paths.KeyFile = viper.GetString("web.key")
	Paths.PeerCertFile = viper.GetString("web.peer-cert")
	Paths.ServersJSON = viper.GetString("web.servers-json")

	Session.SessionIDHeader = viper.GetString("web.session-id-header")
	idleTimeoutDuration := viper.GetInt("idle-session-duration")
	Session.IdleTimeout = time.Duration(idleTimeoutDuration) * time.Minute
	maxTimeoutDuration := viper.GetInt("max-session-duration")
	Session.MaxTimeout = time.Duration(maxTimeoutDuration) * time.Minute

	Enable.Compress = viper.GetBool("web.compress")
	Enable.ReadOnly = viper.GetBool("read-only")

	// Logging Config
	if viper.IsSet("quiet") && !viper.IsSet("verbose") {
		Log.Println("Option --quiet is deprecated and has been replaced by --verbose=false, which is enabled by default.")
		Enable.Verbose = !viper.GetBool("quiet")
	} else {
		Enable.Verbose = viper.GetBool("verbose")
	}

	// Logging Config - Files
	if _, err := os.Stat(getLogPath("")); os.IsNotExist(err) {
		os.MkdirAll(getLogPath(""), 0755)
	}

	Paths.AllLogFile = getLogPath(getLogName("ALL"))
	Paths.AccessLogFile = getLogPath(getLogName("ACCESS"))
	Paths.ErrorDBLogFile = getLogPath("heavydb.ERROR")
	Paths.InfoDBLogFile = getLogPath("heavydb.INFO")
	Paths.WarningDBLogFile = getLogPath("heavydb.WARNING")
	Paths.AccessIQLogFile = getLogPath("heavyiq.ACCESS")
	Paths.AllIQLogFile = getLogPath("heavyiq.ALL")
	Paths.BuildIQLogFile = getLogPath("heavyiq_build.log")
	Paths.ConsoleIQLogFile = getLogPath("heavyiq_gunicorn.log")
	Paths.ChromaIQLogFile = getLogPath("heavyiq_chroma.log")
	Paths.GuidanceIQLogFile = getLogPath("heavyiq.RAG")

	allLogFile, err := os.OpenFile(Paths.AllLogFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		Log.Fatal("Error opening ALL log file: ", err)
	}
	createLogSymLink("ALL")

	accessLogFile, err := os.OpenFile(Paths.AccessLogFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		Log.Fatal("Error opening ACCESS log file: ", err)
	}
	createLogSymLink("ACCESS")

	// Logging Config - Logger
	if !Enable.Verbose {
		Log.SetOutput(allLogFile)
		Other.AccessLog = accessLogFile
	} else {
		Log.SetOutput(io.MultiWriter(os.Stdout, allLogFile))
		Other.AccessLog = io.MultiWriter(os.Stdout, accessLogFile)
	}

	// Logging Config - Close the log files on process exit
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		allLogFile.Close()
		accessLogFile.Close()
	}()

	// Immerse DB Configuration Persistence Config
	Other.DBConfig = make(map[string][]byte)
	fp := filepath.Join(Paths.DataDir, constants.DBConfigJSONFileName)
	_, err = os.Stat(fp)
	if !os.IsNotExist(err) {
		DBConfigJSON, err := gabs.ParseJSONFile(fp)
		if err != nil {
			Log.Error("Error reading Immerse DB configuration persistence, only temporal persistence available")
		} else {
			dbConfigList := DBConfigJSON.Children()
			for _, c := range dbConfigList {
				dbc := c.ChildrenMap()
				for k, v := range dbc {
					Other.DBConfig[k] = []byte(v.String())
				}
			}
		}
	}

	// Immerse Instance Configuration Persistence Config
	fp = filepath.Join(Paths.DataDir, constants.InstanceConfigJSONFileName)
	_, err = os.Stat(fp)
	if !os.IsNotExist(err) {
		InstanceConfigJSON, err := gabs.ParseJSONFile(fp)
		if err != nil {
			Log.Error("Error reading Immerse instance configuration persistence, only temporal persistence available")
			Other.InstanceConfig = gabs.New()
			Other.InstanceConfig.Set(true, constants.InstanceConfigLoadError)

		} else {
			Other.InstanceConfig = InstanceConfigJSON
		}
	}

	// Viper Get Conditional Config
	backendURLStr := viper.GetString("web.backend-url")
	if backendURLStr == "" {
		s := "http"
		if viper.IsSet("ssl-cert") && viper.IsSet("ssl-private-key") {
			s = "https"
		}
		backendURLStr = s + "://localhost:" + strconv.Itoa(viper.GetInt("http-port"))
	}

	Paths.BackendURL, err = url.Parse(backendURLStr)
	if err != nil {
		Log.Fatal("Could not parse Backend URL: ", err)
	}

	// copy the backend url to binary backend url
	Paths.BinaryBackendURL, _ = url.Parse(backendURLStr)
	binaryBackendURLStr := viper.GetString("web.binary-backend-url")
	if binaryBackendURLStr == "" {
		// change the port
		port := strconv.Itoa(viper.GetInt("http-binary-port"))
		Paths.BinaryBackendURL.Host = Paths.BackendURL.Hostname() + ":" + port
	} else {
		Paths.BinaryBackendURL, err = url.Parse(binaryBackendURLStr)
		if err != nil {
			Log.Fatal("Could not parse Binary Backend URL: ", err)
		}
	}

	if Paths.BackendURL.Scheme == "https" {
		isv := !viper.GetBool("web.enable-cert-verification")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: isv}
	}

	Paths.JWTKeyFile = viper.GetString("web.jwt-key-file")

	// Geocoder CSV
	geocoderCSVPath := viper.GetString("web.geocoder-csv")
	if geocoderCSVPath != "" {
		Enable.Geocoder = true
		// Load the CSV data into memory.
		if err := LoadCSVData("locations.csv"); err != nil {
			Log.Fatalf("Failed to load Gecoder CSV data: %v", err)
		}
	}

	jupyterURLStr := viper.GetString("web.jupyter-url")
	if jupyterURLStr != "" {
		Jupyter.JupyterURL, err = url.Parse(jupyterURLStr)
		if err != nil {
			Log.Fatal("Could not parse Jupyter URL: ", err)
		}
	}

	jupyterDesktopURLStr := viper.GetString("web.jupyter-desktop-url")
	if jupyterDesktopURLStr != "" {
		Jupyter.DesktopURL, err = url.Parse(jupyterDesktopURLStr)
		if err != nil {
			Log.Fatal("Could not parse Jupyter Desktop URL: ", err)
		}
		Enable.JupyterDesktop = true
	}

	if !viper.IsSet("web.servers-json") {
		Paths.ServersJSON = filepath.Join(Paths.Frontend, "servers.json")
	}

	serversJSON, err = gabs.ParseJSONFile(Paths.ServersJSON)
	if err != nil {
		Log.Println("Using system servers.json; could not parse servers.json file at " + Paths.ServersJSON)
	} else {
		Log.Println("Using servers.json file at " + Paths.ServersJSON)

		SAMLURLString, ok := serversJSON.Index(0).Search("SAMLurl").Data().(string)
		if ok {
			Other.SAMLurl, err = url.Parse(SAMLURLString)
			if err != nil {
				Log.Fatal("Could not parse SAML URL: ", err)
			}
		}

		returnURLFlag, ok := serversJSON.Index(0).Search("returnUrl").Data().(bool)
		if ok {
			Other.ReturnURLFlag = returnURLFlag
		}
	}

	for _, rp := range viper.GetStringSlice("web.reverse-proxy") {
		s := strings.SplitN(rp, ":", 2)
		if len(s) != 2 {
			Log.Fatalln("Could not parse reverse proxy string:", rp)
		}
		path := s[0]
		if len(path) == 0 {
			Log.Fatalln("Zero-length path passed for reverse proxy:", rp)
		}

		path = strings.TrimRight(path, "/")
		path = "/" + strings.TrimLeft(path, "/")

		target, err := url.Parse(s[1])
		if err != nil {
			Log.Fatal(err)
		}
		if target.Scheme == "" {
			Log.Fatalln("Missing URL scheme, need full URL including http/https:", target)
		}
		Other.Proxies = append(Other.Proxies, ReverseProxy{path, target})
	}

	if os.Getenv("TMPDIR") != "" {
		tmpDir = os.Getenv("TMPDIR")
	}
	if viper.IsSet("tmpdir") {
		tmpDir = viper.GetString("web.tmpdir")
	}
	if tmpDir != "" {
		err = os.MkdirAll(tmpDir, 0750)
		if err != nil {
			Log.Fatal("Could not create temp dir: ", err)
		}
		os.Setenv("TMPDIR", tmpDir)
	}

	// Non-Viper Config Set
	Other.Commit = commit
	Other.ServersJSONParams = []string{"username", "password", "database", "loadDashboard"}

	if Enable.CustomIntegrationAuth {
		c := 64
		b := make([]byte, c)
		_, err = rand.Read(b)
		if err != nil {
			Log.Fatal("Could not generate random seed for query param cookie encryption: ", err)
		}
		Session.SessionStore = sessions.NewCookieStore(b)
		Session.SessionStore.MaxAge(0)
	}

	Session.Durations = sessionTimeoutDurations{
		IdleSessionDuration: int64(Session.IdleTimeout.Seconds()) * 1000,
		MaxSessionDuration:  int64(Session.MaxTimeout.Seconds()) * 1000,
	}

	if len(Paths.JWTKeyFile) > 0 {
		Session.PrivateKey, err = getKeyFromFile(Paths.JWTKeyFile)
	} else {
		Session.PrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	}
	if err != nil {
		Log.Fatal("Could not parse keys for encryption: ", err)
	}

	Other.IndexHTMLbytes, err = os.ReadFile(filepath.Join(Paths.Frontend, "index.html"))
	if err != nil {
		Log.Error("Could not read index.html. Ensure --frontend option is correctly configured: ", err)
	}

	// Server Mode Config
	Enable.Development = viper.GetBool("web.development")
	Enable.LocalDataCatalog = len(Paths.DataCatalogDir) > 0
	Enable.SAML = Other.SAMLurl != nil && len(Other.SAMLurl.String()) > 0
	Enable.Jupyter = Jupyter.JupyterURL != nil
	Enable.ReverseProxies = len(Other.Proxies) > 0
	Enable.StripXHeadersMiddleware = len(Other.StripXHeaders) > 0

	// Server Settings
	if viper.IsSet("config") {
		Log.Println("Using config file at " + viper.ConfigFileUsed())
	} else {
		Log.Println("Using command line options and system default settings; no config file path set via --config")
	}

	if Enable.Development {
		Log.Printf("Server configuration options: %+v", viper.AllSettings())
	}

	Enable.UltraSecureMode = viper.GetBool("web.ultra-secure-mode")
	HTTP.SecureACAOUri = viper.GetString("web.secure-acao-uri")

	if HTTP.SecureACAOUri != "" && Enable.Development {
		Log.Warn("--secure-acao-uri and --development flag enabled. ACAO header will default to '*' w/ --development enabled")
	}

	if Enable.UltraSecureMode && HTTP.SecureACAOUri == "" {
		Log.Fatal("--secure-acao-uri value required in order to enable --ultra-secure-mode")
	}
}
