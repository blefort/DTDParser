package main

import (
	"testing"

	"github.com/blefort/DTDParser/DTD"
	log "github.com/sirupsen/logrus"
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
	testAttlistDTD(t, "tests/attlist.dtd", true)

	// - load the generated DTD
	// - compare it to data stored in a json file
	testAttlistDTD(t, "tmp/attlist.dtd", false)
}

// testAttlistDTD main testing func for attlist
func testAttlistDTD(t *testing.T, path string, recreate bool) {
	var tests []AttrTestResult
	var dir string

	if recreate {
		dir = "tmp"
	} else {
		dir = "tmp2"
	}

	// New parser
	p := newParser(dir)

	p.Parse(path)
	tests = loadAttlistTests("tests/attlist.json")

	if len(p.Collection) != len(tests) {
		t.Errorf("Number of elements in the collection (%d) differs from number of tests (%d), please update either your DTD test or the corresponding json file", len(p.Collection), len(tests))
		t.SkipNow()
	}

	for idx, test := range tests {

		AttlistBlock := p.Collection[idx].(*DTD.Attlist)

		//log.Tracef("Attlist: test: %#v", test)
		//log.Tracef("Attlist: parsed: %#v", AttlistBlock)

		if len(AttlistBlock.Attributes) == 0 {
			t.Errorf("Attlist: Not attribute definition found in '%#v'", AttlistBlock)
		}

		t.Run("Attlist: Check Attributes Count", checkIntValue(len(AttlistBlock.Attributes), len(test.Attributes), test))

		for attrID, attr := range AttlistBlock.Attributes {

			attrTest := test.Attributes[attrID]

			t.Run("Attlist:Attribute:Check name", checkStrValue(attr.Name, attrTest.Name, attr))
			t.Run("Attlist:Attribute:Check default value", checkStrValue(attr.Default, attrTest.Default, attr))
			t.Run("Attlist:Attribute:Check #REQUIRED", checkBoolValue(attr.Required, attrTest.Required, attr))
			t.Run("Attlist:Attribute:Check #IMPLIED", checkBoolValue(attr.Implied, attrTest.Implied, attr))
			t.Run("Attlist:Attribute:Check #FIXED", checkBoolValue(attr.Fixed, attrTest.Fixed, attr))

			for idx, entity := range attr.Entities {
				t.Run("Attlist:Entity #"+string(idx), checkStrValue(entity, attrTest.Entities[idx], attr))
			}

		}
		t.Run("Attlist: Check name", checkStrValue(AttlistBlock.GetName(), test.Name, test))
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
