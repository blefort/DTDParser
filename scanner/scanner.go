// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package scanner

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/blefort/DTDParser/DTD"
)

// DTDScanner represents a DTD scanner
type DTDScanner struct {
	Data         *bufio.Scanner
	WithComments bool
	Filepath     string
	CurrentLine  int
	CurrentChar  int
	CharCount    int
	LineCount    int
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
	fmt.Printf("%t", sc.scanResult)
	return sc.scanResult
}

// Next Move to the next character
func (sc *DTDScanner) next() bool {
	sc.CharCount++
	sc.scanResult = sc.Data.Scan()
	if sc.isEndOfLine() {
		sc.LineCount++
		sc.Log.Debugf("Line is %d", sc.LineCount)
	}
	return sc.scanResult
}

// Next Move to the next character
func (sc *DTDScanner) Previous() bool {
	sc.CharCount--
	ret := sc.Data.Scan()
	if sc.isEndOfLine() {
		sc.LineCount++
		sc.Log.Debugf("Line is %d", sc.LineCount)
	}
	return ret
}

// Scan the string to find the next block
func (sc *DTDScanner) Scan() (DTD.IDTDBlock, []*word, error) {

	fmt.Sprintf("Scan starts")
	t1 := time.Now().UnixMilli()

	//sc.Log.Debug("Seeking for next block")
	fmt.Sprintf("seek")
	s := sc.seekUntilNextBlock()

	//sc.Log.Warnf("block", zap.Object("Sentence", s))
	fmt.Sprintf("seek %v", s)

	sc.findDTDBlockType(s)

	if s.DTDType == DTD.UNIDENTIFIED {
		return nil, s.getWords(false), errors.New("Unidentified block")
	}

	if s.DTDType == DTD.COMMENT {
		comment := sc.ParseComment(s)
		t2 := time.Now().UnixMilli()
		diff := t2 - t1
		sc.Log.Infof("Commment found at line '%d' in '%d'", sc.CurrentLine, diff)
		return comment, s.getWords(false), nil
	}

	if s.DTDType == DTD.ENTITY {
		entity := sc.ParseEntity(s)
		t2 := time.Now().UnixMilli()
		diff := t2 - t1
		sc.Log.Infof("ENTITY '%s' (line %d) in '%d'", entity.GetName(), sc.CurrentLine, diff)
		sc.Log.Warnf("Entity name", entity.Name)
		sc.Log.Warnf("Entity value", entity.Value)
		sc.Log.Warnf("Entity external", entity.IsExternal)
		sc.Log.Warnf("Entity public", entity.Public)
		sc.Log.Warnf("Entity system", entity.System)
		sc.Log.Warnf("Entity url", entity.Url)
		sc.Log.Warnf("Will render as", entity.Render())
		return entity, s.getWords(false), nil
	}

	if s.DTDType == DTD.ELEMENT {
		element := sc.ParseElement(s)
		t2 := time.Now().UnixMilli()
		diff := t2 - t1
		sc.Log.Infof("ELEMENT '%s' (line %d) in '%d'", element.GetName(), sc.CurrentLine, diff)
		return element, s.getWords(false), nil
	}

	if s.DTDType == DTD.NOTATION {
		notation := sc.ParseNotation(s)
		t2 := time.Now().UnixMilli()
		diff := t2 - t1
		sc.Log.Infof("NOTATION '%s' (line %d) in '%d'", notation.GetName(), sc.CurrentLine, diff)
		return notation, s.getWords(false), nil
	}

	if s.DTDType == DTD.ATTLIST {
		attlist := sc.ParseAttlist(s)
		t2 := time.Now().UnixMilli()
		diff := t2 - t1
		sc.Log.Infof("ATTLIST '%s' (line %d) in '%d'", attlist.GetName(), sc.CurrentLine, diff)
		sc.logOutputAttributes(&attlist.Attributes)
		return attlist, s.getWords(false), nil
	}

	return nil, s.getWords(false), errors.New("Unidentified block")

}

