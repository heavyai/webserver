// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/middlewares"
)

func configureInsecureServer(echoInstance *echo.Echo) *http.Server {
	echoInstance.Server.ReadTimeout = config.HTTP.ConnTimeout
	echoInstance.Server.WriteTimeout = config.HTTP.ConnTimeout
	return echoInstance.Server
}

func configureSecureServer(echoInstance *echo.Echo) (server *http.Server, err error) {
	var cert []byte
	if cert, err = os.ReadFile(config.Paths.CertFile); err != nil {
		return
	}

	var key []byte
	if key, err = os.ReadFile(config.Paths.KeyFile); err != nil {
		return
	}

	tlsConfig, err := getTLSConfig(cert, key)
	if err != nil {
		config.Log.Fatal("Could not build TLSConfig: ", err)
		return
	}

	server = echoInstance.TLSServer
	server.ReadTimeout = config.HTTP.ConnTimeout
	server.WriteTimeout = config.HTTP.ConnTimeout
	server.TLSConfig = tlsConfig

	server.Addr = config.HTTP.Addr
	if !echoInstance.DisableHTTP2 {
		server.TLSConfig.NextProtos = append(server.TLSConfig.NextProtos, "h2")
	}

	return
}

func getTLSConfig(cert []byte, key []byte) (tlsConfig *tls.Config, err error) {
	tlsConfig = &tls.Config{
		CipherSuites:     config.HTTP.TLSConfig.CiperSuites,
		MinVersion:       config.HTTP.TLSConfig.MinVersion,
		MaxVersion:       config.HTTP.TLSConfig.MaxVersion,
		CurvePreferences: config.HTTP.TLSConfig.Curves,
		Certificates:     make([]tls.Certificate, 1),
	}
	if tlsConfig.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
		return
	}
	if config.Enable.HTTPSAuth {
		peerCert, err := os.ReadFile(config.Paths.PeerCertFile)
		if err != nil {
			config.Log.Fatal("Error opening peer file:", err, config.Paths.PeerCertFile)
			return tlsConfig, err
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(peerCert)
		tlsConfig.ClientCAs = certPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.BuildNameToCertificate()
	}
	return
}

// configureGracefulShutdown - Wait for interrupt signal to gracefully shutdown the server
func configureGracefulShutdown(e *echo.Echo, timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		config.Log.Fatal("Shutting down: ", err)
	}
}

func configureRedirectListener() {
	redirectListener := echo.New()
	redirectListener.HideBanner = true
	redirectListener.Pre(middlewares.RedirectPort)
	config.Log.Fatal("Shutting down http redirect listener: ", redirectListener.Start(":"+strconv.Itoa(config.HTTP.HTTPSRedirectPort)))
}
