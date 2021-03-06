package engine

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

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
	eng := New()
	err := eng.Execute(`if (!$file.exists('test')) { throw new Error(); }`)
	if err != nil {
		t.Error(err)
	}

	err = eng.Execute(`if ($file.exists('test2')) { throw new Error(); }`)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove("test")
}

//WRITE
func TestFileWrite(t *testing.T) {
	eng := New()
	err := eng.Execute(`$file.write('write_test.txt', 'test');`)
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
	eng := New()
	err := eng.Execute(`$file.write('/blah/test.txt', 'test');`)
	if err == nil {
		t.Fail()
	}
}

//COPY
func TestFileCopy(t *testing.T) {
	createTestFile()
	eng := New()
	err := eng.Execute(`$file.copy('test', 'test1');`)
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
	eng := New()
	err := eng.Execute(`$file.copy('test', '/blah/test1');`)
	if err == nil {
		t.Fail()
	}
}

//MOVE
func TestFileMove(t *testing.T) {
	createTestFile()
	eng := New()
	err := eng.Execute(`$file.move('test', 'test1');`)
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
	eng := New()
	err := eng.Execute(`$file.move('test', '/blah/test');`)
	if err == nil {
		t.Fail()
	}

	defer os.Remove("test")
}

//SIZE
func TestFileSize(t *testing.T) {
	createTestFile()
	eng := New()
	err := eng.Execute(`if ($file.size('test') != 4) { throw new Error(); }`)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove("test")
}

//DELETE
func TestFileDelete(t *testing.T) {
	createTestFile()
	eng := New()
	err := eng.Execute(`$file.delete('test')`)
	if err != nil {
		t.Error(err)
	}
}

//MKDIR
func TestFileMkdir(t *testing.T) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	eng := New()
	err := eng.Execute(`$file.mkdir('` + dir + `/test/test1');`)
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
	eng := New()
	err := eng.Execute(`if ($file.readString('test') != 'test') { throw new Error(); }`)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove("test")
}

func TestFileInfo(t *testing.T) {
	createTestFile()
	defer os.Remove("test")
	script := `if ($file.info("test").size != 4) {throw new Error();}`
	eng := New()
	err := eng.Execute(script)
	if err != nil {
		t.Fatal(err)
	}
}
