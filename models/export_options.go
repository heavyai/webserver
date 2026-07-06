// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

// ExportOptions - expected form fields for /export endpoint
type ExportOptions struct {
	SQL         string `form:"sql"`
	FileName    string `form:"filename"`
	LayerName   string `form:"layername"`
	FileType    string `form:"filetype"`
	Compression string `form:"compression"`
}
