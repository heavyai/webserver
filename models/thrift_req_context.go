// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/heavyai/webserver/internal/db_thrift_client/heavy"
)

// ThriftReqContext is used for proxying thrift requests
type ThriftReqContext struct {
	*ImmerseReqContext
	binary        bool
	requestBody   []byte
	protoFactory  thrift.TProtocolFactory
	thriftMethod  string
	messageTypeID thrift.TMessageType
	seqID         int32
	sessionID     string
}

// dumpRequestBody - Dump Request Body util
func dumpRequestBody(req *http.Request) (requestBody []byte, err error) {
	if req.Body != nil { // Read
		requestBody, err = io.ReadAll(req.Body)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(requestBody)) // Reset

	return requestBody, err
}

// thriftMethodHasSession - thrift methods with session id
func thriftMethodHasSession(method string) bool {
	switch method {
	case
		"connect",
		"get_version",
		"krb5_connect",
		"execute_query_step",
		"broadcast_serialized_rows",
		"execute_next_render_step":
		return false
	}
	return true
}

// thriftWhitelistedMethod - thrift methods that are whitelisted
func thriftWhitelistedMethod(method string) bool {
	switch method {
	case
		"broadcast_serialized_rows",
		"connect",
		"krb5_connect",
		"execute_query_step",
		"execute_next_render_step",
		"get_device_parameters",
		"get_version":
		return true
	}
	return false
}

// NewThriftReqContext will build a new context from the existing context
func NewThriftReqContext(immerseCtx *ImmerseReqContext, binary bool) (ctx *ThriftReqContext, err error) {
	var protoFactory thrift.TProtocolFactory
	var requestBody []byte
	requestBody, err = dumpRequestBody(immerseCtx.Request())
	if err != nil {
		return nil, err
	}

	// construct an appropriate protocol factory
	if binary {
		protoFactory = thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{})
	} else {
		protoFactory = thrift.NewTJSONProtocolFactory()
	}

	// read the thrift method and other thrift metadata
	var thriftMethod string
	var messageTypeID thrift.TMessageType
	var seqID int32
	reader := bytes.NewReader(requestBody)
	transport := thrift.NewStreamTransportR(reader)
	proto := protoFactory.GetProtocol(transport)
	thriftMethod, messageTypeID, seqID, err = proto.ReadMessageBegin(context.Background())
	if err != nil {
		return nil, err
	}

	// retrieve the session id
	sessionID := ""
	if thriftMethodHasSession(thriftMethod) {
		// session id is always the first argument
		if _, err = proto.ReadStructBegin(context.Background()); err != nil {
			return nil, err
		}

		_, typeID, fieldID, err := proto.ReadFieldBegin(context.Background())
		if err != nil {
			return nil, err
		}
		if typeID != thrift.STRING || fieldID != 1 {
			return nil, errors.New("Could not get session id from thrift call")
		}

		sessionID, err = proto.ReadString(context.Background())
		if err != nil {
			return nil, err
		}
	}

	ctx = &ThriftReqContext{
		immerseCtx,
		binary,
		requestBody,
		protoFactory,
		thriftMethod,
		messageTypeID,
		seqID,
		sessionID,
	}
	return ctx, nil
}

// IsBinary - returns true if the request was using the binary protocol
func (ctx ThriftReqContext) IsBinary() bool {
	return ctx.binary
}

// GetThriftMethod - retrieve the thrift method for this call
func (ctx ThriftReqContext) GetThriftMethod() string {
	return ctx.thriftMethod
}

// HasSessionID - returns true if this thrift call has a session id
func (ctx ThriftReqContext) HasSessionID() bool {
	return thriftMethodHasSession(ctx.thriftMethod)
}

// IsWhitelistedMethod - returns true for whitelisted thrift methods
func (ctx ThriftReqContext) IsWhitelistedMethod() bool {
	return thriftWhitelistedMethod(ctx.thriftMethod)
}

// GetSessionID - retrieve the session id for this call
func (ctx ThriftReqContext) GetSessionID() string {
	return ctx.sessionID
}

