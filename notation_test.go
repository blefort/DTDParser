package main

import (
	"testing"

	"github.com/blefort/DTDParser/DTD"
	"github.com/blefort/DTDParser/formatter"
	DTDformat "github.com/blefort/DTDParser/formatter/DTDformat"
)

// loadElementTests Load element tests
func loadNotationTests(file string) []DTD.Notation {
	var tests []DTD.Notation
	loadJSON(file, &tests)
	return tests
}

// TestParseCommentBlock Test parser for result
func TestParseNotationBlock(t *testing.T) {
	// - parse the DTD test
	// - compare it to data stored in a json file
	// - render it in the tmp dir
	t.Log("Start tests on 'tests/notation.dtd'")
	testNotationDTD(t, "tests/notation.dtd", true)

	// - load the generated DTD
	// - compare it to data stored in a json file
	//	t.Log("Start tests on 'tmp/notation.dtd'")
	testNotationDTD(t, "tmp/notation.dtd", false)
}

// testCommentDTD Main func holding tests
func testNotationDTD(t *testing.T, path string, recreate bool) {
	var tests []DTD.Notation

	// New parser
	var dir string

	if recreate {
		dir = "tmp"
	} else {
		dir = "tmp2"
	}

	// New parser
	p := newParser(dir)

	// new DTD formatter
	format := DTDformat.New(log)
	formatter := formatter.NewFormatter(p, format, dir, "notation.dtd", log)

	p.Parse(path)

	tests = loadNotationTests("tests/notation.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx].(*DTD.Notation)

		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name, parsedBlock, test))
		t.Run("Check PublicID", checkStrValue(parsedBlock.PublicID, test.PublicID, parsedBlock, test))
		t.Run("Check SystemID", checkStrValue(parsedBlock.SystemID, test.SystemID, parsedBlock, test))
		t.Run("Check System", checkBoolValue(parsedBlock.System, test.System, parsedBlock, test))
		t.Run("Check Public", checkBoolValue(parsedBlock.Public, test.Public, parsedBlock, test))
	}
	t.Run("Render DTD", render(formatter))
}