// ParseComment Parse a string and return pointer to DTD.Comment
func (sc *DTDScanner) ParseComment(s *sentence) *DTD.Comment {
	var c DTD.Comment

	c.Value = s.readSequence()
	return &c
}

// ParseEntity Parse a string and return pointer to a DTD.Element
// @ref https://www.w3.org/TR/xml11/#elemdecls
//
// Element Declaration
// [45]   	elementdecl	   ::=   	'<!ELEMENT' S Name S contentspec S? '>'	[VC: Unique Element Type Declaration]
// [46]   	contentspec	   ::=   	'EMPTY' | 'ANY' | Mixed | children
//
func (sc *DTDScanner) ParseElement(s *sentence) *DTD.Element {
	var e DTD.Element

	words := s.getWords(true)

	// for i, w := range words {
	// 	sc.Log.Warnf("-" + fmt.Sprintf("%d", i) + " [" + w.Read() + "] ")
	// }

	if len(words) < 2 {
		sc.Log.Fatalf("Not enough arguments in sentence '", s.sequence, "' (count was", len(words), ")")
	}
	e.Name = words[1].Read()
	e.Value = words[2].Read()
	return &e
}

// ParseNotation Parse a string and return pointer to a DTD.Notation
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

	for i, w := range words {
		sc.Log.Warnf("-" + fmt.Sprintf("%d", i) + " [" + w.Read() + "] ")

		if w.Read() == "PUBLIC" {
			n.Public = true
		}

		if w.Read() == "SYSTEM" {
			n.System = true
		}

	}

	n.Name = words[1].Read()

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

