// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package scanner

import (
	"bufio"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/blefort/DTDParser/DTD"
)

// DTDScanner represents a DTD scanner
type DTDScanner struct {
	Data         *bufio.Scanner
	WithComments bool
	Filepath     string
	CurrentLine  int // first line of a block
	CurrentChar  int
	CharCount    int
	LineCount    int  // line processed by the scanner
	CRLF         bool // 0 linux / 1 dos
	init         bool
	scanResult   bool
	Log          *zap.SugaredLogger
}

// NewScanner returns a new DTD Scanner
func NewScanner(path string, s string, log *zap.SugaredLogger) *DTDScanner {
	var scanner DTDScanner
	scanner.Data = bufio.NewScanner(strings.NewReader(s))
	scanner.Data.Split(bufio.ScanRunes)
	scanner.Filepath = path
	scanner.CharCount = 0
	scanner.LineCount = 1
	scanner.CRLF = false // assume Linux
	scanner.init = false
	scanner.scanResult = true
	scanner.Log = log
	return &scanner
}

// Next Move to the next Block
func (sc *DTDScanner) NextBlock() bool {
	return sc.scanResult
}

// Next Move to the next character
func (sc *DTDScanner) next() bool {
	sc.CharCount++
	sc.scanResult = sc.Data.Scan()
	if sc.isEndOfLine() {
		sc.LineCount++
	}
	return sc.scanResult
}

// Next Move to the next character
func (sc *DTDScanner) Previous() bool {
	sc.CharCount--
	ret := sc.Data.Scan()
	if sc.isEndOfLine() {
		sc.LineCount++
	}
	return ret
}

// Scan the string to find the next block
func (sc *DTDScanner) Scan() (DTD.IDTDBlock, []*word, error) {

	// seek until next DTD block
	s := sc.seekUntilNextBlock()

	// determine block type
	sc.findDTDBlockType(s)

	if s.DTDType == DTD.UNIDENTIFIED {
		return nil, s.getWords(false), fmt.Errorf("could not identify DTD block at line %d", sc.LineCount)
	}

	if s.DTDType == DTD.COMMENT {
		comment := sc.ParseComment(s)
		return comment, s.getWords(false), nil
	}

	if s.DTDType == DTD.ENTITY {
		entity := sc.ParseEntity(s)
		return entity, s.getWords(false), nil
	}

	if s.DTDType == DTD.ELEMENT {
		element := sc.ParseElement(s)
		return element, s.getWords(false), nil
	}

	if s.DTDType == DTD.NOTATION {
		notation := sc.ParseNotation(s)
		return notation, s.getWords(false), nil
	}

	if s.DTDType == DTD.ATTLIST {
		attlist := sc.ParseAttlist(s)
		sc.logOutputAttributes(&attlist.Attributes)
		return attlist, s.getWords(false), nil
	}

	return nil, s.getWords(false), fmt.Errorf("could not identify DTD block at line %d", sc.LineCount)

}

// ParseComment Use the information in the sentence to return a pointer to a DTD.Comment
func (sc *DTDScanner) ParseComment(s *sentence) *DTD.Comment {
	var c DTD.Comment
	sc.Log.Info("Comment found line ", sc.CurrentLine)
	c.Value = s.readSequence()
	return &c
}

// ParseEntity Use the information in the sentence to return a pointer to a DTD.Element
// @ref https://www.w3.org/TR/xml11/#elemdecls
//
// Element Declaration
// [45]   	elementdecl	   ::=   	'<!ELEMENT' S Name S contentspec S? '>'	[VC: Unique Element Type Declaration]
// [46]   	contentspec	   ::=   	'EMPTY' | 'ANY' | Mixed | children
//
func (sc *DTDScanner) ParseElement(s *sentence) *DTD.Element {
	var e DTD.Element

	words := s.getWords(true)
	l := len(words)

	if len(words) < 2 {
		sc.Log.Errorf("not enough arguments in sentence '%s'. Count was (%d)", s.sequence, len(words))
	}
	e.Name = words[1].Read()

	for i := 2; i < l; i++ {
		e.Value = e.Value + " " + words[i].Read()
	}

	sc.Log.Info("ParseElement ", e.Name)
	return &e
}

// ParseNotation Use the information in the sentence to return a pointer to a DTD.Notation
// @ref https://www.w3.org/TR/xml11/#Notations
//
// Element Declaration
//
// [82]  NotationDec ::= '<!NOTATION' S Name S (ExternalID | PublicID) S? '>'  [VC: Unique Notation Name]
// [83]  PublicID    ::= 'PUBLIC' S PubidLiteral
//
func (sc *DTDScanner) ParseNotation(s *sentence) *DTD.Notation {
	var n DTD.Notation

	words := s.getWords(true)

	l := len(words)

	for _, w := range words {

		if w.Read() == "PUBLIC" {
			n.Public = true
		}

		if w.Read() == "SYSTEM" {
			n.System = true
		}

	}

	n.Name = words[1].Read()
	sc.Log.Info("ParseNotation ", n.Name)

	if l > 3 && n.Public {
		n.PublicID = words[3].Read()
	}

	if l > 3 && n.System {
		n.SystemID = words[3].Read()
	}

	if l > 4 {
		n.SystemID = words[4].Read()
	}

	return &n
}

