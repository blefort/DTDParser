package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/blefort/DTDParser/DTD"
	DTDParser "github.com/blefort/DTDParser/parser"
)

const dirTest = "tests/"

// CommentTestResult struct to test comment
type CommentTestResult struct {
	Name  string
	Value string
}

// EntityTestResult struct to test entity
type EntityTestResult struct {
	Name      string
	Value     string
	Parameter bool
	Url       string
}

// AttrTestResult struct to attributes
type AttrTestResult struct {
	Name       string
	Type       int
	Attributes []DTD.Attribute
}

// TestMain Test Initialization
func TestMain(m *testing.M) {
	//log.SetLevel(log.TraceLevel)
	os.Exit(m.Run())
}

// loadJSON load a json file and return the result in v
func loadJSON(file string, v interface{}) {
	filebuffer, err := ioutil.ReadFile(file)

	if err != nil {
		panic(err)
	}

	// parse file
	errj := json.Unmarshal(filebuffer, v)
	if errj != nil {
		panic(errj)
	}
}

// loadCommentTests Load comment tests
func loadCommentTests(file string) []CommentTestResult {
	var tests []CommentTestResult
	loadJSON(file, &tests)
	return tests
}

// loadEntityTests Load entity tests
func loadEntityTests(file string) []EntityTestResult {
	var tests []EntityTestResult
	loadJSON(file, &tests)
	return tests
}

// loadElementTests Load element tests
func loadElementTests(file string) []DTD.Element {
	var tests []DTD.Element
	loadJSON(file, &tests)
	return tests
}

// loadElementTests Load element tests
func loadNotationTests(file string) []DTD.Notation {
	var tests []DTD.Notation
	loadJSON(file, &tests)
	return tests
}

// loadAttlistTests Load attribute tests
func loadAttlistTests(file string) []AttrTestResult {
	var tests []AttrTestResult
	loadJSON(file, &tests)
	return tests
}

// TestParseCommentBlock Test parser for result
func TestParseCommentBlock(t *testing.T) {
	var tests []CommentTestResult

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.Parse("tests/comment.dtd")
	p.SetOutputPath("tmp")

	tests = loadCommentTests("tests/comment.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]

		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name))
		t.Run("Check value", checkStrValue(parsedBlock.GetValue(), test.Value))
		t.Run("Render", render(p))

	}
}

// TestParseCommentBlock Test parser for result
func TestParseElementBlock(t *testing.T) {
	var tests []DTD.Element

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.Parse("tests/element.dtd")
	p.SetOutputPath("tmp")

	tests = loadElementTests("tests/element.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]

		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name))
		t.Run("Check value", checkStrValue(parsedBlock.GetValue(), test.Value))
		t.Run("Render", render(p))

	}
}

// TestParseCommentBlock Test parser for result
func TestParseNotationBlock(t *testing.T) {
	var tests []DTD.Notation

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.Parse("tests/notation.dtd")
	p.SetOutputPath("tmp")

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
		t.Run("Check System", checkBoolValue(parsedBlock.Public, test.Public))
		t.Run("Render", render(p))

	}
}

func TestParseEntityBlock(t *testing.T) {
	var tests []EntityTestResult

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.Parse("tests/entity.dtd")
	p.SetOutputPath("tmp")

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
		t.Run("Render", render(p))

	}
}

func TestParseAttlistBlock(t *testing.T) {
	var tests []AttrTestResult

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.Parse("tests/attlist.dtd")
	p.SetOutputPath("tmp")

	tests = loadAttlistTests("tests/attlist.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		AttlistBlock := p.Collection[idx].(*DTD.Attlist)

		log.Tracef("Attlist: test: %#v", test)
		log.Tracef("Attlist: parsed: %#v", AttlistBlock)

		if len(AttlistBlock.Attributes) == 0 {
			t.Errorf("Attlist: Not attribute definition found in '%#v'", AttlistBlock)
		}

		t.Run("Attlist: Check Attributes Count", checkIntValue(len(AttlistBlock.Attributes), len(test.Attributes)))

		for attrID, attr := range AttlistBlock.Attributes {

			attrTest := test.Attributes[attrID]
			t.Log("Nexts 2 lines shows #1 expected, #2 found")
			t.Logf("%#v", attrTest)
			t.Logf("%#v", attr)

			t.Run("Attlist:Attribute:Check name", checkStrValue(attr.Name, attrTest.Name))
			t.Run("Attlist:Attribute:Check default value", checkStrValue(attr.Default, attrTest.Default))
			t.Run("Attlist:Attribute:Check #REQUIRED", checkBoolValue(attr.Required, attrTest.Required))
			t.Run("Attlist:Attribute:Check #IMPLIED", checkBoolValue(attr.Implied, attrTest.Implied))
			t.Run("Attlist:Attribute:Check #FIXED", checkBoolValue(attr.Fixed, attrTest.Fixed))
		}
		t.Run("Attlist: Check name", checkStrValue(AttlistBlock.GetName(), test.Name))
		t.Run("Render", render(p))

	}
}

// Render
func render(p *DTDParser.Parser) func(*testing.T) {
	return func(t *testing.T) {
		p.Render("")
	}
}

/**
 * Below are tests for func that should never be called
 */

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

// CommentExported() Helper to test DTD.comment.GetUrl()
func CommentGetUrl() {
	var c DTD.Comment
	ret := c.GetUrl()
	log.Tracef("CommentUrl( return %s", ret)
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

// TestAttlistPanic Test func that should never be called
func TestAttlistPanic(t *testing.T) {
	assertPanic(t, AttlistExported)
	assertPanic(t, AttlistGetParameter)
	assertPanic(t, AttlistGetUrl)
}

// AttlistExported() Helper to test DTD.Attlist.GetExported()
func AttlistExported() {
	var a DTD.Attlist
	ret := a.GetExported()
	log.Tracef("AttlistExported( return %t", ret)
}

// AttlistExported() Helper to test DTD.Attlist.GetParameter()
func AttlistGetParameter() {
	var a DTD.Attlist
	ret := a.GetParameter()
	log.Tracef("AttlistExported( return %t", ret)
}

// AttlistExported() Helper to test DTD.Attlist.GetUrl()
func AttlistGetUrl() {
	var a DTD.Attlist
	ret := a.GetUrl()
	log.Tracef("AttlistUrl( return %s", ret)
}

/**
 * Helpers
 */

// checkStrValue Check if the block found from the parser has the expected value
func checkStrValue(a string, b string) func(*testing.T) {
	return func(t *testing.T) {
		log.Tracef("Received string value, '%s' to be compared to expected value '%s'", a, b)
		if a != b {
			t.Errorf("Received wrong value, '%s' instead of '%s'", a, b)
		}
	}
}

// checkBoolValue Check if the block found from the parser has the expected value
func checkBoolValue(a bool, b bool) func(*testing.T) {
	return func(t *testing.T) {
		log.Tracef("Received bool value, '%t' to be compared to expected value '%t'", a, b)
		if a != b {
			t.Errorf("Received wrong bool value, '%t' instead of '%t'", a, b)
		}
	}
}

// checkIntValue Check if the block found from the parser has the expected value
func checkIntValue(a int, b int) func(*testing.T) {
	return func(t *testing.T) {
		log.Tracef("Received int value, '%d' to be compared to expected value '%d'", a, b)
		if a != b {
			t.Errorf("Received wrong int value, '%d' instead of '%d'", a, b)
		}
	}
}

// assertPanic Helper to test panic
// @see https://stackoverflow.com/questions/31595791/how-to-test-panics
func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}
