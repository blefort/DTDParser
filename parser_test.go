package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	DTDParser "github.com/blefort/DTDParser/parser"
)

const dirTest = "tests/"

var overwrite bool

// TestMain Test Initialization
func TestMain(m *testing.M) {

	verbosity := flag.String("verbose", "v", "Verbose v, vv or trace")
	overwriteF := flag.Bool("overwrite", false, "Overwrite output file")

	flag.Parse()

	if *verbosity == "v" {
		log.SetLevel(logrus.InfoLevel)
	}

	if *verbosity == "vv" {
		log.SetLevel(logrus.DebugLevel)
	}

	if *verbosity == "trace" {
		log.SetLevel(logrus.TraceLevel)
	}

	if *overwriteF {
		overwrite = true
	}

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

//  newParser() Instantiate parser and configure it
func newParser() *DTDParser.Parser {
	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.SetOutputPath("tmp")

	if overwrite {
		p.Overwrite = overwrite
	}
	return p
}

// Render
func render(p *DTDParser.Parser) func(*testing.T) {
	return func(t *testing.T) {
		p.RenderDTD("")
	}
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