// SetSessionID - update the session id for this call - only call this if the
// thrift method has a session id
func (ctx ThriftReqContext) SetSessionID(sessionID string) (err error) {
	// Thrift has a stream transport with separate read/write buffers, which
	// seems useful here... however, the transport keeps internal state which
	// makes interleaving reads and writes impossible. So, we need separate
	// stream transports for our reads and writes.
	inBuffer := bytes.NewReader(ctx.requestBody)
	outBuffer := &bytes.Buffer{}
	inTransport := thrift.NewStreamTransportR(inBuffer)
	outTransport := thrift.NewStreamTransportW(outBuffer)
	inProto := ctx.protoFactory.GetProtocol(inTransport)
	outProto := ctx.protoFactory.GetProtocol(outTransport)

	// skip message header / write header
	if _, _, _, err = inProto.ReadMessageBegin(context.Background()); err != nil {
		return err
	}
	if err = outProto.WriteMessageBegin(context.Background(), ctx.thriftMethod, ctx.messageTypeID, ctx.seqID); err != nil {
		return err
	}

	// begin reading struct / write struct begin
	var structName string
	if structName, err = inProto.ReadStructBegin(context.Background()); err != nil {
		return err
	}
	if err = outProto.WriteStructBegin(context.Background(), structName); err != nil {
		return err
	}

	// session id is the first argument
	var fieldName string
	if fieldName, _, _, err = inProto.ReadFieldBegin(context.Background()); err != nil {
		return err
	}
	if _, err = inProto.ReadString(context.Background()); err != nil {
		return err
	}
	if err = inProto.ReadFieldEnd(context.Background()); err != nil {
		return err
	}

	// write new session id
	if err = outProto.WriteFieldBegin(context.Background(), fieldName, thrift.STRING, 1); err != nil {
		return nil
	}
	if err = outProto.WriteString(context.Background(), sessionID); err != nil {
		return err
	}
	if err = outProto.WriteFieldEnd(context.Background()); err != nil {
		return err
	}

	// copy the rest of the fields
	if err = copyUntilEndOfStruct(context.Background(), inProto, outProto); err != nil {
		return err
	}

	// and finally end the message - no need to read from inProto 'cause we're
	// done and don't care
	if err = outProto.WriteMessageEnd(context.Background()); err != nil {
		return err
	}

	// flush
	if err = outProto.Flush(context.Background()); err != nil {
		return err
	}

	ctx.updateRequest(outBuffer.Bytes())
	ctx.sessionID = sessionID
	return nil
}

// GetCredentials - if this is a call to connect, we can retrieve the
// credentials. Otherwise, this will throw an error.
func (ctx ThriftReqContext) GetCredentials() (args *heavy.HeavyConnectArgs, err error) {
	reader := bytes.NewReader(ctx.requestBody)
	transport := thrift.NewStreamTransportR(reader)
	proto := ctx.protoFactory.GetProtocol(transport)
	if _, _, _, err = proto.ReadMessageBegin(context.Background()); err != nil {
		return nil, err
	}

	args = &heavy.HeavyConnectArgs{}
	err = args.Read(context.Background(), proto)
	return args, err
}

// SetCredentials - update credentials. This will completely rewrite the thrift
// call to be a connect, so, ya know, don't call this unless that's what ya
// want.
func (ctx ThriftReqContext) SetCredentials(args *heavy.HeavyConnectArgs) (err error) {
	buffer := &bytes.Buffer{}
	transport := thrift.NewStreamTransportRW(buffer)
	proto := ctx.protoFactory.GetProtocol(transport)
	if err = proto.WriteMessageBegin(context.Background(), ctx.thriftMethod, ctx.messageTypeID, ctx.seqID); err != nil {
		return err
	}

	if err = args.Write(context.Background(), proto); err != nil {
		return err
	}

	if err = proto.WriteMessageEnd(context.Background()); err != nil {
		return err
	}

	if err = proto.Flush(context.Background()); err != nil {
		return err
	}

	ctx.updateRequest(buffer.Bytes())
	return nil
}

func (ctx ThriftReqContext) updateRequest(body []byte) {
	request := ctx.Request()
	ctx.requestBody = body
	request.ContentLength = int64(len(body))
	request.Body = io.NopCloser(bytes.NewReader(body))
}

// copyUntilEndOfStruct - copy the rest of the thrift message - it is assumed
// that we are inside a struct when this method is called. I was hoping this
// method wouldn't be necessary, but the JSON proto reads the entire buffer and
// keeps a bunch of internal state with no way to easily copy until the end.
// This function is patterned after the Skip() method in thrift.
func copyUntilEndOfStruct(ctx context.Context, inProto, outProto thrift.TProtocol) (err error) {
	var fieldName string
	var fieldType thrift.TType
	var fieldID int16
	for {
		if fieldName, fieldType, fieldID, err = inProto.ReadFieldBegin(context.Background()); err != nil {
			return err
		}
		if fieldType == thrift.STOP {
			break
		}

		if err = outProto.WriteFieldBegin(ctx, fieldName, fieldType, fieldID); err != nil {
			return err
		}

		if err = copyField(ctx, fieldType, inProto, outProto); err != nil {
			return err
		}

		if err = inProto.ReadFieldEnd(ctx); err != nil {
			return err
		}
		if err = outProto.WriteFieldEnd(ctx); err != nil {
			return err
		}
	}

	if err = outProto.WriteFieldStop(ctx); err != nil {
		return err
	}
	if err = inProto.ReadStructEnd(ctx); err != nil {
		return err
	}
	if err = outProto.WriteStructEnd(ctx); err != nil {
		return err
	}

	return nil
}

