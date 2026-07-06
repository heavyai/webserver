// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/labstack/echo/v4"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/internal/util"
	"github.com/heavyai/webserver/models"
)

// ThriftAuthEncryption - Thrift Auth Encryption Middleware
func ThriftAuthEncryption(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if !config.Enable.EncryptedCredentials {
			return next(ctx)
		}
		immerseCtx := ctx.(*models.ThriftReqContext)

		if SkipThriftNonConnect(immerseCtx) {
			return next(ctx)
		}

		err := rewriteBodyThriftCredentials(immerseCtx)
		if err != nil {
			return err
		}

		return next(immerseCtx)
	}
}

func rewriteBodyThriftCredentials(ctx *models.ThriftReqContext) (err error) {
	args, err := ctx.GetCredentials()
	if err != nil {
		return err
	}

	args.User, err = util.DecryptString(args.User)
	if err != nil {
		config.Log.Error("Could not decrypt username: ", err)
		return err
	}
	args.Passwd, err = util.DecryptString(args.Passwd)
	if err != nil {
		config.Log.Error("Could not decrypt password: ", err)
		return err
	}
	args.Dbname, err = util.DecryptString(args.Dbname)
	if err != nil {
		config.Log.Error("Could not decrypt database name: ", err)
		return err
	}

	err = ctx.SetCredentials(args)
	if err != nil {
		config.Log.Error("Could not update credentials: ", err)
	}

	return nil
}
