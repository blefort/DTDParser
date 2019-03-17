package main

import (
	"testing"

	"github.com/blefort/DTDParser/DTD"
)

// loadElementTests Load element tests
func loadNotationTests(file string) []DTD.Notation {
	var tests []DTD.Notation
	loadJSON(file, &tests)
	return tests
}

// TestParseCommentBlock Test parser for result
func TestParseNotationBlock(t *testing.T) {
	var tests []DTD.Notation

	// New parser
	p := newParser()
	p.Parse("tests/notation.dtd")

	tests = loadNotationTests("tests/notation.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx].(*DTD.Notation)

		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name))
		t.Run("Check value", checkStrValue(parsedBlock.GetValue(), test.Value))
		t.Run("Check ID", checkStrValue(parsedBlock.ID, test.ID))
		t.Run("Check System", checkBoolValue(parsedBlock.System, test.System))
		t.Run("Check Public", checkBoolValue(parsedBlock.Public, test.Public))
	}
	t.Run("Render DTD", render(p))
}
