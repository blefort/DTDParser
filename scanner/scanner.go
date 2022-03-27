// Copyright 2019 Bertrand Lefort. All rights reserved.
// Use of this source code is governed under MIT License
// that can be found in the LICENSE file.

// Package scanner allows to extract information from the DTD and create corresponding DTD structs
package scanner

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/blefort/DTDParser/DTD"
)

type parsedBlock struct {
	id         string
	fullString string
	blockType  int
	name       string
	value      string
	entity     bool
	public     bool
	system     bool
	required   bool
	implied    bool
	fixed      bool
	uri        string
}

func (p *parsedBlock) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("Name", p.name)
	enc.AddString("blockType", DTD.Translate(p.blockType))
	if p.id != "" {
		enc.AddString("id", p.id)
	}
	if p.public {
		enc.AddBool("public", p.public)
	}
	if p.system {
		enc.AddBool("system", p.system)
	}
	if p.required {
		enc.AddBool("required", p.required)
	}
	if p.implied {
		enc.AddBool("implied", p.implied)
	}
	if p.fixed {
		enc.AddBool("fixed", p.fixed)
	}
	if p.value != "" {
		enc.AddString("value", p.value)
	}
	if p.uri != "" {
		enc.AddString("uri", p.uri)
	}
	return nil
}

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
func (sc *DTDScanner) Scan() (DTD.IDTDBlock, error) {

	fmt.Sprintf("Scan starts")
	t1 := time.Now().UnixMilli()

	//sc.Log.Debug("Seeking for next block")
	fmt.Sprintf("seek")
	s := sc.seekUntilNextBlock()

	//sc.Log.Warnf("block", zap.Object("Sentence", s))
	fmt.Sprintf("seek %v", s)

	sc.findDTDBlockType(s)

	if s.DTDType == DTD.UNIDENTIFIED {
		return nil, errors.New("Unidentified block")
	}

	if s.DTDType == DTD.COMMENT {
		comment := sc.ParseComment(s)
		t2 := time.Now().UnixMilli()
		diff := t2 - t1
		sc.Log.Infof("Commment found at line '%d' in '%d'", sc.CurrentLine, diff)
		return comment, nil
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
		return entity, nil
	}

	// if s.DTDType == DTD.ATTLIST {
	// 	attlist := sc.ParseAttlist(p)
	// 	t2 := time.Now().UnixMilli()
	// 	diff := t2 - t1
	// 	sc.Log.Infof("ATTLIST '%s' (line %d) in '%d'", attlist.GetName(), sc.CurrentLine, diff)
	// 	sc.logOutputAttributes(&attlist.Attributes)
	// 	return attlist, nil
	// }

	// if s.DTDType == DTD.ELEMENT {
	// 	element := sc.ParseElement(p)
	// 	t2 := time.Now().UnixMilli()
	// 	diff := t2 - t1
	// 	sc.Log.Infof("ELEMENT '%s' (line %d) in '%d'", element.GetName(), sc.CurrentLine, diff)
	// 	return element, nil
	// }

	// if s.DTDType == DTD.NOTATION {
	// 	notation := sc.ParseNotation(p)
	// 	t2 := time.Now().UnixMilli()
	// 	diff := t2 - t1
	// 	sc.Log.Infof("NOTATION '%s' (line %d) in '%d'", notation.GetName(), sc.CurrentLine, diff)
	// 	return notation, nil
	// }

	return nil, errors.New("Unidentified block")

}

