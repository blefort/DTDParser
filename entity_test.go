package main

import (
	"testing"

	"github.com/blefort/DTDParser/DTD"
)

// loadEntityTests Load entity tests
func loadEntityTests(file string) []DTD.Entity {
	var tests []DTD.Entity
	loadJSON(file, &tests)
	return tests
}

func TestParseEntityBlock(t *testing.T) {
	// - parse the DTD test
	// - compare it to data stored in a json file
	// - render it in the tmp dir
	log.Warn("Start tests on 'tests/entity.dtd'")
	testEntityDTD(t, "tests/entity.dtd", true)

	// - load the generated DTD
	// - compare it to data stored in a json file
	//	log.Warn("Start tests on generated 'tmp/entity.dtd'")
	//testEntityDTD(t, "tmp/entity.dtd", false)
}

// testEntityDTD Main testing func for entity
func testEntityDTD(t *testing.T, path string, recreate bool) {
	var tests []DTD.Entity

	var dir string

	if recreate {
		dir = "tmp"
	} else {
		dir = "tmp2"
	}

	// New parser
	p := newParser(dir)
	p.Parse(path)

	tests = loadEntityTests("tests/entity.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		entityBlock := p.Collection[idx].(*DTD.Entity)

		t.Run("Check name", checkStrValue(entityBlock.Name, test.Name, entityBlock, test))
		t.Run("Check value", checkStrValue(entityBlock.Value, test.Value, entityBlock, test))
		t.Run("Check Parameter", checkBoolValue(entityBlock.Parameter, test.Parameter, entityBlock, test))
		t.Run("Check System", checkBoolValue(entityBlock.System, test.System, entityBlock, test))
		t.Run("Check Public", checkBoolValue(entityBlock.Public, test.Public, entityBlock, test))
		t.Run("Check External Entity", checkBoolValue(entityBlock.IsExternal, test.IsExternal, entityBlock, test))
		t.Run("Check Url", checkStrValue(entityBlock.Url, test.Url, entityBlock, test))
	}
	t.Run("Render DTD", render(p))
}
