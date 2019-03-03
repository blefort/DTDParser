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
	Ntype int
	Value string
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

func TestParseComment(t *testing.T) {
	var tests []testResult

	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.Parse("tests/test.dtd")

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

		t.Run("Check DTD Type", checkType(parsedBlock, expectedBlock))
		t.Run("Check value", checkValue(parsedBlock.GetValue(), test.Value))

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

// checkValue Check if the block found from the parser has the expected value
func checkValue(parsed string, expected string) func(*testing.T) {
	return func(t *testing.T) {
		if parsed != expected {
			t.Errorf("Received wrong value, '%s' instead if '%s'", parsed, expected)
		}
	}
}

// getType instantiate a type from its representation
func getType(test testResult) DTD.IDTDBlock {
	switch test.Ntype {
	case DTD.COMMENT:
		var comment DTD.Comment
		return &comment
	default:
		log.Panicf("type '%d' not found", test.Ntype)
	}
	return nil
}
