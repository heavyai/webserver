// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package db_thrift_client

import (
	"fmt"
	"net/url"
	"os"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

func New() (*heavy.HeavyClient, error) {
	var protocolFactory thrift.TProtocolFactory
	var hostUrl *url.URL
	if config.Other.EnableBinaryThrift {
		protocolFactory = thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{})
		hostUrl = config.Paths.BinaryBackendURL
	} else {
		protocolFactory = thrift.NewTJSONProtocolFactory()
		hostUrl = config.Paths.BackendURL
	}
	transport, err := thrift.NewTHttpClient(hostUrl.String())
	if err != nil {
		_, err := fmt.Fprintln(os.Stderr, "Error creating transport", err)
		if err != nil {
			config.Log.Error("Couldn't write error to STDERR")
		}
		os.Exit(1)
	}
	if err := transport.Open(); err != nil {
		_, err := fmt.Fprintln(os.Stderr, "Error opening socket to ", hostUrl.String(), " ", err)
		if err != nil {
			config.Log.Error("Couldn't write error to STDERR")
		}
		os.Exit(1)
	}
	inProtocol := protocolFactory.GetProtocol(transport)
	outProtocol := protocolFactory.GetProtocol(transport)
	return heavy.NewHeavyClient(thrift.NewTStandardClient(inProtocol, outProtocol)), nil
}