// ParseEntity Parse a string and return pointer to a DTD.Entity
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

	for i, w := range words {

		sc.Log.Warnf("-" + fmt.Sprintf("%d", i) + " [" + w.Read() + "] ")

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

// ParseAttlist Parse a string and return pointer to a DTD.Attlist
//
// [52]   	AttlistDecl	   ::=   	'<!ATTLIST' S Name AttDef* S? '>'
// [53]   	AttDef	   ::=   	S Name S AttType S DefaultDecl
//
func (sc *DTDScanner) ParseAttlist(s *sentence) *DTD.Attlist {
	var attlist DTD.Attlist
	words := s.getWords(true)
	attlist.Name = words[1].Read()
	sc.parseAttributes(words[2:len(words)], &attlist.Attributes)
	return &attlist
}

func (sc *DTDScanner) parseAttributes(words []*word, attributes *[]DTD.Attribute) {

	i := -1
	l := len(words)

	for i, w := range words {
		sc.Log.Warnf("- " + fmt.Sprintf("%d", i) + " [" + w.Read() + "] ")
	}

	// for idx, part := range parts {
	// 	sc.Log.Debugf("part %d: %s", idx, part)
	// }

	if l == 0 {
		//	panic("Unable to scan Attlist")
	}

	nextWord := func(i *int, l int) bool {
		if *i+1 < l {
			*i++
			sc.Log.Debugf("Next word, i: %d, %s ", *i, words[*i])
			return true
		}
		return false
	}

	for nextWord(&i, l) {

		var attr DTD.Attribute

		// instruction varies depending the type, we don't know in advance

		// reference to an entity
		if words[i].Read()[0:1] == "%" {
			sc.Log.Warnf("- Ref. to an entity found: %s", words[i].Read())
			attr.Value = words[i].Read()
			attr.IsEntity = true
			*attributes = append(*attributes, attr)
			continue
		}

		// // first word is always the attribute name
		attr.Name = words[i].Read()
		sc.Log.Warnf("Processing attribute: '%s'", attr.Name)

		if !nextWord(&i, l) {
			sc.Log.Fatalf("Not enough arguments to loop through attributes i:%d", i)
		}

		// CASE 2
		// The others words are processed by group of 3 to 4 depending the type

		// // Type is always in the second position
		attr.Type = DTD.SeekAttributeType(words[i].Read())
		sc.Log.Warnf("attribute type is %d", attr.Type)

		if attr.Type == DTD.CDATA { // 20
			nextWord(&i, l)
			sc.checkDefaultValue(words, &i, &attr)
		} else if attr.Type == DTD.TOKEN_ID ||
			attr.Type == DTD.TOKEN_IDREF ||
			attr.Type == DTD.TOKEN_IDREFS ||
			attr.Type == DTD.TOKEN_ENTITY ||
			attr.Type == DTD.TOKEN_ENTITIES ||
			attr.Type == DTD.TOKEN_NMTOKEN ||
			attr.Type == DTD.TOKEN_NMTOKENS {
			nextWord(&i, l)
			sc.checkDefaultValue(words, &i, &attr)
			sc.checkDefaultValue(words, &i, &attr)
		} else if attr.Type == DTD.ENUM_NOTATION {
			nextWord(&i, l)
			sc.checkDefaultValue(words, &i, &attr)
			sc.checkDefaultValue(words, &i, &attr)
		} else if attr.Type == DTD.ENUM_ENUM {

			sc.checkDefaultValue(words, &i, &attr)
			sc.checkDefaultValue(words, &i, &attr)
		} else {
			sc.Log.Fatalf("unmanaged attribute type %d", attr.Type)
		}

		*attributes = append(*attributes, attr)
		// sc.Log.Debugf("*Attr rendered: %s", attr.Render())

		// if attr.Type == 0 {
		// 	sc.Log.Fatalf("Could not identify attribute type at line %d, name: '%s', value: '%s'", sc.LineCount, attr.Name, words[i].Read())
		// }

	}
}

// heckDefaultValue
func (sc *DTDScanner) checkDefaultValue(w []*word, i *int, attr *DTD.Attribute) {

	ini := *i

	if *i > len(w)-1 {
		return
	}

	sc.checkRequired(w, i, attr)
	sc.checkImplied(w, i, attr)
	sc.checkFixed(w, i, attr)

	if *i == ini {
		attr.Value = w[*i].Read()
		*i++
	}

}

// checkRequired
func (sc *DTDScanner) checkRequired(w []*word, i *int, attr *DTD.Attribute) {
	if *i > len(w)-1 {
		return
	}
	if w[*i].Read() == "#REQUIRED" {
		*i++
		attr.Required = true
	}
}

// checkImplied
func (sc *DTDScanner) checkImplied(w []*word, i *int, attr *DTD.Attribute) {
	if *i > len(w)-1 {
		return
	}
	if w[*i].Read() == "#IMPLIED" {
		*i++
		attr.Implied = true
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
	}
}

// output attributes in the log
func (sc *DTDScanner) logOutputAttributes(attributes *[]DTD.Attribute) {
	for _, attr := range *attributes {
		sc.Log.Debugf("- Attribute: %s", attr.Render())
	}
}

// assignIfEntityValue test if v is not public system or empty before assigning it
func assignIfEntityValue(e *DTD.Entity, v string) {
	//if v != "" && !isPublic(v) && !isSystem(v) {
	//	e.Value = v
	//}
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
	sc.Log.Warnf("start scanning sentence with char '" + sc.Data.Text() + "'")

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

func (sc *DTDScanner) findDTDBlockType(s *sentence) {

	words := s.getWords(true)

	if len(words) == 0 {
		return
	}

	w := words[0].Read()

	if len(w) > 3 && w[0:4] == "<!--" {
		s.DTDType = DTD.COMMENT
	} else if w == "<!ATTLIST" {
		s.DTDType = DTD.ATTLIST
	} else if w == "<!ELEMENT" {
		s.DTDType = DTD.ELEMENT
	} else if w == "<!NOTATION" {
		s.DTDType = DTD.NOTATION
	} else if w == "<!ENTITY" {
		s.DTDType = DTD.ENTITY
	}
	sc.Log.Warnf(fmt.Sprintf("%d", s.DTDType))
}
