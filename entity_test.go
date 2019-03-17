package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"
)

// EntityTestResult struct to test entity
type EntityTestResult struct {
	Name      string
	Value     string
	Parameter bool
	Url       string
}

// loadEntityTests Load entity tests
func loadEntityTests(file string) []EntityTestResult {
	var tests []EntityTestResult
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
	var tests []EntityTestResult

	// New parser
	p := newParser()
	p.Parse(path)

	tests = loadEntityTests("tests/entity.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]

		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name))
		t.Run("Check value", checkStrValue(parsedBlock.GetValue(), test.Value))
		t.Run("Check Parameter", checkBoolValue(parsedBlock.GetParameter(), test.Parameter))
		t.Run("Check Url", checkStrValue(parsedBlock.GetUrl(), test.Url))
	}
	t.Run("Render DTD", render(p))
}
