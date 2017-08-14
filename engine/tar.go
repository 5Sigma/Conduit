package engine

import (
	"archive/tar"
	"fmt"
	"github.com/robertkrimen/otto"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func _tar_compress(call otto.FunctionCall) otto.Value {
	var (
		baseDir string
	)
	source, _ := call.Argument(0).ToString()
	target, _ := call.Argument(1).ToString()

	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf("%s.tar", filename))
	tarfile, err := os.Create(target)
	if err != nil {
		jsThrow(call, err)
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		jsThrow(call, err)
	}

	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	err = filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tarball.WriteHeader(header); err != nil {
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
			_, err = io.Copy(tarball, file)
			return err
		})
	return otto.Value{}
}

func _tar_decompress(call otto.FunctionCall) otto.Value {
	source, _ := call.Argument(0).ToString()
	target, _ := call.Argument(1).ToString()

	reader, err := os.Open(source)
	if err != nil {
		jsThrow(call, err)
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			jsThrow(call, err)
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				jsThrow(call, err)
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
			info.Mode())
		if err != nil {
			jsThrow(call, err)
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			jsThrow(call, err)
		}
	}
	return otto.Value{}
}
