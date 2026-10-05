// Package fileprocessor предоставляет утилиты для обработки файлов.
package fileprocessor

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

// ========================================================
// Код ниже не менять!
// ========================================================

// File - интерфейс файла
type File interface {
	io.Reader
	io.Closer
}

// FileOpener - интерфейс провайдера файла
type FileOpener interface {
	Open(name string) (File, error)
}

var _ FileOpener = (*FileSystem)(nil)

// FileSystem - реализация FileOpener
type FileSystem struct{}

// Open - открывает и возращает файл
func (fs FileSystem) Open(name string) (File, error) {
	return os.Open(name)
}

var FS = FileSystem{}

type FileError struct {
	Path string
	Err  error
}

// ========================================================
// Ваша реализация ниже
// ========================================================

func (e *FileError) Error() string {
	return fmt.Sprintf("ошибка с файлом: {%s}: {%s}", e.Path, e.Err)
}

func (e *FileError) Unwrap() error {
	return e.Err
}

var _ error = (*FileError)(nil)

// ProcessFiles - принимает FileOpener и пути до файлов
func ProcessFiles(fs FileOpener, paths ...string) (int, error) {
	resultCount := 0
	var errs error = nil
	for _, path := range paths {
		if linesCount, err := ProcessFile(fs, path); err != nil {
			errs = errors.Join(errs, err)
		} else {
			resultCount += linesCount
		}
	}
	return resultCount, errs
}

func ProcessFile(fs FileOpener, path string) (linesCount int, errs error) {
	file, err := fs.Open(path)
	if err != nil {
		return 0, &FileError{Path: path, Err: err}
	}

	defer func() {
		err := file.Close()
		if err != nil {
			errs = errors.Join(errs, &FileError{Path: path, Err: err})
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linesCount++
	}
	if err := scanner.Err(); err != nil {
		errs = &FileError{Path: path, Err: err}
	}

	return linesCount, errs
}

// ========================================================
