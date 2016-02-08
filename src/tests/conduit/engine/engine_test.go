package engine

import (
	"conduit/engine"
	"testing"
)

func TestExecuteFunction(t *testing.T) {
	var script = `
		var test = "test";
		test2 = "test";
		$local = function() {
			return true;
		};
	`
	eng := engine.New()
	res, err := eng.ExecuteFunction("$local", script)
	if err != nil {
		t.Fatal(err)
	}
	if res != "true" {
		t.Fatal("No Return value")
	}
}

func TestValidate(t *testing.T) {
	var script = `
		var test = 
	`
	eng := engine.New()
	err := eng.Validate(script)
	if err == nil {
		t.Fatal("No error reported for bad syntax")
	}
}
