package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/blefort/DTDParser/DTD"
)

// loadElementTests Load element tests
func loadElementTests(file string) []DTD.Element {
	var tests []DTD.Element
	loadJSON(file, &tests)
	return tests
}

// TestParseCommentBlock Test parser for result
func TestParseElementBlock(t *testing.T) {
	// orginal file
	testElementDTD(t, "tests/element.dtd")
	testElementDTD(t, "tmp/element.dtd")
}

func testElementDTD(t *testing.T, path string) {
	var tests []DTD.Element

	// New parser
	p := newParser()

	p.Parse(path)
	tests = loadElementTests("tests/element.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]

		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name))
		t.Run("Check value", checkStrValue(parsedBlock.GetValue(), test.Value))
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
