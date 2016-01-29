package engine

import (
	"conduit/engine"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func fileExists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}

func createTestFile() {
	var filepath string = "test"
	if !fileExists(filepath) {
		os.Create(filepath)
	}

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString("test")
	if err != nil {
		panic(err)
	}
}

//EXISTS
func TestFileExists(t *testing.T) {
	createTestFile()

	err := engine.Execute(`if (!$file.exists('test')) { throw new Error(); }`)
	if err != nil {
		t.Error(err)
	}

	err = engine.Execute(`if ($file.exists('test2')) { throw new Error(); }`)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove("test")
}

//WRITE
func TestFileWrite(t *testing.T) {
	err := engine.Execute(`$file.write('write_test.txt', 'test');`)
	if err != nil {
		t.Error(err)
	}

	data, err := ioutil.ReadFile("write_test.txt")
	if err != nil {
		t.Error(err)
	}
	if string(data) != "test" {
		t.Error("expecting contents to be 'test' but got", string(data))
	}
	defer os.Remove("write_test.txt")
}

func TestFileWriteBadPath(t *testing.T) {
	err := engine.Execute(`$file.write('/blah/test.txt', 'test');`)
	if err == nil {
		t.Fail()
	}
}

//COPY
func TestFileCopy(t *testing.T) {
	createTestFile()

	err := engine.Execute(`$file.copy('test', 'test1');`)
	if err != nil {
		t.Error(err)
	}
	if !fileExists("test1") {
		t.Error("destination file does not exist")
	}
	defer os.Remove("test")
	defer os.Remove("test1")
}

func TestFileCopyBadPath(t *testing.T) {
	err := engine.Execute(`$file.copy('test', '/blah/test1');`)
	if err == nil {
		t.Fail()
	}
}

//MOVE
func TestFileMove(t *testing.T) {
	createTestFile()

	err := engine.Execute(`$file.move('test', 'test1');`)
	if err != nil {
		t.Error(err)
	}
	if !fileExists("test1") {
		t.Error("destination file does not exist")
	}

	defer os.Remove("test1")
}

func TestFileMoveBadPath(t *testing.T) {
	createTestFile()

	err := engine.Execute(`$file.move('test', '/blah/test');`)
	if err == nil {
		t.Fail()
	}

	defer os.Remove("test")
}

//SIZE
func TestFileSize(t *testing.T) {
	createTestFile()

	err := engine.Execute(`if ($file.size('test') != 4) { throw new Error(); }`)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove("test")
}

//DELETE
func TestFileDelete(t *testing.T) {
	createTestFile()

	err := engine.Execute(`$file.delete('test')`)
	if err != nil {
		t.Error(err)
	}
}

//MKDIR
func TestFileMkdir(t *testing.T) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	err := engine.Execute(`$file.mkdir('` + dir + `/test/test1');`)
	if err != nil {
		t.Error(err)
	}

	if !fileExists(dir + "/test/test1") {
		t.Error("Directory does not exist: ", dir)
	}

	defer os.RemoveAll(dir + "/test/test1")
}

//READSTRING
func TestFileReadString(t *testing.T) {
	createTestFile()

	err := engine.Execute(`if ($file.readString('test') != 'test') { throw new Error(); }`)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove("test")
}
