package engine

import (
	"compress/gzip"
	"fmt"
	"github.com/robertkrimen/otto"
	"io"
	"os"
	"path/filepath"
)

func _gzip_compress(call otto.FunctionCall) otto.Value {
	source, _ := call.Argument(0).ToString()
	target, _ := call.Argument(1).ToString()
	reader, err := os.Open(source)
	if err != nil {
		jsThrow(call, err)
	}

	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf("%s.gz", filename))
	writer, err := os.Create(target)
	if err != nil {
		jsThrow(call, err)
	}

	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = filename
	defer archiver.Close()
	_, err = io.Copy(archiver, reader)
	if err != nil {
		jsThrow(call, err)
	}

	return otto.Value{}
}

func _gzip_decompress(call otto.FunctionCall) otto.Value {
	source, _ := call.Argument(0).ToString()
	target, _ := call.Argument(1).ToString()

	reader, err := os.Open(source)
	if err != nil {
		jsThrow(call, err)
	}

	archive, err := gzip.NewReader(reader)
	if err != nil {
		jsThrow(call, err)
	}

	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		jsThrow(call, err)
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	if err != nil {
		jsThrow(call, err)
	}
	return otto.Value{}
}