// copyField - copies a single thrift field
func copyField(ctx context.Context, fieldType thrift.TType, inProto, outProto thrift.TProtocol) (err error) {
	switch fieldType {
	case thrift.BOOL:
		var v bool
		if v, err = inProto.ReadBool(ctx); err != nil {
			return err
		}
		if err = outProto.WriteBool(ctx, v); err != nil {
			return err
		}

	case thrift.BYTE:
		var v int8
		if v, err = inProto.ReadByte(ctx); err != nil {
			return err
		}
		if err = outProto.WriteByte(ctx, v); err != nil {
			return err
		}

	case thrift.I16:
		var v int16
		if v, err = inProto.ReadI16(ctx); err != nil {
			return nil
		}
		if err = outProto.WriteI16(ctx, v); err != nil {
			return err
		}

	case thrift.I32:
		var v int32
		if v, err = inProto.ReadI32(ctx); err != nil {
			return nil
		}
		if err = outProto.WriteI32(ctx, v); err != nil {
			return err
		}

	case thrift.I64:
		var v int64
		if v, err = inProto.ReadI64(ctx); err != nil {
			return err
		}
		if err = outProto.WriteI64(ctx, v); err != nil {
			return err
		}

	case thrift.DOUBLE:
		var v float64
		if v, err = inProto.ReadDouble(ctx); err != nil {
			return err
		}
		if err = outProto.WriteDouble(ctx, v); err != nil {
			return err
		}

	case thrift.STRING:
		var v string
		if v, err = inProto.ReadString(ctx); err != nil {
			return err
		}
		if err = outProto.WriteString(ctx, v); err != nil {
			return err
		}

	case thrift.STRUCT:
		var v string
		if v, err = inProto.ReadStructBegin(ctx); err != nil {
			return err
		}
		if err = outProto.WriteStructBegin(ctx, v); err != nil {
			return err
		}
		if err = copyUntilEndOfStruct(ctx, inProto, outProto); err != nil {
			return err
		}

	case thrift.MAP:
		var keyType, valueType thrift.TType
		var size int
		if keyType, valueType, size, err = inProto.ReadMapBegin(ctx); err != nil {
			return err
		}
		if err = outProto.WriteMapBegin(ctx, keyType, valueType, size); err != nil {
			return err
		}

		for i := 0; i < size; i++ {
			if err = copyField(ctx, keyType, inProto, outProto); err != nil {
				return err
			}
			if err = copyField(ctx, valueType, inProto, outProto); err != nil {
				return err
			}
		}

		if err = inProto.ReadMapEnd(ctx); err != nil {
			return err
		}
		if err = outProto.WriteMapEnd(ctx); err != nil {
			return err
		}

	case thrift.SET:
		var elemType thrift.TType
		var size int
		if elemType, size, err = inProto.ReadSetBegin(ctx); err != nil {
			return err
		}
		if err = outProto.WriteSetBegin(ctx, elemType, size); err != nil {
			return err
		}

		for i := 0; i < size; i++ {
			if err = copyField(ctx, elemType, inProto, outProto); err != nil {
				return err
			}
		}

		if err = inProto.ReadSetEnd(ctx); err != nil {
			return err
		}
		if err = outProto.WriteSetEnd(ctx); err != nil {
			return err
		}

	case thrift.LIST:
		var elemType thrift.TType
		var size int
		if elemType, size, err = inProto.ReadListBegin(ctx); err != nil {
			return err
		}
		if err = outProto.WriteListBegin(ctx, elemType, size); err != nil {
			return err
		}

		for i := 0; i < size; i++ {
			if err = copyField(ctx, elemType, inProto, outProto); err != nil {
				return err
			}
		}

		if err = inProto.ReadListEnd(ctx); err != nil {
			return err
		}
		if err = outProto.WriteListEnd(ctx); err != nil {
			return err
		}

	default:
		return errors.New("Unknown thrift field type")
	}

	return nil
}
