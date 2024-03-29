package main

import (
	"testing"
)

// CommentTestResult struct to test comment
// @todo Replace with DTD.Comment struct
type CommentTestResult struct {
	Name  string
	Value string
}

// loadCommentTests Load comment tests
func loadCommentTests(file string) []CommentTestResult {
	var tests []CommentTestResult
	loadJSON(file, &tests)
	return tests
}

// TestParseCommentBlock Test parser for result
func TestParseCommentBlock(t *testing.T) {
	// - parse the DTD test
	// - compare it to data stored in a json file
	// - render it in the tmp dir
	t.Log("Start tests on 'tests/comment.dtd'")
	testCommentDTD(t, "tests/comment.dtd", true)

	// - load the generated DTD
	// - compare it to data stored in a json file
	t.Log("Start tests on 'tmp/comment.dtd'")
	testCommentDTD(t, "tmp/comment.dtd", false)
}

// testCommentDTD Main func holding tests
func testCommentDTD(t *testing.T, path string, recreate bool) {
	var tests []CommentTestResult
	var dir string

	if recreate {
		dir = "tmp"
	} else {
		dir = "tmp2"
	}

	// New parser
	p := newParser(dir)

	p.Parse(path)
	tests = loadCommentTests("tests/comment.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]

		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name, parsedBlock, test))
		t.Run("Check value", checkStrValue(parsedBlock.GetValue(), test.Value, parsedBlock, test))
	}

	t.Run("Render DTD", render(p))
}
