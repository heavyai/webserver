// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"bytes"
	"os"
	"time"
)

// VirtualFileInfo - Virtual File Info model
type VirtualFileInfo struct {
	FileName string
	Data     []byte
}

// Name - Virtual File Info model Name method
func (vfi VirtualFileInfo) Name() string { return vfi.FileName }

// Size - Virtual File Info model Size method
func (vfi VirtualFileInfo) Size() int64 { return int64(len(vfi.Data)) }

// Mode - Virtual File Info model Mode method
func (vfi VirtualFileInfo) Mode() os.FileMode { return 0444 } // Read for all

// ModTime - Virtual File Info model ModTime method
func (vfi VirtualFileInfo) ModTime() time.Time { return time.Time{} }

// IsDir - Virtual File Info model IsDir method
func (vfi VirtualFileInfo) IsDir() bool { return false }

// Sys - Virtual File Info model Sys method
func (vfi VirtualFileInfo) Sys() interface{} { return nil }

// VirtualFile - Virtual File model
type VirtualFile struct {
	*bytes.Reader
	VFI VirtualFileInfo
}

// Close - Virtual File model Close method
func (vf *VirtualFile) Close() error { return nil } // Noop, nothing to do

// Readdir - Virtual File model Readdir method
func (vf *VirtualFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil // We are not a directory but a single file
}

// Stat - Virtual File model Stat method
func (vf *VirtualFile) Stat() (os.FileInfo, error) {
	return vf.VFI, nil
}

// Source:
// https://stackoverflow.com/questions/52697277/simples-way-to-make-a-byte-into-a-virtual-file-object-in-golang
