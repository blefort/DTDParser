package main

import (
	DTDParser "github.com/blefort/DTDParser/parser"
	"testing"
)

func TestParseComment(t *testing.T) {
	// New parser
	p := DTDParser.NewDTDParser()

	// Configure parser
	p.WithComments = true
	p.IgnoreExtRefIssue = true
	p.SetVerbose(DTDParser.LOG_FULL)
	p.Parse("tests/test.dtd")
}
