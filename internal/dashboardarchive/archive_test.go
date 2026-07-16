// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package dashboardarchive

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestWrite(t *testing.T) {
	files := []File{
		{Name: "first.json", Contents: []byte("first dashboard")},
		{Name: "second.json", Contents: []byte("second dashboard")},
	}

	var output bytes.Buffer
	if err := Write(&output, files); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	archive, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader() error: %v", err)
	}
	if len(archive.File) != len(files) {
		t.Fatalf("archive contains %d files, want %d", len(archive.File), len(files))
	}

	for index, archivedFile := range archive.File {
		want := files[index]
		if archivedFile.Name != want.Name {
			t.Errorf("archive file %d name = %q, want %q", index, archivedFile.Name, want.Name)
		}
		if archivedFile.Method != zip.Store {
			t.Errorf("archive file %q method = %d, want Store", archivedFile.Name, archivedFile.Method)
		}
		if archivedFile.Mode().Perm() != 0444 {
			t.Errorf("archive file %q mode = %o, want 0444", archivedFile.Name, archivedFile.Mode().Perm())
		}

		reader, err := archivedFile.Open()
		if err != nil {
			t.Fatalf("open archive file %q: %v", archivedFile.Name, err)
		}
		contents, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			t.Fatalf("read archive file %q: %v", archivedFile.Name, readErr)
		}
		if closeErr != nil {
			t.Fatalf("close archive file %q: %v", archivedFile.Name, closeErr)
		}
		if !bytes.Equal(contents, want.Contents) {
			t.Errorf("archive file %q contents = %q, want %q", archivedFile.Name, contents, want.Contents)
		}
	}
}

type failingWriter struct {
	err error
}

func (writer failingWriter) Write([]byte) (int, error) {
	return 0, writer.err
}

func TestWriteReturnsWriterError(t *testing.T) {
	want := errors.New("write failed")
	err := Write(failingWriter{err: want}, []File{
		{Name: "dashboard.json", Contents: []byte("dashboard")},
	})
	if !errors.Is(err, want) {
		t.Fatalf("Write() error = %v, want %v", err, want)
	}
}
