package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/blefort/DTDParser/DTD"
	DTDParser "github.com/blefort/DTDParser/parser"
)

const dirTest = "tests/"

type CommentTestResult struct {
	Name  string
	Value string
}

type EntityTestResult struct {
	Name      string
	Value     string
	Parameter bool
	Url       string
}

type AttrTestResult struct {
	Name       string
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

// Load tests
func loadCommentTests(file string) []CommentTestResult {
	// Open file
	var tests []CommentTestResult
	loadJSON(file, &tests)
	return tests
}

// Load tests
func loadEntityTests(file string) []EntityTestResult {
	// Open file
	var tests []EntityTestResult
	loadJSON(file, &tests)
	return tests
}

// Load tests
func loadAttlistTests(file string) []AttrTestResult {
	// Open file
	var tests []AttrTestResult
	loadJSON(file, &tests)
	return tests
}

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
	log.Warnf("tests %+v", tests)
	log.Warnf("collection %+v", p.Collection)

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]

		t.Run("Check DTD Type", func(t *testing.T) {
			if !DTD.IsCommentType(parsedBlock) {
				t.Errorf("Received wrong value, '%s' instead of 'comment'", parsedBlock)
			}
		})
		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name))
		t.Run("Check value", checkStrValue(parsedBlock.GetValue(), test.Value))
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
	log.Warnf("tests %+v", tests)
	log.Warnf("collection %+v", p.Collection)

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]

		t.Run("Check DTD Type", func(t *testing.T) {
			if !DTD.IsEntityType(parsedBlock) {
				t.Errorf("Received wrong value, '%s' instead of 'entity'", parsedBlock)
			}
		})
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
	log.Warnf("tests %+v", tests)
	log.Warnf("collection %+v", p.Collection)

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]

		t.Run("Check DTD Type", func(t *testing.T) {
			if !DTD.IsAttlistType(parsedBlock) {
				t.Errorf("Received wrong value, '%s' instead of Attlist", parsedBlock)
			}
		})
		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name))
		t.Run("Render", render(p))

	}
}

// checkType Check if the block found from the parser has the expected type
func checkType(parsed DTD.IDTDBlock, expected DTD.IDTDBlock) func(*testing.T) {
	return func(t *testing.T) {
		if reflect.TypeOf(parsed) != reflect.TypeOf(expected) {
			t.Error("Received wrong type")
		}
	}
}

// checkStrValue Check if the block found from the parser has the expected value
func checkStrValue(parsed string, expected string) func(*testing.T) {
	return func(t *testing.T) {
		if parsed != expected {
			t.Errorf("Received wrong value, '%s' instead of '%s'", parsed, expected)
		}
	}
}

// checkBoolValue Check if the block found from the parser has the expected value
func checkBoolValue(a bool, b bool) func(*testing.T) {
	return func(t *testing.T) {
		if a != b {
			t.Errorf("Received wrong bool value, '%t' instead of '%t'", a, b)
		}
	}
}

// Render
func render(p *DTDParser.Parser) func(*testing.T) {
	return func(t *testing.T) {
		p.Render("")
	}
}
