// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"
	"strings"
	"testing"

	"github.com/mholt/archives"

	"github.com/heavyai/webserver/models"
)

func TestWriteDashboardArchiveProducesReadableZip(t *testing.T) {
	expected := map[string]string{
		"Sales Dashboard_8.4.0_2026-11-10T10:22:00-07:00.json":     "Sales Dashboard\n{\"owner\":\"analyst\"}\n{\"charts\":[1,2,3]}",
		"Unicode Δ Dashboard_8.4.0_2026-11-10T10:22:00-07:00.json": strings.Repeat("dashboard-state-", 4096),
	}

	files := make([]archives.FileInfo, 0, len(expected))
	for name, contents := range expected {
		files = append(files, dashboardArchiveTestFile(name, []byte(contents)))
	}

	var output bytes.Buffer
	if err := writeDashboardArchive(&output, files); err != nil {
		t.Fatal(err)
	}

	reader, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil {
		t.Fatalf("dashboard export is not a readable ZIP archive: %v", err)
	}
	if len(reader.File) != len(expected) {
		t.Fatalf("expected %d archived dashboards, got %d", len(expected), len(reader.File))
	}

	for _, archivedFile := range reader.File {
		expectedContents, ok := expected[archivedFile.Name]
		if !ok {
			t.Errorf("unexpected file %q in dashboard archive", archivedFile.Name)
			continue
		}

		file, err := archivedFile.Open()
		if err != nil {
			t.Fatal(err)
		}
		contents, readErr := io.ReadAll(file)
		closeErr := file.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		if string(contents) != expectedContents {
			t.Errorf("unexpected contents for %q", archivedFile.Name)
		}
		delete(expected, archivedFile.Name)
	}

	if len(expected) != 0 {
		t.Fatalf("dashboard archive omitted files: %v", expected)
	}
}

func dashboardArchiveTestFile(name string, data []byte) archives.FileInfo {
	fileInfo := models.VirtualFileInfo{FileName: name, Data: data}
	return archives.FileInfo{
		FileInfo:      fileInfo,
		NameInArchive: name,
		Open: func() (fs.File, error) {
			return &models.VirtualFile{
				Reader: bytes.NewReader(data),
				VFI:    fileInfo,
			}, nil
		},
	}
}
