package engine

import (
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
	eng := New()
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
	eng := New()
	err := eng.Validate(script)
	if err == nil {
		t.Fatal("No error reported for bad syntax")
	}
}

func TestGetStringVar(t *testing.T) {
	script := ` var test = "test123"; `
	eng := New()
	v, err := eng.GetVar("test", script)
	if err != nil {
		t.Fatal(err)
	}
	if v.(string) != "test123" {
		t.Fatal("Return value not valid")
	}
}

func TestGetNumberVar(t *testing.T) {
	script := ` var test = 1`
	eng := New()
	v, err := eng.GetVar("test", script)
	if err != nil {
		t.Fatal(err)
	}
	if v.(int64) != 1 {
		t.Fatal("Return value not valid")
	}
}
