// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

import (
	"net/url"
)

const (
	// DefaultMetricsIntercomKey - constant
	DefaultMetricsIntercomKey = "x4gxr1qv"
	// DefaultMetricsPendoKey - constant
	DefaultMetricsPendoKey = "8c6103ad-dffb-462e-46c5-e050f9d739f6"
	// DefaultMetricsPosthogKey - constant
	DefaultMetricsPosthogKey = "7u8fgnpiB5ejisehRjsvjo6I6bLxnItx96y0TBMnRZM"
)

var (
	// PosthogAPIHost - constant
	PosthogAPIHost = &url.URL{
		Scheme:     "https",
		Host:       "us.i.posthog.com",
		Path:       "",
		RawPath:    "",
		Opaque:     "",
		User:       nil,
		ForceQuery: false,
		RawQuery:   "",
		Fragment:   "",
	}
)
