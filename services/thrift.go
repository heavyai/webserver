// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"

	uuid "github.com/nu7hatch/gouuid"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/db_thrift_client"
	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

// SQLExecute - util method
func SQLExecute(sessionID heavy.TSessionId, query string) (result *heavy.TQueryResult_, err error) {
	thriftClient, err := CreateThriftClientWithTimeout()
	if err != nil {
		return
	}
	defer thriftClient.CancelFunction()

	nonce, err := uuid.NewV4()
	if err != nil {
		config.Log.Error("Could not create nonce for SQLExecute query", err)
		return
	}
	result, err = thriftClient.Client.SqlExecute(
		thriftClient.Ctx,
		sessionID,
		query,
		true,
		nonce.String(), -1, -1,
	)
	return
}

// CreateThriftClientWithTimeout - CreateThriftClientWithTimeout method
func CreateThriftClientWithTimeout() (client ThriftClient, err error) {
	thriftClientCtx, cancelFunction := context.WithTimeout(context.Background(), config.HTTP.ConnTimeout)

	thriftClient, err := CreateThriftClient()
	if err != nil {
		cancelFunction()
		return
	}

	return ThriftClient{
		Client:         thriftClient,
		Ctx:            thriftClientCtx,
		CancelFunction: cancelFunction,
	}, nil
}

// CreateThriftClient - creates an Thrift client for DB
func CreateThriftClient() (thriftClient *heavy.HeavyClient, err error) {
	thriftClient, err = db_thrift_client.New()
	if err != nil {
		config.Log.Error("Could not create thrift client", err)
		return
	}
	return thriftClient, nil
}

// ThriftClient - ThriftClient struct
type ThriftClient struct {
	Client         *heavy.HeavyClient
	Ctx            context.Context
	CancelFunction func()
}