// ParseEntity Use the information in the sentence to return a pointer to a DTD.Entity
// @ref https://www.w3.org/TR/xml11/#sec-entity-decl
//
// Entity Declaration
// [70]   	EntityDecl ::=   	GEDecl | PEDecl
// [71]   	GEDecl     ::=   	'<!ENTITY' S Name S EntityDef S? '>'
// [72]   	PEDecl     ::=   	'<!ENTITY' S '%' S Name S PEDef S? '>'
// [73]   	EntityDef  ::=   	EntityValue | (ExternalID NDataDecl?)
// [74]   	PEDef	   ::=   	EntityValue | ExternalID
//
func (sc *DTDScanner) ParseEntity(s *sentence) *DTD.Entity {
	var e DTD.Entity

	words := s.getWords(true)
	idx := 0
	pidx := 1

	sc.logWords(&words)

	for i, w := range words {

		if w.Read() == "%" {
			e.Parameter = true
			pidx = i + 1
		}

		if w.Read() == "PUBLIC" {
			e.Public = true
			e.IsExternal = true
			idx = i
		}

		if w.Read() == "SYSTEM" {
			e.System = true
			e.IsExternal = true
			idx = i
		}
	}

	e.Name = words[pidx].Read()
	sc.Log.Info("ParseEntity ", e.Name)

	if e.System {
		e.Url = words[idx+1].Read()
	} else if e.Public {
		e.Url = words[idx+2].Read()
		e.Value = words[idx+1].Read()
	} else if e.Parameter {
		e.Value = words[3].Read()
	} else {
		e.Value = words[len(words)-1].Read()
	}

	return &e
}

// ParseAttlist Use the information in the sentence to return a pointer to a DTD.Attlist
//
// [52]   	AttlistDecl	   ::=   	'<!ATTLIST' S Name AttDef* S? '>'
// [53]   	AttDef	   ::=   	S Name S AttType S DefaultDecl
//
func (sc *DTDScanner) ParseAttlist(s *sentence) *DTD.Attlist {
	var attlist DTD.Attlist

	words := s.getWords(true)
	sc.logWords(&words)
	attlist.Name = words[1].Read()
	sc.Log.Info("ParseAttlist ", attlist.Name)
	sc.parseAttributes(words[2:len(words)], &attlist.Attributes)
	return &attlist
}

// parseAttributes Use the information in the sentence.words to return a pointer to *[]DTD.Attribute
func (sc *DTDScanner) parseAttributes(words []*word, attributes *[]DTD.Attribute) {

	i := -1
	l := len(words)

	sc.Log.Info("ParseAttributes")

	sc.logWords(&words)

	nextWord := func(i *int, l int) bool {
		if *i+1 < l {
			*i++
			sc.Log.Debugf("i: %d", *i)
			return true
		}
		return false
	}

	for nextWord(&i, l) {

		var attr DTD.Attribute

		// instruction varies depending the type, we don't know in advance
		sc.Log.Debugf("Processing word: %s", words[i].Read())
		// reference to an entity
		if words[i].Read()[0:1] == "%" {
			sc.Log.Debugf("- reference to an entity found: %s", words[i].Read())
			attr.Value = words[i].Read()
			attr.IsEntity = true
			*attributes = append(*attributes, attr)
			continue
		}

		// // first word is always the attribute name
		attr.Name = words[i].Read()
		sc.Log.Debugf("processing attribute: '%s'", attr.Name)

		if !nextWord(&i, l) {
			sc.Log.Fatalf("Not enough arguments to loop through attributes i:%d", i)
		}

		// CASE 2
		// The others words are processed by group of 3 to 4 depending the type

		// // Type is always in the second position
		attr.Type = DTD.SeekAttributeType(words[i].Read())
		sc.Log.Debugf("attribute type is %d", attr.Type)

		if attr.Type == DTD.CDATA { // 20
			nextWord(&i, l)
			sc.checkDefaultValue(words, &i, &attr)
		} else if attr.Type == DTD.TOKEN_ID ||
			attr.Type == DTD.TOKEN_IDREF ||
			attr.Type == DTD.TOKEN_IDREFS ||
			attr.Type == DTD.TOKEN_ENTITY ||
			attr.Type == DTD.TOKEN_ENTITIES ||
			attr.Type == DTD.TOKEN_NMTOKENS {
			nextWord(&i, l)
			sc.checkDefaultValue(words, &i, &attr)
		} else if attr.Type == DTD.TOKEN_NMTOKEN {
			nextWord(&i, l)
			sc.checkDefaultValue(words, &i, &attr)
		} else if attr.Type == DTD.ENUM_NOTATION {
			nextWord(&i, l)
			sc.checkDefaultValue(words, &i, &attr)
		} else if attr.Type == DTD.ENUM_ENUM {
			sc.checkDefaultValue(words, &i, &attr)
		} else {
			sc.Log.Fatalf("unmanaged attribute type %d", attr.Type)
		}

		sc.logAttribute(&attr)
		*attributes = append(*attributes, attr)

	}
}

