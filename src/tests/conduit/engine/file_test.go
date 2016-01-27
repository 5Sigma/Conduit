package engine

import (
	"conduit/engine"
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	_, err := os.Create("test")
	if err != nil {
		t.Fail()
	}
	defer os.Remove("test")

	err = engine.Execute(`if (!$file.exists('test')) { throw new Error(); }`)
	if err != nil {
		t.Fail()
	}

	err = engine.Execute(`if ($file.exists('test2')) { throw new Error(); }`)
	if err != nil {
		t.Fail()
	}
}
