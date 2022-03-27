package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	DTDParser "github.com/blefort/DTDParser/parser"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const dirTest = "tests/"

var overwrite bool
var log *zap.SugaredLogger

// TestMain Test Initialization
func TestMain(m *testing.M) {

	//verbosity := flag.String("verbose", "v", "Verbose v, vv or trace")
	overwriteF := flag.Bool("overwrite", false, "Overwrite output file")

	flag.Parse()

	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "m",
			LevelKey:       "",
			TimeKey:        "",
			NameKey:        "",
			CallerKey:      "",
			StacktraceKey:  "stack",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}

	cfg.EncoderConfig.TimeKey = zapcore.OmitKey

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // flushes buffer, if any

	log = logger.Sugar()

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
func newParser(dir string) *DTDParser.Parser {

	fmt.Sprintf("new parser %p", log)
	// New parser
	p := DTDParser.NewDTDParser(log)

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.SetOutputPath(dir)

	if overwrite {
		p.Overwrite = overwrite
	}
	return p
}

// Render
func render(p *DTDParser.Parser) func(*testing.T) {
	return func(t *testing.T) {
		fmt.Printf("pointer: %p", log)
		p.RenderDTD("")
	}
}

/**
 * Helpers
 */

// checkStrValue Check if the block found from the parser has the expected value
func checkStrValue(a string, b string, i interface{}, f interface{}) func(*testing.T) {
	return func(t *testing.T) {
		log.Debugf("Received string value, '%s' to be compared to expected value '%s'", a, b)
		if a != b {
			ac := strings.ReplaceAll(a, "\n", "\\n")
			bc := strings.ReplaceAll(b, "\n", "\\n")
			t.Errorf("Received wrong value, '%s' instead of '%s' - %+v - %+v", ac, bc, i, f)
		}
	}
}

// checkBoolValue Check if the block found from the parser has the expected value
func checkBoolValue(a bool, b bool, i interface{}, f interface{}) func(*testing.T) {
	return func(t *testing.T) {
		log.Debugf("Received bool value, '%t' to be compared to expected value '%t'", a, b)
		if a != b {
			t.Errorf("Received wrong bool value, '%t' instead of '%t' - %+v - %+v", a, b, i, f)
		}
	}
}

// checkIntValue Check if the block found from the parser has the expected value
func checkIntValue(a int, b int, i interface{}) func(*testing.T) {
	return func(t *testing.T) {
		log.Debugf("Received int value, '%d' to be compared to expected value '%d'", a, b)
		if a != b {
			t.Errorf("Received wrong int value, '%d' instead of '%d'- %+v", a, b, i)
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
