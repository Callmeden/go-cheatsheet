// main_test.go
package fileprocessor

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func createTempFile(t *testing.T, name, content string) {
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file %s: %v", name, err)
	}
	t.Cleanup(func() { os.Remove(name) })
}

func TestProcessFiles(t *testing.T) {
	t.Run("all files exist", func(t *testing.T) {
		createTempFile(t, "file1.txt", "line1\nline2\n")
		createTempFile(t, "file2.txt", "line1\n")

		lines, err := ProcessFiles(FS, "file1.txt", "file2.txt")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if lines != 3 {
			t.Errorf("expected 3 total lines, got %d", lines)
		}
	})

	t.Run("one file does not exist", func(t *testing.T) {
		createTempFile(t, "file3.txt", "line1\n")

		lines, err := ProcessFiles(FS, "file3.txt", "nonexistent.txt")

		var fileErr *FileError
		if !errors.As(err, &fileErr) {
			t.Fatalf("expected FileError, got %v", err)
		}
		if fileErr.Path != "nonexistent.txt" {
			t.Errorf("expected FileError.Path = nonexistent.txt, got %v", fileErr.Path)
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("expected error is os.ErrNotExist, got %v", err)
		}
		if lines != 1 {
			t.Errorf("expected 1 total line, got %d", lines)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		lines, err := ProcessFiles(FS, "nonexistent1.txt", "nonexistent2.txt")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		var joinErr interface{ Unwrap() []error }
		if !errors.As(err, &joinErr) {
			t.Fatalf("expected joined error, got %v", err)
		}
		errList := joinErr.Unwrap()
		if len(errList) != 2 {
			t.Fatalf("expected 2 errors, got %d", len(errList))
		}

		for i, e := range []string{"nonexistent1.txt", "nonexistent2.txt"} {
			var fe *FileError
			if !errors.As(errList[i], &fe) || fe.Path != e || !errors.Is(fe.Err, os.ErrNotExist) {
				t.Errorf("unexpected error[%d]: %v", i, errList[i])
			}
		}

		if lines != 0 {
			t.Errorf("expected 0 total lines, got %d", lines)
		}
	})
}

func TestProcessFiles_WithMocks(t *testing.T) {
	mockOpener := MockOpener{
		Files: map[string]*MockFile{
			"read_error.txt": {
				Content: "line1\nline2\n",
				ReadErr: errors.New("mock read error"),
			},
			"close_error.txt": {
				Content:  "line1\nline2\n",
				CloseErr: errors.New("mock close error"),
			},
		},
	}

	_, err := ProcessFiles(mockOpener, "read_error.txt", "close_error.txt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var joinErr interface{ Unwrap() []error }
	if !errors.As(err, &joinErr) {
		t.Fatalf("expected joined error, got %v", err)
	}

	errorsList := joinErr.Unwrap()
	if len(errorsList) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errorsList))
	}

	if !strings.Contains(errorsList[0].Error(), "mock read error") {
		t.Errorf("expected read error, got %v", errorsList[0])
	}
	if !strings.Contains(errorsList[1].Error(), "mock close error") {
		t.Errorf("expected close error, got %v", errorsList[1])
	}
}

type MockFile struct {
	Content     string
	ReadErr     error
	CloseErr    error
	ReadCounter int
}

func (m *MockFile) Read(p []byte) (int, error) {
	if m.ReadErr != nil {
		return 0, m.ReadErr
	}
	m.ReadCounter++
	copy(p, m.Content)
	return len(m.Content), io.EOF
}

func (m *MockFile) Close() error {
	return m.CloseErr
}

type MockOpener struct {
	Files map[string]*MockFile
	Errs  map[string]error
}

func (mo MockOpener) Open(name string) (File, error) {
	if err, ok := mo.Errs[name]; ok {
		return nil, err
	}
	if file, ok := mo.Files[name]; ok {
		return file, nil
	}
	return nil, os.ErrNotExist
}