// ParseComment Parse a string and return pointer to DTD.Comment
func (sc *DTDScanner) ParseComment(s *sentence) *DTD.Comment {
	var c DTD.Comment

	// for _, w := range s.words {
	// 	stopped := "not stopped"
	// 	if w.stopped() {
	// 		stopped = "stopped"
	// 	}
	// 	sc.Log.Warnf("- [" + w.read() + "] " + stopped)
	// }
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
func (sc *DTDScanner) ParseElement(p *parsedBlock) *DTD.Element {
	var e DTD.Element
	e.Name = p.name
	e.Value = p.value
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
func (sc *DTDScanner) ParseNotation(p *parsedBlock) *DTD.Notation {
	var n DTD.Notation
	n.Name = p.name
	n.Public = p.public
	n.System = p.system
	n.Url = p.uri
	n.ID = p.id
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

	words := s.getWords()
	idx := 0

	for i, w := range words {

		//stopped := "not stopped"
		//if w.stopped() {
		//	stopped = "stopped"
		//}

		//sc.Log.Warnf("- [" + w.read() + "] " + stopped)

		if w.read() == "%" {
			e.Parameter = true
		}

		if w.read() == "PUBLIC" {
			e.Public = true
			e.IsExternal = true
			idx = i
		}

		if w.read() == "SYSTEM" {
			e.System = true
			e.IsExternal = true
			idx = i
		}
	}

	if e.Parameter {
		e.Name = s.words[2].read()
	} else {
		e.Name = s.words[1].read()
	}

	if e.System {
		e.Url = words[idx+1].read()
	} else if e.Public {
		e.Url = words[idx+2].read()
		e.Value = words[idx+1].read()
	} else if e.Parameter {
		e.Value = s.words[3].read()
	} else {
		e.Value = words[2].read()
	}

	return &e
}

// ParseAttlist Parse a string and return pointer to a DTD.Attlist
//
// [52]   	AttlistDecl	   ::=   	'<!ATTLIST' S Name AttDef* S? '>'
// [53]   	AttDef	   ::=   	S Name S AttType S DefaultDecl
//
func (sc *DTDScanner) ParseAttlist(p *parsedBlock) *DTD.Attlist {
	var attlist DTD.Attlist
	attlist.Name = p.name
	sc.parseAttributes(p.value, &attlist.Attributes)
	return &attlist
}

func (sc *DTDScanner) parseAttributes(s string, attributes *[]DTD.Attribute) {

	i := -1
	s2 := sc.normalizeSpace(s)
	parts := sc.SeekWords(s2)
	l := len(parts)

	sc.Log.Debugf("parse Attr: parts %v, s was '%s'", parts, s)

	// for idx, part := range parts {
	// 	sc.Log.Debugf("part %d: %s", idx, part)
	// }

	if l == 0 {
		panic("Unable to scan Attlist")
	}

	nextWord := func(i *int, l int) bool {
		if *i+1 < l {
			*i++
			sc.Log.Debugf("Next word, i: %d, %s ", *i, parts[*i])
			return true
		}
		return false
	}

	for nextWord(&i, l) {

		var attr DTD.Attribute

		// empty string
		if sc.normalizeSpace(parts[i]) == "" {
			continue
		}

		// instruction varies depending the type, we don't know in advance

		// reference to an entity
		if strings.HasPrefix(parts[i], "%") {
			attr.Value = parts[i]
			sc.Log.Debugf("- Ref. to an entity found: %s", attr.Render())
			attr.IsEntity = true
			*attributes = append(*attributes, attr)
			continue
		}

		// first word is always the attribute name
		attr.Name = sc.normalizeSpace(parts[i])
		sc.Log.Debugf("Processing attribute: '%s'", attr.Name)

		if !nextWord(&i, l) {
			sc.Log.Fatalf("Not enough arguments to loop through attributes i:%d", i)

		}

		// CASE 2
		// The others parts are processed by group of 3 to 4 depending the type

		// Type is always in the second position
		attr.Type = DTD.SeekAttributeType(parts[i])

		sc.Log.Debugf("attribute type is %d", attr.Type)

		if attr.Type == DTD.CDATA { // 20
			nextWord(&i, l)
			//		sc.checkNextTwoArguments(parts, &i, &attr)
		}

		if attr.Type == DTD.TOKEN_ID ||
			attr.Type == DTD.TOKEN_IDREF ||
			attr.Type == DTD.TOKEN_IDREFS ||
			attr.Type == DTD.TOKEN_ENTITY ||
			attr.Type == DTD.TOKEN_ENTITIES ||
			attr.Type == DTD.TOKEN_NMTOKEN ||
			attr.Type == DTD.TOKEN_NMTOKENS {
			nextWord(&i, l)
			//		sc.checkNextTwoArguments(parts, &i, &attr)
		}
		if attr.Type == DTD.ENUM_NOTATION {
			nextWord(&i, l)
			attr.Value = parts[i]
			nextWord(&i, l)
			//		sc.checkDefaultDeclaration(&attr, parts[i])
		}
		if attr.Type == DTD.ENUM_ENUM {
			attr.Value = parts[i]
			nextWord(&i, l)
			//		sc.checkDefaultDeclaration(&attr, parts[i])
		}

		*attributes = append(*attributes, attr)
		sc.Log.Debugf("*Attr rendered: %s", attr.Render())

		if attr.Type == 0 {
			sc.Log.Fatalf("Could not identify attribute type at line %d, name: '%s', value: '%s'", sc.LineCount, attr.Name, parts[i])
		}

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

// SeekWords Walk a string and identify every words
func (sc *DTDScanner) SeekWords(s string) []string {

	r2 := `"([^"]+)"|\((.*)\)[\+|\?|\*]?|([^\s]+)`

	regex := regexp.MustCompile(r2)
	parts := regex.FindAllString(s, -1)

	sc.Log.Debugf("seekWords FindAllString found %#v", parts)
	return parts
}

// // extractDeclaration call the rigt sekk method depending the declaration
// func (sc *DTDScanner) extractDeclaration(s string, declaration int) (*parsedBlock, error) {

// 	if declaration == BLOCK_XML {
// 		return sc.SeekXMLParts(s)
// 	}

// 	//sc.Log.Debugf("Block line: '%s'", s)
// 	if declaration == BLOCK_COMMENT {
// 		return sc.SeekCommentParts(s)
// 	}

// 	if sc.normalizeSpace(s) == "" {
// 		return nil, errors.New("End of DTD")
// 	}

// 	return sc.SeekBlockParts(s)
// }

//SeekXMLParts extract Comment information from string using a Regex
// func (sc *DTDScanner) SeekXMLParts(s string) (*parsedBlock, error) {
// 	var p parsedBlock
// 	p.blockType = "XMLDECL"
// 	p.value = s
// 	return &p, nil
// }

// SeekCommentParts extract Comment information from string using a Regex
func (sc *DTDScanner) SeekCommentParts(s string) (*parsedBlock, error) {

	var p parsedBlock
	var r string

	r = `<!--([\s\S\n]*?)-->`
	regex := regexp.MustCompile(r)
	parts := regex.FindAllStringSubmatch(s, -1)

	p.value = strings.TrimSpace(parts[0][1])
	p.blockType = DTD.COMMENT

	sc.Log.Debugf("SeekCommentPart, parsed: , name: [%s], type: [%s], entity: [%s], value: [%s], s was [%s]", p.name, p.blockType, p.entity, p.value, s)

	return &p, nil
}

// SeekBlockParts extract DTD information from string using a Regex
// func (sc *DTDScanner) SeekBlockParts(s string) (*parsedBlock, error) {

// 	var p parsedBlock
// 	var r string

// 	sc.Log.Debugf("SeekBlockParts received %s", s)

// 	r = `<\!(ENTITY|ELEMENT|ATTLIST|COMMENT|NOTATION)\s*(\%)?\s*(\S+)?\s*([^>]+)?>\s*(%[^>\s]+)?`

// 	regex := regexp.MustCompile(r)
// 	parts := regex.FindAllStringSubmatch(s, -1)

// 	p.fullString = parts[0][0]
// 	//p.blockType = parts[0][1]
// 	p.entity = parts[0][2]
// 	p.name = parts[0][3]
// 	p.value = parts[0][4]

// 	sc.Log.Debugf("SeekBlockParts, parsed: \n-name: %s\n-type:%s\n-entity:%s\n-value:%s, s was '%s'", p.name, p.blockType, p.entity, p.value, s)

// 	return &p, nil
// }

// isQuoted returns true if a character is quote or a double quote
func isQuoted(s string) bool {
	return s == "\"" || s == "'"
}

// checkEnumDefaultValue
func (sc *DTDScanner) checkEnumDefaultValue(attr *DTD.Attribute, parts []string) {

	//	attr.Default = sc.checkDefaultValue(parts[0])
	sc.Log.Debugf("Enum Default Value: '%s'", attr.Default)
}

// trimQuotes Trim surrounding quotes
func trimQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}

// IsStartChar Determine if a character is the beginning of a DTD block
func (sc *DTDScanner) IsStartChar() bool {
	return sc.Data.Text() == "<"
}

// SeekExportedEntity Seek an exported DTD entity
// For example:
//  <!ENTITY % concept-dec
//           PUBLIC "-//OASIS//ENTITIES DITA 1.2 Concept//EN"
//           "concept.ent">%concept-dec;
//
// %concept-dec means that the entity is exported
func (sc *DTDScanner) SeekExportedEntity() string {
	var s string
	for sc.next() {
		if sc.Data.Text() == ";" {
			return s
		}
		s += sc.Data.Text()
	}
	return s
}

// isWhitespace Determine if a string is a whitespace
func (sc *DTDScanner) isWhitespace() bool {
	return sc.Data.Text() == " " || sc.Data.Text() == "\t" || sc.isEndOfLine()
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

	//	var s string
	var p parsedBlock
	//	var contentPosition int
	sentence := newsentence("<", ">", sc.Log)

	sc.CurrentLine = sc.LineCount
	sc.Log.Warnf("start scanning sentence with char '" + sc.Data.Text() + "'")
	sentence.scan(sc.Data.Text())

	for sc.next() {

		if sentence.scan(sc.Data.Text()) {
			sentence.read()
			return sentence
		}
	}

	p.blockType = DTD.UNIDENTIFIED
	return sentence

}

func (sc *DTDScanner) findDTDBlockType(s *sentence) {
	w := s.words[0].read()
	sc.Log.Warnf("word is" + w)
	if len(w) > 2 && w[0:3] == "!--" {
		s.DTDType = DTD.COMMENT
	} else if w == "!ATTLIST" {
		s.DTDType = DTD.ATTLIST
	} else if w == "!ELEMENT" {
		s.DTDType = DTD.ELEMENT
	} else if w == "!NOTATION" {
		s.DTDType = DTD.NOTATION
	} else if w == "!ENTITY" {
		s.DTDType = DTD.ENTITY
	}
	sc.Log.Warnf(fmt.Sprintf("%d", s.DTDType))
}

// normalizeSpace Convert Line breaks, multiple space into a single space
func (sc *DTDScanner) normalizeSpace(s string) string {
	regexLineBreak := regexp.MustCompile(`(?s)(\r?\n)|\t`)
	s1 := regexLineBreak.ReplaceAllString(s, " ")
	space := regexp.MustCompile(`\s+|\t`)
	nm := strings.Trim(space.ReplaceAllString(s1, " "), " ")
	return nm
}
