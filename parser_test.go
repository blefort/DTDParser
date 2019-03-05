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

type testResult struct {
	Ntype     int
	Name      string
	Value     string
	Parameter bool
	Url       string
}

// TestMain Test Initialization
func TestMain(m *testing.M) {
	log.SetLevel(log.TraceLevel)
	os.Exit(m.Run())
}

// Load tests
func loadTests(file string) []testResult {
	// Open file
	var tests []testResult

	filebuffer, err := ioutil.ReadFile(file)

	if err != nil {
		panic(err)
	}

	// parse file
	errj := json.Unmarshal(filebuffer, &tests)
	if errj != nil {
		panic(errj)
	}
	return tests
}

func TestParseDTDBlock(t *testing.T) {
	var tests []testResult

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.Parse("tests/test.dtd")
	p.SetOutputPath("tmp")

	tests = loadTests("tests/test.json")
	log.Warnf("tests %+v", tests)
	log.Warnf("collection %+v", p.Collection)

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		parsedBlock := p.Collection[idx]
		expectedBlock := getType(test)

		log.Tracef("parsedBlock is: %+v and expectedBlock is: %+v", parsedBlock, expectedBlock)

		t.Run("Check DTD Type", checkType(parsedBlock, expectedBlock))
		t.Run("Check name", checkStrValue(parsedBlock.GetName(), test.Name))
		t.Run("Check value", checkStrValue(parsedBlock.GetValue(), test.Value))

		if DTD.IsEntityType(expectedBlock) {
			t.Run("Check Parameter", checkBoolValue(parsedBlock.GetParameter(), test.Parameter))
			t.Run("Check Url", checkStrValue(parsedBlock.GetUrl(), test.Url))
		}

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
			t.Errorf("Received wrong value, '%s' instead if '%s'", parsed, expected)
		}
	}
}

// checkBoolValue Check if the block found from the parser has the expected value
func checkBoolValue(a bool, b bool) func(*testing.T) {
	return func(t *testing.T) {
		if a != b {
			t.Errorf("Received wrong bool value, '%t' instead if '%t'", a, b)
		}
	}
}

// Render
func render(p *DTDParser.Parser) func(*testing.T) {
	return func(t *testing.T) {
		p.Render("")
	}
}

// getType instantiate a type from its representation
func getType(test testResult) DTD.IDTDBlock {
	switch test.Ntype {
	case DTD.COMMENT:
		var comment DTD.Comment
		return &comment
	case DTD.ENTITY:
		var entity DTD.Entity
		return &entity
	default:
		log.Panicf("type '%d' not found", test.Ntype)
	}
	return nil
}
