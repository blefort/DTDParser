package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/blefort/DTDParser/DTD"
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
	testCommentDTD(t, "tests/comment.dtd")

	// - load the generated DTD
	// - compare it to data stored in a json file
	t.Log("Start tests on 'tmp/comment.dtd'")
	testCommentDTD(t, "tmp/comment.dtd")
}

// testCommentDTD Main func holding tests
func testCommentDTD(t *testing.T, path string) {
	var tests []CommentTestResult

	// New parser
	p := newParser()

	p.Parse(path)
	tests = loadCommentTests("tests/comment.json")

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

// TestCommentPanic Test func that should never be called
func TestCommentPanic(t *testing.T) {
	assertPanic(t, CommentExported)
	assertPanic(t, CommentGetParameter)
	assertPanic(t, CommentGetUrl)
}

// CommentExported() Helper to test DTD.comment.GetExported()
func CommentExported() {
	var c DTD.Comment
	ret := c.GetExported()
	log.Tracef("CommentExported( return %t", ret)
}

// CommentExported() Helper to test DTD.comment.GetParameter()
func CommentGetParameter() {
	var c DTD.Comment
	ret := c.GetParameter()
	log.Tracef("CommentExported( return %t", ret)
}

//  CommentGetUrl() Helper to test DTD.comment.GetUrl()
func CommentGetUrl() {
	var c DTD.Comment
	ret := c.GetUrl()
	log.Tracef("CommentUrl( return %s", ret)
}
