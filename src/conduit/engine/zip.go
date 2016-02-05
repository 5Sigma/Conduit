package engine

import (
	"archive/zip"
	"github.com/robertkrimen/otto"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func _zip_compress(call otto.FunctionCall) otto.Value {
	source, _ := call.Argument(0).ToString()
	target, _ := call.Argument(1).ToString()

	zipfile, err := os.Create(target)
	if err != nil {
		jsThrow(call, err)
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		jsThrow(call, err)
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		jsThrow(call, err)
	}
	return otto.Value{}
}

func _zip_decompress(call otto.FunctionCall) otto.Value {
	archive, _ := call.Argument(0).ToString()
	target, _ := call.Argument(1).ToString()
	reader, err := zip.OpenReader(archive)
	if err != nil {
		jsThrow(call, err)
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		jsThrow(call, err)
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			jsThrow(call, err)
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			jsThrow(call, err)
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			jsThrow(call, err)
		}
	}
	return otto.Value{}
}