// heckDefaultValue
func (sc *DTDScanner) checkDefaultValue(w []*word, i *int, attr *DTD.Attribute) {

	if *i > len(w)-1 {
		return
	}

	sc.checkFixed(w, i, attr)

	if attr.Fixed {
		attr.Value = w[*i].Read()
		sc.Log.Debugf("Attribute value is %s", attr.Value)
		return
	}

	// Required and implied appears to be always the last value
	sc.checkRequired(w, i, attr)
	if attr.Required {
		return
	}

	sc.checkImplied(w, i, attr)
	if attr.Implied {
		return
	}

	if !attr.Fixed && !attr.Required && !attr.Implied {
		attr.Value = w[*i].Read()
		sc.Log.Debugf("Attribute value is %s", attr.Value)
		*i++
	}

	if *i < len(w) {
		sc.Log.Debugf("Second round")
		sc.checkDefaultValue(w, i, attr)
	}

}

// checkRequired
func (sc *DTDScanner) checkRequired(w []*word, i *int, attr *DTD.Attribute) {
	if *i > len(w)-1 {
		return
	}
	if w[*i].Read() == "#REQUIRED" {
		attr.Required = true
		sc.Log.Debug("REQUIRED Detected")
	}
}

// checkImplied
func (sc *DTDScanner) checkImplied(w []*word, i *int, attr *DTD.Attribute) {
	if *i > len(w)-1 {
		return
	}
	if w[*i].Read() == "#IMPLIED" {
		attr.Implied = true
		sc.Log.Debug("IMPLIED Detected")
	}
}

// checkFixed
func (sc *DTDScanner) checkFixed(w []*word, i *int, attr *DTD.Attribute) {
	if *i > len(w)-1 {
		return
	}
	if w[*i].Read() == "#FIXED" {
		*i++
		attr.Fixed = true
		sc.Log.Debug("FIXED Detected")
	}
}

// isEndOfLine Identitfy a carriage return
func (sc *DTDScanner) isEndOfLine() bool {

	if sc.Data.Text() == "\n" {
		return true
	}
	if sc.Data.Text() == "\r" {
		sc.CRLF = true
		return false
	}
	return false
}

// seekUntilNextBlock return string until next block is found
func (sc *DTDScanner) seekUntilNextBlock() *sentence {

	//	var contentPosition int
	sentence := newsentence("<", ">", sc.Log)

	sc.CurrentLine = sc.LineCount
	sc.Log.Debugf(fmt.Sprintf("start scanning sentence with char '%s'", sc.Data.Text()))

	sentence.scan(sc.Data.Text())

	for sc.next() {

		if sentence.scan(sc.Data.Text()) {
			sentence.read()
			sc.next()
			return sentence
		}
	}

	return sentence
}

// findDTDBlockType Determine DTD type with the first word of the sentence
func (sc *DTDScanner) findDTDBlockType(s *sentence) {

	words := s.getWords(true)

	if len(words) == 0 {
		return
	}

	if len(words[0].Read()) > 3 && words[0].Read()[0:4] == "<!--" {
		s.DTDType = DTD.COMMENT
	} else if words[0].Read() == "<!ATTLIST" {
		s.DTDType = DTD.ATTLIST
	} else if words[0].Read() == "<!ELEMENT" {
		s.DTDType = DTD.ELEMENT
	} else if words[0].Read() == "<!NOTATION" {
		s.DTDType = DTD.NOTATION
	} else if words[0].Read() == "<!ENTITY" {
		s.DTDType = DTD.ENTITY
	}
	sc.Log.Debugf(fmt.Sprintf("block type is (%d)", s.DTDType))
}

// logOutputAttributes helper function to output attributes in the log
func (sc *DTDScanner) logAttribute(attr *DTD.Attribute) {
	sc.Log.Infof(fmt.Sprintf(" - attribute: '%s'", attr.Render()))
}

// logOutputAttributes helper function to output attributes in the log
func (sc *DTDScanner) logOutputAttributes(attributes *[]DTD.Attribute) {
	for i, attr := range *attributes {
		sc.Log.Debugf(fmt.Sprintf(" - attribute (%d): '%s'", i, attr.Render()))
	}
}

// logWords helper function to log words
func (sc *DTDScanner) logWords(words *[]*word) {
	for i, w := range *words {
		sc.Log.Debugf(fmt.Sprintf(" - word [%d] '%s'", i, w.Read()))
	}
}
