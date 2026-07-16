// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package dashboardarchive

import (
	"archive/zip"
	"io"
	"time"
)

type File struct {
	Name     string
	Contents []byte
}

func Write(writer io.Writer, files []File) (err error) {
	archive := zip.NewWriter(writer)
	defer func() {
		if closeErr := archive.Close(); err == nil {
			err = closeErr
		}
	}()

	for _, file := range files {
		header := &zip.FileHeader{
			Name:   file.Name,
			Method: zip.Store,
		}
		header.SetMode(0444)
		header.SetModTime(time.Time{})
		entry, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := entry.Write(file.Contents); err != nil {
			return err
		}
	}

	return nil
}
