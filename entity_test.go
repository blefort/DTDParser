package main

import (
	"testing"

	"github.com/blefort/DTDParser/DTD"

	log "github.com/Sirupsen/logrus"
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
	testEntityDTD(t, "tests/entity.dtd")

	// - load the generated DTD
	// - compare it to data stored in a json file
	log.Warn("Start tests on generated 'tmp/entity.dtd'")
	testEntityDTD(t, "tmp/entity.dtd")
}

// testEntityDTD Main testing func for entity
func testEntityDTD(t *testing.T, path string) {
	var tests []DTD.Entity

	// New parser
	p := newParser()
	p.Parse(path)

	tests = loadEntityTests("tests/entity.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		entityBlock := p.Collection[idx].(*DTD.Entity)

		t.Run("Check name", checkStrValue(entityBlock.Name, test.Name))
		t.Run("Check value", checkStrValue(entityBlock.Value, test.Value))
		t.Run("Check Parameter", checkBoolValue(entityBlock.Parameter, test.Parameter))
		t.Run("Check System", checkBoolValue(entityBlock.System, test.System))
		t.Run("Check Public", checkBoolValue(entityBlock.Public, test.Public))
		t.Run("Check External Entity", checkBoolValue(entityBlock.ExternalDTD, test.ExternalDTD))
		t.Run("Check Url", checkStrValue(entityBlock.Url, test.Url))
	}
	t.Run("Render DTD", render(p))
}
