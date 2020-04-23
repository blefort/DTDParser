package main

import (
	"testing"

	"github.com/blefort/DTDParser/DTD"
	log "github.com/sirupsen/logrus"
)

// loadElementTests Load element tests
func loadElementTests(file string) []DTD.Element {
	var tests []DTD.Element
	loadJSON(file, &tests)
	return tests
}

// TestParseCommentBlock Test parser for result
func TestParseElementBlock(t *testing.T) {
	// - parse the DTD test
	// - compare it to data stored in a json file
	// - render it in the tmp dir
	testElementDTD(t, "tests/element.dtd", true)

	// - load the generated DTD
	// - compare it to data stored in a json file
	testElementDTD(t, "tmp/element.dtd", false)
}

// testElementDTD main tests for elementDTD
func testElementDTD(t *testing.T, path string, recreate bool) {
	var tests []DTD.Element
	var dir string

	if recreate {
		dir = "tmp"
	} else {
		dir = "tmp2"
	}

	// New parser
	p := newParser(dir)

	p.Parse(path)
	tests = loadElementTests("tests/element.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]

		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name, test))
		t.Run("Check value", checkStrValue(parsedBlock.GetValue(), test.Value, test))
	}
	t.Run("Render DTD", render(p))
}

// TestElementPanic Test func that should never be called
func TestElementPanic(t *testing.T) {
	assertPanic(t, ElementExported)
	assertPanic(t, ElementGetParameter)
	assertPanic(t, ElementGetUrl)
}

// ElementExported() Helper to test DTD.Element.GetExported()
func ElementExported() {
	var e DTD.Element
	ret := e.GetExported()
	log.Tracef("ElementExported( return %t", ret)
}

// ElementExported() Helper to test DTD.Element.GetParameter()
func ElementGetParameter() {
	var e DTD.Element
	ret := e.GetParameter()
	log.Tracef("ElementExported( return %t", ret)
}

// ElementExported() Helper to test DTD.Element.GetUrl()
func ElementGetUrl() {
	var e DTD.Element
	ret := e.GetUrl()
	log.Tracef("ElementUrl( return %s", ret)
}
