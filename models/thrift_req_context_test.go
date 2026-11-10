// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

type thriftWritable interface {
	Write(context.Context, thrift.TProtocol) error
}

func TestThriftReqContextRewritesSessionForBinaryAndJSONProtocols(t *testing.T) {
	for _, testCase := range thriftProtocolTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			original := &heavy.HeavySqlExecuteArgs{
				Session:      "old-session",
				Query:        "SELECT * FROM flights",
				ColumnFormat: true,
				Nonce:        "request-123",
				FirstN:       25,
				AtMostN:      100,
			}
			body := encodeThriftMessage(t, testCase.factory, "sql_execute", 17, original)
			immerseCtx := newTestImmerseContext(body)

			thriftCtx, err := NewThriftReqContext(immerseCtx, testCase.binary)
			if err != nil {
				t.Fatal(err)
			}
			if thriftCtx.GetThriftMethod() != "sql_execute" {
				t.Fatalf("expected sql_execute method, got %q", thriftCtx.GetThriftMethod())
			}
			if thriftCtx.GetSessionID() != "old-session" {
				t.Fatalf("expected original session, got %q", thriftCtx.GetSessionID())
			}
			if thriftCtx.IsBinary() != testCase.binary {
				t.Fatalf("expected binary=%t, got %t", testCase.binary, thriftCtx.IsBinary())
			}

			if err := thriftCtx.SetSessionID("new-session"); err != nil {
				t.Fatal(err)
			}

			rewritten := &heavy.HeavySqlExecuteArgs{}
			decodeThriftRequest(t, testCase.factory, immerseCtx.Request(), "sql_execute", 17, rewritten)

			if rewritten.Session != "new-session" {
				t.Errorf("expected rewritten session, got %q", rewritten.Session)
			}
			if rewritten.Query != original.Query {
				t.Errorf("expected query %q, got %q", original.Query, rewritten.Query)
			}
			if rewritten.ColumnFormat != original.ColumnFormat {
				t.Errorf("expected column format %t, got %t", original.ColumnFormat, rewritten.ColumnFormat)
			}
			if rewritten.Nonce != original.Nonce {
				t.Errorf("expected nonce %q, got %q", original.Nonce, rewritten.Nonce)
			}
			if rewritten.FirstN != original.FirstN {
				t.Errorf("expected first_n %d, got %d", original.FirstN, rewritten.FirstN)
			}
			if rewritten.AtMostN != original.AtMostN {
				t.Errorf("expected at_most_n %d, got %d", original.AtMostN, rewritten.AtMostN)
			}
		})
	}
}

func TestThriftReqContextRewritesConnectCredentialsForBinaryAndJSONProtocols(t *testing.T) {
	for _, testCase := range thriftProtocolTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			original := &heavy.HeavyConnectArgs{
				User:   "immerse",
				Passwd: "original-password",
				Dbname: "heavyai",
			}
			body := encodeThriftMessage(t, testCase.factory, "connect", 23, original)
			immerseCtx := newTestImmerseContext(body)

			thriftCtx, err := NewThriftReqContext(immerseCtx, testCase.binary)
			if err != nil {
				t.Fatal(err)
			}
			if thriftCtx.HasSessionID() {
				t.Fatal("connect must not be parsed as a session-bearing call")
			}

			credentials, err := thriftCtx.GetCredentials()
			if err != nil {
				t.Fatal(err)
			}
			if credentials.User != original.User || credentials.Passwd != original.Passwd || credentials.Dbname != original.Dbname {
				t.Fatalf("unexpected decoded credentials: %#v", credentials)
			}

			credentials.Passwd = "replacement-password"
			if err := thriftCtx.SetCredentials(credentials); err != nil {
				t.Fatal(err)
			}

			rewritten := &heavy.HeavyConnectArgs{}
			decodeThriftRequest(t, testCase.factory, immerseCtx.Request(), "connect", 23, rewritten)
			if rewritten.User != original.User {
				t.Errorf("expected user %q, got %q", original.User, rewritten.User)
			}
			if rewritten.Passwd != "replacement-password" {
				t.Errorf("expected replacement password, got %q", rewritten.Passwd)
			}
			if rewritten.Dbname != original.Dbname {
				t.Errorf("expected database %q, got %q", original.Dbname, rewritten.Dbname)
			}
		})
	}
}

type thriftProtocolTestCase struct {
	name    string
	binary  bool
	factory thrift.TProtocolFactory
}

func thriftProtocolTestCases() []thriftProtocolTestCase {
	return []thriftProtocolTestCase{
		{
			name:    "binary",
			binary:  true,
			factory: thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{}),
		},
		{
			name:    "json",
			binary:  false,
			factory: thrift.NewTJSONProtocolFactory(),
		},
	}
}

func newTestImmerseContext(body []byte) *ImmerseReqContext {
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	response := httptest.NewRecorder()
	return &ImmerseReqContext{Context: echo.New().NewContext(request, response)}
}

func encodeThriftMessage(
	t *testing.T,
	factory thrift.TProtocolFactory,
	method string,
	sequenceID int32,
	value thriftWritable,
) []byte {
	t.Helper()

	buffer := &bytes.Buffer{}
	protocol := factory.GetProtocol(thrift.NewStreamTransportRW(buffer))
	ctx := context.Background()
	if err := protocol.WriteMessageBegin(ctx, method, thrift.CALL, sequenceID); err != nil {
		t.Fatal(err)
	}
	if err := value.Write(ctx, protocol); err != nil {
		t.Fatal(err)
	}
	if err := protocol.WriteMessageEnd(ctx); err != nil {
		t.Fatal(err)
	}
	if err := protocol.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func decodeThriftRequest(
	t *testing.T,
	factory thrift.TProtocolFactory,
	request *http.Request,
	expectedMethod string,
	expectedSequenceID int32,
	value interface {
		Read(context.Context, thrift.TProtocol) error
	},
) {
	t.Helper()

	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	if request.ContentLength != int64(len(body)) {
		t.Fatalf("expected content length %d, got %d", len(body), request.ContentLength)
	}

	protocol := factory.GetProtocol(thrift.NewStreamTransportR(bytes.NewReader(body)))
	method, messageType, sequenceID, err := protocol.ReadMessageBegin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if method != expectedMethod {
		t.Fatalf("expected method %q, got %q", expectedMethod, method)
	}
	if messageType != thrift.CALL {
		t.Fatalf("expected call message type, got %d", messageType)
	}
	if sequenceID != expectedSequenceID {
		t.Fatalf("expected sequence ID %d, got %d", expectedSequenceID, sequenceID)
	}
	if err := value.Read(context.Background(), protocol); err != nil {
		t.Fatal(err)
	}
	if err := protocol.ReadMessageEnd(context.Background()); err != nil {
		t.Fatal(err)
	}
}
