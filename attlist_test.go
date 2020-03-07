package main

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/blefort/DTDParser/DTD"
)

// AttrTestResult struct to attributes
// @todo replace by dtd.Attlist
type AttrTestResult struct {
	Name       string
	Type       int
	Attributes []DTD.Attribute
}

// loadAttlistTests Load attribute tests
func loadAttlistTests(file string) []AttrTestResult {
	var tests []AttrTestResult
	loadJSON(file, &tests)
	return tests
}

// TestParseAttlistBlock Test for the attlist parser
func TestParseAttlistBlock(t *testing.T) {
	// - parse the DTD test
	// - compare it to data stored in a json file
	// - render it in the tmp dir
	testAttlistDTD(t, "tests/attlist.dtd")

	// - load the generated DTD
	// - compare it to data stored in a json file
	testAttlistDTD(t, "tmp/attlist.dtd")
}

// testAttlistDTD main testing func for attlist
func testAttlistDTD(t *testing.T, path string) {
	var tests []AttrTestResult

	// New parser
	p := newParser()

	p.Parse(path)
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
	}
	t.Run("Render DTD", render(p))
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

// AttlistGetUrl() Helper to test DTD.Attlist.GetUrl()
func AttlistGetUrl() {
	var a DTD.Attlist
	ret := a.GetUrl()
	log.Tracef("AttlistUrl( return %s", ret)
}
